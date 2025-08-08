package controller

import (
	"net/http"
	"nhj-poc/domain/api"
	"nhj-poc/domain/model"
	"nhj-poc/service"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

func InsertPayment(c *gin.Context) {
	var pAPI api.Payment

	if err := c.ShouldBindJSON(&pAPI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	var pModel model.Payment
	if err := copier.Copy(&pModel, &pAPI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.InsertPayment(pModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment created successfully"})
}

func InsertTransaction(c *gin.Context) {
	var tAPI api.Transaction

	if err := c.ShouldBindJSON(&tAPI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	var tModel model.Transaction
	if err := copier.Copy(&tModel, &tAPI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := service.InsertTransaction(tModel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction created successfully",
	})
}

func UpdatePaymentStatus(c *gin.Context) {
	var updatePaymentStatusAPI api.UpdatePaymentStatus
	if err := c.ShouldBindJSON(&updatePaymentStatusAPI); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	if err := service.UpdatePaymentStatusByID(updatePaymentStatusAPI.PaymentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment status update successfully"})
}
