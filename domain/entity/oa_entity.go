package entity

import (
	"database/sql"
	"time"
)

type OA struct {
	OAId                   string           `gorm:"column:oa_id" json:"oa_id"`
	OAName                 *sql.NullString  `gorm:"column:oa_name" json:"oa_name"`
	Capacity               *sql.NullInt16   `gorm:"column:capacity" json:"capacity"`
	Ranking                *sql.NullInt16   `gorm:"column:ranking" json:"ranking"`
	C2CPercentage          *sql.NullFloat64 `gorm:"column:c2c_percentage" json:"c2c_percentage"`
	CRLPercentage          *sql.NullFloat64 `gorm:"column:crl_percentage" json:"crl_percentage"`
	PostalList             *sql.NullString  `gorm:"column:postal_list" json:"postal_list"`
	LocationLatitude       *sql.NullFloat64 `gorm:"column:location_latitude" json:"location_latitude"`
	LocationLongitude      *sql.NullFloat64 `gorm:"column:location_longitude" json:"location_longitude"`
	LocationUpdateDateTime *time.Time       `gorm:"column:location_update_datetime" json:"location_update_datetime"`
}

func (OA) TableName() string {
	return "oa"
}
