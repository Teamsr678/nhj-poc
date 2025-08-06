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

	r.Run(":8080")
}
