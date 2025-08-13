package controller

import (
	"net/http"
	"nhj-poc/service"

	"github.com/gin-gonic/gin"
)

func UpdateAssignmentsByProductType(c *gin.Context) {
	if err := service.UpdateAssignmentsByProductType(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "worklist update successfully"})
}
