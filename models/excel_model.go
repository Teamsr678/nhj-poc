package models

type Account struct {
	AccountID         *string
	CustomerID        *string
	ProductType       *string
	OutstandingAmount *int32
	OverdueAmount     *int32
	DaysPastDue       *int32
	SelfCured         *string
	TopUpScore        *string
	LossOnSale        *int32
	LossOnClaim       *string
}

type Customer struct {
	CustomerID         *string
	CustomerName       *string
	OccupationID       *int32
	RegisterAddress    *string
	RegisterTambol     *string
	RegisterAmphur     *string
	RegisterProvince   *string
	RegisterPostalCode *string
	CurrentAddress     *string
	CurrentTambol      *string
	CurrentAmphur      *string
	CurrentProvince    *string
	CurrentPostalCode  *string
}
type Occupation struct {
	OccupationID   int32
	OccupationName string
}
