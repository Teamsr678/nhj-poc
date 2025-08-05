package controller

import (
	"net/http"
	"nhj-poc/services"

	"github.com/gin-gonic/gin"
)

func UploadExcel(c *gin.Context) {
	err := services.ProcessExcelUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload successful",
	})
}
