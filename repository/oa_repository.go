package repository

import (
	"nhj-poc/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllOA(db *gorm.DB) ([]entity.OA, error) {
	var results []entity.OA
	if err := db.Model(&entity.OA{}).
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "oa_id"}, Desc: false},
		}}).
		Find(&results).Error; err != nil {
		return results, err
	}
	return results, nil
}
