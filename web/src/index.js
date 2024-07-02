import React, { useState } from "react";
import L from "leaflet";
import { Map, TileLayer, GeoJSON } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import axios from "axios";
import { render } from "react-dom";
import './App.css'
import Overlay from './overlay';

const styles = {
    fontFamily: "sans-serif",
    textAlign: "center",
    height: "100%"
};

const mapStyle = {
    height: "800px",
};

const baseURL = "http://localhost:8080/search";

export default function App() {
    const [post, setPost] = useState(null);
    const [selectedFeature, setSelectedFeature] = useState(null);
    const [postcode, setPostcode] = useState("85041");
    const handleSearch = () => {
        axios
            .get(`${baseURL}?postcode=${postcode}`, {
                headers: { Authorization: `Bearer ` },
            })
            .then((response) => {
                setPost(response.data.census_data.geojson);
            })
            .catch((error) => {
                console.error("Error fetching data:", error);
            });
    };

    const onDrillDown = (e) => {
        setSelectedFeature(e.target.feature.properties);
    };

    const onCloseOverlay = () => {
        setSelectedFeature(null);
    };
    return (
        <div style={styles}>
            <div>
                <label>Postcode: </label>
                <input
                    type="text"
                    value={postcode}
                    onChange={(e) => setPostcode(e.target.value)}
                    placeholder="Enter postcode"
                />
                <button onClick={handleSearch}>Search</button>
            </div>
            {post && (
                <Map
                    style={mapStyle}
                    bounds={L.geoJSON(post).getBounds()}
                    key={JSON.stringify(post)}
                    scrollWheelZoom={true}
                    zoom={13}
                >
                    <TileLayer
                        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                        attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
                    />
                    <GeoJSON data={post} color="blue" opacity="1" onEachFeature={onEachFeature} />

                </Map>
            )}
            {selectedFeature && <Overlay featureProperties={selectedFeature} onClose={onCloseOverlay} />}
        </div>
    );

    function onEachFeature(_, layer) {
        layer.on({
            click: onDrillDown,

        });
    }


}

render(<App />, document.getElementById("root"));