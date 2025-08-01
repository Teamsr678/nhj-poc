package main

import (
	"fmt"
	"nhj-poc/controllers"
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

	r.POST("/upload-excel", controllers.UploadExcel)

	r.Run(":8080")
}
