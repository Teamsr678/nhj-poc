package main

import (
	"context"
	"fmt"
	"log"
	"nhj-poc/controller"
	"nhj-poc/database"
	"nhj-poc/routine"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	r := gin.Default()

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found or could not be loaded")
	}

	database.Connect()
	callRoutine()

	r.POST("/insert-payment", controller.InsertPayment)
	r.PUT("/update-payment-status", controller.UpdatePaymentStatus)
	r.POST("/upload-excel", controller.UploadExcel)
	r.POST("/insert-transaction", controller.InsertTransaction)

	r.GET("/get-map-link", controller.GetMapsLinkHandler)
	r.POST("/update-location", controller.UpdateLocationHandler)
	r.GET("/get-locations", controller.GetLocationsHandler)
	r.GET("/get-route", controller.GetRouteHandler)

	r.PUT("/update-assignments-by-product-type", controller.UpdateAssignmentsByProductType)

	r.Run(":8080")
}

func callRoutine() {
	_, err := routine.StartUpdatePaymentStatusJob(context.Background())
	if err != nil {
		log.Fatalf("failed to start batch routine: %v", err)
	}
}
