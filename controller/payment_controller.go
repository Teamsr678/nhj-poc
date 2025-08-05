package controller

import (
	"net/http"
	"nhj-poc/domain/api"
	"nhj-poc/domain/model"
	"nhj-poc/services"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

func InsertPayment(c *gin.Context) {
	var payment api.Payment

	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	var pModel model.Payment
	if err := copier.Copy(&pModel, &payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.InsertPayment(pModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment created successfully"})
}

func InsertTransaction(c *gin.Context) {
	err := services.InsertTransaction()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successful",
	})
}
