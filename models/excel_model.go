package models

type Excel struct {
	AccountID          string `json:"accountId"`
	CustomerID         string `json:"customerId"`
	CustomerName       string `json:"customerName"`
	Occupation         string `json:"occupation"`
	RegisterAddress    string `json:"registerAddress"`
	RegisterTambol     string `json:"registerTambol"`
	RegisterAmphur     string `json:"registerAmphur"`
	RegisterProvince   string `json:"registerProvince"`
	RegisterPostalCode *int   `json:"registerPostalCode"`
	CurrentAddress     string `json:"currentAddress"`
	CurrentTambol      string `json:"currentTambol"`
	CurrentAmphur      string `json:"currentAmphur"`
	CurrentProvince    string `json:"currentProvince"`
	CurrentPostalCode  *int   `json:"currentPostalCode"`
	ProductType        string `json:"productType"`
	OutstandingAmount  string `json:"outstandingAmount"`
	OverdueAmount      string `json:"overdueAmount"`
	DaysPastDue        *int   `json:"daysPastDue"`
	SelfCured          string `json:"selfCured"`
	TopUpScore         string `json:"topUpScore"`
	LossOnSale         *int   `json:"lossOnSale"`
	LossOnClaim        string `json:"lossOnClaim"`
}

type Account struct {
	AccountID         string `gorm:"column:account_id" json:"account_id"`
	CustomerID        string `gorm:"column:customer_id" json:"customer_id"`
	ProductType       string `gorm:"column:product_type" json:"product_type"`
	OutstandingAmount *int   `gorm:"column:outstanding_amount" json:"outstanding_amount"`
	OverdueAmount     *int   `gorm:"column:overdue_amount" json:"overdue_amount"`
	DaysPastDue       *int   `gorm:"column:days_past_due" json:"days_past_due"`
	SelfCured         string `gorm:"column:self_cured" json:"self_cured"`
	TopUpScore        string `gorm:"column:top_up_score" json:"top_up_score"`
	LossOnSale        *int   `gorm:"column:loss_on_sale" json:"loss_on_sale"`
	LossOnClaim       string `gorm:"column:loss_on_claim" json:"loss_on_claim"`
}

func (Account) TableName() string {
	return "account"
}

type Customer struct {
	CustomerID         string `gorm:"column:customer_id" json:"customer_id"`
	CustomerName       string `gorm:"column:customer_name" json:"customer_name"`
	OccupationID       *int   `gorm:"column:occupation_id" json:"occupation_id"`
	RegisterAddress    string `gorm:"column:register_address" json:"register_address"`
	RegisterTambol     string `gorm:"column:register_tambol" json:"register_tambol"`
	RegisterAmphur     string `gorm:"column:register_amphur" json:"register_amphur"`
	RegisterProvince   string `gorm:"column:register_province" json:"register_province"`
	RegisterPostalCode *int   `gorm:"column:register_postal_code" json:"register_postal_code"`
	CurrentAddress     string `gorm:"column:current_address" json:"current_address"`
	CurrentTambol      string `gorm:"column:current_tambol" json:"current_tambol"`
	CurrentAmphur      string `gorm:"column:current_amphur" json:"current_amphur"`
	CurrentProvince    string `gorm:"column:current_province" json:"current_province"`
	CurrentPostalCode  *int   `gorm:"column:current_postal_code" json:"current_postal_code"`
}

func (Customer) TableName() string {
	return "customer"
}

type Occupation struct {
	OccupationID   int    `gorm:"primaryKey;autoIncrement;not null" json:"occupation_id"`
	OccupationName string `gorm:"column:occupation_name" json:"occupation_name"`
}

func (Occupation) TableName() string {
	return "occupation"
}
