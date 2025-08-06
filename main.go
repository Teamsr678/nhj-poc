package main

import (
	"fmt"
	"nhj-poc/controller"
	"nhj-poc/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	r := gin.Default()

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found or could not be loaded")
	}

	database.Connect()

	r.POST("/insert-payment", controller.InsertPayment)
	r.POST("/upload-excel", controller.UploadExcel)
	r.POST("/insert-transaction", controller.InsertTransaction)

	r.GET("/get-map-link", controller.GetMapsLinkHandler)
	r.POST("/update-location", controller.UpdateLocationHandler)
	r.GET("/get-locations", controller.GetLocationsHandler)
	r.GET("/get-route", controller.GetRouteHandler)

	r.Run(":8080")
}
