package entity

import "database/sql"

type Assignments struct {
	AssignmentsID int             `gorm:"primaryKey;autoIncrement;not null" json:"assignments_id"`
	AccountID     *sql.NullString `gorm:"column:account_id" json:"account_id"`
	OaID          *sql.NullString `gorm:"column:oa_id" json:"oa_id"`
	AssignBy      *sql.NullString `gorm:"column:assign_by" json:"assign_by"`
}

func (Assignments) TableName() string {
	return "assignments"
}
