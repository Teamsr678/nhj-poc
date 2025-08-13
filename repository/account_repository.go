package repository

import (
	"nhj-poc/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllAccount(db *gorm.DB) ([]entity.Account, error) {
	var results []entity.Account
	if err := db.Model(&entity.Account{}).
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "outstanding_amount"}, Desc: true},
			{Column: clause.Column{Name: "account_id"}, Desc: false},
		}}).
		Find(&results).Error; err != nil {
		return results, err
	}
	return results, nil
}
