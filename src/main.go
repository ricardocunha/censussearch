package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type PostcodeResponse struct {
	Postcode           string     `json:"postcode"`
	CensusTract2010IDs []string   `json:"census_tract_2010_ids"`
	CensusData         CensusData `json:"census_data"`
}

type GeoJSON struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type CensusData struct {
	GeoJSON            GeoJSON `json:"geojson"`
	CensusTract        string  `json:"census_tract"`
	State              string  `json:"state"`
	County             string  `json:"county"`
	Population         int     `json:"total_population"`
	MedianHousingValue int     `json:"median_housing_value"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   Geometry               `json:"geometry"`
	ID         string                 `json:"id"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates [][][][]float64 `json:"coordinates"`
}

type TableCensusJSON struct {
	CensusTract2010ID string `gorm:"primaryKey;column:census_tract_2010_id;type:varchar(11);not null"`
	GeoJSON           string `gorm:"column:geojson;type:longtext"`
}

func (TableCensusJSON) TableName() string {
	return "census_json"
}

type TablePostcodeCensusTract struct {
	ID                int    `gorm:"primaryKey;autoIncrement;column:id;type:int(11);not null"`
	Postcode          string `gorm:"column:postcode;type:varchar(10);collate:utf8mb3_unicode_ci;not null"`
	CensusTract2010ID string `gorm:"column:census_tract_2010_id;type:varchar(11);not null"`
}

type TableCensusData struct {
	ID                 int    `gorm:"primaryKey;autoIncrement;column:id;type:int(11);not null"`
	CensusTract2010ID  string `gorm:"column:census_tract_2010_id;type:varchar(11);not null"`
	State              string `gorm:"column:state;type:varchar(50)"`
	CountyName         string `gorm:"column:county_name;type:varchar(100)"`
	TotalPopulation    int    `gorm:"column:total_population;type:integer(11)"`
	MedianHousingValue int    `gorm:"column:median_housing_value;type:integer(11)"`
	IsLowIncome        string `gorm:"column:is_low_income;type:varchar(11)"`
}

func (TableCensusData) TableName() string {
	return "census_data"
}

func main() {
	dsn := "root:root@tcp(census_db:3306)/census"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		searchGeo(db, w, r)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func searchGeo(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	postcode := r.URL.Query().Get("postcode")
	var censusTract2010IDs []string
	err := db.Table("postcode_census_tract").
		Select("census_tract_2010_id").
		Where("postcode = ?", postcode).
		Pluck("census_tract_2010_id", &censusTract2010IDs).Error
	if err != nil {
		log.Println("Error querying postcode:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var censusData CensusData
	if len(censusTract2010IDs) > 0 {

		var censusDatas []TableCensusData
		err := db.Where("census_tract_2010_id IN (?)", censusTract2010IDs).Find(&censusDatas).Error
		if err != nil {
			log.Println("Error querying census_data:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var censusJSONs []TableCensusJSON
		err = db.Where("census_tract_2010_id IN (?)", censusTract2010IDs).Find(&censusJSONs).Error
		if err != nil {
			log.Println("Error querying census_json:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		geoJSON := GeoJSON{
			Type:     "FeatureCollection",
			Features: []Feature{},
		}

		for _, censusJSON := range censusJSONs {
			if censusJSON.GeoJSON != "" {
				censusData = *getCensusData(censusJSON.CensusTract2010ID, censusDatas)
				var geometry Geometry
				err := json.Unmarshal([]byte(censusJSON.GeoJSON), &geometry)
				if err != nil {
					fmt.Println("Error parsing geometry:", err)
					return
				}
				feature := Feature{
					Type: "Feature",
					Properties: map[string]interface{}{
						"census_tract_id":      censusJSON.CensusTract2010ID,
						"total_population":     censusData.Population,
						"median_housing_value": censusData.MedianHousingValue,
					},
					Geometry: geometry,
				}
				geoJSON.Features = append(geoJSON.Features, feature)
			}
		}
		censusData.GeoJSON = geoJSON
	}

	response := PostcodeResponse{
		Postcode:           postcode,
		CensusTract2010IDs: censusTract2010IDs,
		CensusData:         censusData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getCensusData(censusTractId string, data []TableCensusData) *CensusData {
	for _, item := range data {
		if item.CensusTract2010ID == censusTractId {
			return &CensusData{
				CensusTract:        censusTractId,
				Population:         item.TotalPopulation,
				MedianHousingValue: item.MedianHousingValue,
				County:             item.CountyName,
			}
		}
	}
	return nil
}
