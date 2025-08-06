package controller

import (
	"fmt"
	"net/http"
	"nhj-poc/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetMapsLinkHandler(c *gin.Context) {
	latStr := c.Query("lat")
	if latStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'lat' query parameter"})
		return
	}
	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for 'lat' parameter"})
		return
	}
	lonStr := c.Query("lon")
	if lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'lon' query parameter"})
		return
	}
	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for 'lon' parameter"})
		return
	}

	mapsLink := services.GenerateMapsLink(latitude, longitude)
	c.JSON(http.StatusOK, gin.H{"maps_link": mapsLink})
}

func GetLocationsHandler(c *gin.Context) {
	locations, err := services.GetAllOAs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func UpdateLocationHandler(c *gin.Context) {
	oaID := c.PostForm("oa_id")
	if oaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'employee_id' in form data"})
		return
	}
	latStr := c.PostForm("lat")
	lonStr := c.PostForm("lon")
	latitude, errLat := strconv.ParseFloat(latStr, 64)
	longitude, errLon := strconv.ParseFloat(lonStr, 64)

	if errLat != nil || errLon != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude or longitude"})
		return
	}

	var now = time.Now()
	err := services.UpdateLocationOA(oaID, &latitude, &longitude, &now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Location updated for employee %s", oaID)})
}

func GetRouteHandler(c *gin.Context) {
	contentType := "text/html"
	employeeID := c.Query("employee_id")
	if employeeID == "" {
		c.Data(http.StatusBadRequest, contentType, []byte("<h1>Error: Missing 'employee_id' query parameter</h1>"))
		return
	}

	destLatStr := c.Query("dest_lat")
	destLonStr := c.Query("dest_lon")
	if destLatStr == "" || destLonStr == "" {
		c.Data(http.StatusBadRequest, contentType, []byte("<h1>Error: Missing 'dest_lat' or 'dest_lon' query parameter</h1>"))
		return
	}

	destLat, errLat := strconv.ParseFloat(destLatStr, 64)
	destLon, errLon := strconv.ParseFloat(destLonStr, 64)
	if errLat != nil || errLon != nil {
		c.Data(http.StatusBadRequest, contentType, []byte("<h1>Error: Invalid destination latitude or longitude</h1>"))
		return
	}
	htmlContent, err := services.GenerateMapHTML(employeeID, destLat, destLon)
	if err != nil {
		c.Data(http.StatusInternalServerError, contentType, []byte("<h1>Error: GenerateMap fail</h1>"))
	}
	c.Data(http.StatusOK, contentType, []byte(*htmlContent))
}
