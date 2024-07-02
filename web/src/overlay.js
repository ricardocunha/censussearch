import React from 'react';
import './Overlay.css'; // Import the CSS file for styling the overlay

const Overlay = ({ featureProperties, onClose }) => {
    return (
        <div className="overlay">
            <div className="overlay-content">
                <h2>Information</h2>
                <p>Census Tract ID: {featureProperties.census_tract_id}</p>
                <p>Population: {featureProperties.total_population}</p>
                <button onClick={onClose}>Close</button>
            </div>
        </div>
    );
};

export default Overlay;