package services

import (
	"fmt"
	"net/url"
	"nhj-poc/database"
	"nhj-poc/entity"
	"nhj-poc/models"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func loadEnvVar(key string) string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return os.Getenv(key)
}

var (
	googleMapsAPIKey = loadEnvVar("GOOGLE_MAPS_API_KEY")
)

func GenerateMapsLink(latitude, longitude float64) string {
	mapsURL := fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f", latitude, longitude)
	parsedURL, err := url.Parse(mapsURL)
	if err != nil {
		return ""
	}
	return parsedURL.String()
}

func GenerateMapsRouteLink(startLat, startLon, endLat, endLon float64) string {
	mapsURL := fmt.Sprintf("https://www.google.com/maps/embed/v1/directions?key=%s&origin=%f,%f&destination=%f,%f",
		googleMapsAPIKey, startLat, startLon, endLat, endLon)
	parsedURL, err := url.Parse(mapsURL)
	if err != nil {
		return ""
	}
	return parsedURL.String()
}

func UpdateLocationOA(oaID string, latitude *float64, longitude *float64, updateDateTime *time.Time) error {
	var oa entity.OA
	if err := database.DB.First(&oa, "oa_id = ?", oaID).Error; err != nil {
		return err
	}
	updates := make(map[string]interface{})
	if latitude != nil {
		updates["location_latitude"] = *latitude
	}
	if longitude != nil {
		updates["location_longitude"] = *longitude
	}
	if updateDateTime != nil {
		updates["location_update_datetime"] = *updateDateTime
	}
	if len(updates) == 0 {
		return nil
	}

	result := database.DB.Model(&entity.OA{}).Where("oa_id = ?", oaID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update location for OA %s: %w", oaID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no OA record found with oa_id %s to update", oaID)
	}
	return nil
}

func GetAllOAs() (map[string]models.LocationData, error) {
	var oas []entity.OA
	if err := database.DB.Model(&entity.OA{}).Find(&oas).Error; err != nil {
		return nil, err
	}
	locations := make(map[string]models.LocationData)
	for _, oa := range oas {
		var lat, lon float64
		if oa.LocationLatitude != nil {
			lat = oa.LocationLatitude.Float64
		}
		if oa.LocationLongitude != nil {
			lon = oa.LocationLongitude.Float64
		}
		var timestamp time.Time
		if oa.LocationUpdateDateTime != nil {
			timestamp = oa.LocationUpdateDateTime.Local()
		}
		locations[oa.OAId] = models.LocationData{
			Latitude:  lat,
			Longitude: lon,
			Timestamp: timestamp,
		}
	}

	return locations, nil
}

func GenerateMapHTML(oaID string, destLat float64, destLon float64) (*string, error) {
	locations, err := GetAllOAs()
	if err != nil {
		return nil, err
	}
	location, ok := locations[oaID]
	var startLat, startLon float64
	if !ok {
		startLat, startLon = destLat, destLon
	} else {
		startLat, startLon = location.Latitude, location.Longitude
	}

	htmlContent := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Employee Route</title>
		<style>
			body { margin: 0; padding: 0; font-family: sans-serif; }
			h1 { text-align: center; margin: 20px; }
			#map { 
				height: 60vh; 
				width: 90vw; 
				margin: 0 auto; 
				border: 1px solid #ccc; 
				border-radius: 8px; 
				display: flex;
				align-items: center;
				justify-content: center;
			}
			.error-message {
				text-align: center;
				padding: 20px;
				font-size: 1.2em;
				color: #555;
			}
			.controls {
				width: 90vw;
				margin: 20px auto;
				padding: 20px;
				border: 1px solid #ccc;
				border-radius: 8px;
				display: flex;
				flex-direction: column;
				gap: 10px;
			}
			.controls input, .controls button {
				padding: 10px;
				border-radius: 4px;
				border: 1px solid #ddd;
				font-size: 1em;
			}
			.controls button {
				background-color: #4CAF50;
				color: white;
				border: none;
				cursor: pointer;
			}
			.controls button:hover {
				background-color: #45a049;
			}
			#current-location-info {
				margin-top: 10px;
				padding: 10px;
				border: 1px dashed #ccc;
				border-radius: 4px;
				background-color: #f9f9f9;
			}
		</style>
		<script src="https://maps.googleapis.com/maps/api/js?key=%s"></script>
	</head>
	<body>
		<h1>Route for %s</h1>
		<div id="map">
			<div class="error-message" id="map-error"></div>
		</div>
		<div class="controls">
			<h2>Update Employee Location</h2>
			<form id="update-form">
				<input type="hidden" name="employee_id" value="%s">
				<label for="lat">Latitude:</label>
				<input type="number" id="lat" name="lat" step="any" required>
				<label for="lon">Longitude:</label>
				<input type="number" id="lon" name="lon" step="any" required>
				<button type="submit">Update Location</button>
			</form>
			<div id="current-location-info">
				<p><strong>Current Location:</strong> <span id="current-coords">N/A</span></p>
				<p><strong>Last Updated:</strong> <span id="last-updated">N/A</span></p>
			</div>
		</div>
		<script>
			const oaID = "%s";
			const destLat = %f;
			const destLon = %f;
			const googleMapsAPIKey = "%s";

			let map;
			let directionsService;
			let directionsRenderer;
			const mapErrorDiv = document.getElementById("map-error");
			const currentLocationDiv = document.getElementById("current-coords");
			const lastUpdatedDiv = document.getElementById("last-updated");
			const updateForm = document.getElementById("update-form");
			let lastKnownLocation = { lat: %f, lng: %f };

			function initMap() {
				if (googleMapsAPIKey === "") {
					mapErrorDiv.innerHTML = "<h2>API Key is missing!</h2><p>Please provide a valid Google Maps API key.</p>";
					console.error("API Key is missing or is the default placeholder. Please replace 'YOUR_GOOGLE_MAPS_API_KEY' with your actual key.");
					return;
				}

				mapErrorDiv.style.display = 'none';
				const startPoint = lastKnownLocation;
				const endPoint = { lat: destLat, lng: destLon };

				directionsService = new google.maps.DirectionsService();
				directionsRenderer = new google.maps.DirectionsRenderer();
				
				map = new google.maps.Map(document.getElementById("map"), {
					zoom: 12,
					center: startPoint,
				});

				directionsRenderer.setMap(map);

				// Initial route calculation
				calculateAndDisplayRoute(startPoint, endPoint);

				// Start real-time updates
				setInterval(updateMap, 5000);
			}

			function calculateAndDisplayRoute(origin, destination) {
				directionsService.route(
					{
						origin: origin,
						destination: destination,
						travelMode: google.maps.TravelMode.DRIVING,
					},
					(response, status) => {
						if (status === "OK") {
							directionsRenderer.setDirections(response);
						} else {
							mapErrorDiv.style.display = 'block';
							mapErrorDiv.innerHTML = "<h2>Oops! Something went wrong.</h2><p>Directions request failed due to " + status + ".</p>";
							console.error("Directions request failed due to " + status);
						}
					}
				);
			}
			
			function updateMap() {
				fetch('/get-locations')
					.then(response => {
						if (!response.ok) {
							throw new Error('Network response was not ok');
						}
						return response.json();
					})
					.then(data => {
						if (data[oaID]) {
							const startPoint = {
								lat: data[oaID].latitude,
								lng: data[oaID].longitude
							};
							const endPoint = {
								lat: destLat,
								lng: destLon
							};
							
							// Check if location has actually changed before updating the map
							if (startPoint.lat !== lastKnownLocation.lat || startPoint.lng !== lastKnownLocation.lng) {
								calculateAndDisplayRoute(startPoint, endPoint);
								currentLocationDiv.textContent = "Lat: ${startPoint.lat}, Lng: ${startPoint.lng}";
								lastUpdatedDiv.textContent = new Date(data[oaID].timestamp).toLocaleString();
								lastKnownLocation = startPoint;
							}
						} else {
							console.error('Employee location not found.');
						}
					})
					.catch(error => console.error('Error fetching location:', error));
			}

			updateForm.addEventListener("submit", function(event) {
				event.preventDefault();
				const formData = new FormData(updateForm);

				fetch('/update-location', {
					method: 'POST',
					body: formData
				})
				.then(response => response.json())
				.then(data => {
					console.log('Location update response:', data);
					// The setInterval will handle the map update, but we can call it manually for an immediate update.
					updateMap();
				})
				.catch(error => console.error('Error updating location:', error));
			});

			window.onload = initMap;
		</script>
	</body>
	</html>
	`, googleMapsAPIKey, oaID, oaID, oaID, destLat, destLon, googleMapsAPIKey, startLat, startLon)

	return &htmlContent, nil
}
