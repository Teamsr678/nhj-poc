package entity

import "database/sql"

type Account struct {
	AccountID         string          `gorm:"column:account_id" json:"account_id"`
	CustomerID        string          `gorm:"column:customer_id" json:"customer_id"`
	ProductType       *sql.NullString `gorm:"column:product_type" json:"product_type"`
	OutstandingAmount *sql.NullInt32  `gorm:"column:outstanding_amount" json:"outstanding_amount"`
	OverdueAmount     *sql.NullInt32  `gorm:"column:overdue_amount" json:"overdue_amount"`
	DaysPastDue       *sql.NullInt32  `gorm:"column:days_past_due" json:"days_past_due"`
	SelfCured         *sql.NullString `gorm:"column:self_cured" json:"self_cured"`
	TopUpScore        *sql.NullString `gorm:"column:top_up_score" json:"top_up_score"`
	LossOnSale        *sql.NullInt32  `gorm:"column:loss_on_sale" json:"loss_on_sale"`
	LossOnClaim       *sql.NullString `gorm:"column:loss_on_claim" json:"loss_on_claim"`
	EarlyOA           *sql.NullString `gorm:"column:early_oa" json:"early_oa"`
}

func (Account) TableName() string {
	return "account"
}

type Customer struct {
	CustomerID         string          `gorm:"column:customer_id" json:"customer_id"`
	CustomerName       *sql.NullString `gorm:"column:customer_name" json:"customer_name"`
	OccupationID       *sql.NullInt32  `gorm:"column:occupation_id" json:"occupation_id"`
	RegisterAddress    *sql.NullString `gorm:"column:register_address" json:"register_address"`
	RegisterTambol     *sql.NullString `gorm:"column:register_tambol" json:"register_tambol"`
	RegisterAmphur     *sql.NullString `gorm:"column:register_amphur" json:"register_amphur"`
	RegisterProvince   *sql.NullString `gorm:"column:register_province" json:"register_province"`
	RegisterPostalCode *sql.NullString `gorm:"column:register_postal_code" json:"register_postal_code"`
	CurrentAddress     *sql.NullString `gorm:"column:current_address" json:"current_address"`
	CurrentTambol      *sql.NullString `gorm:"column:current_tambol" json:"current_tambol"`
	CurrentAmphur      *sql.NullString `gorm:"column:current_amphur" json:"current_amphur"`
	CurrentProvince    *sql.NullString `gorm:"column:current_province" json:"current_province"`
	CurrentPostalCode  *sql.NullString `gorm:"column:current_postal_code" json:"current_postal_code"`
}

func (Customer) TableName() string {
	return "customer"
}

type Occupation struct {
	OccupationID   int32  `gorm:"primaryKey;autoIncrement;not null" json:"occupation_id"`
	OccupationName string `gorm:"column:occupation_name" json:"occupation_name"`
}

func (Occupation) TableName() string {
	return "occupation"
}
