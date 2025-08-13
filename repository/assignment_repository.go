package repository

import (
	"nhj-poc/domain/entity"

	"gorm.io/gorm"
)

func DeleteAssignments(db *gorm.DB, productType string) error {
	result := db.Where("assign_by = ?", productType).Delete(&entity.Assignments{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
