package data

import "database/sql"

// Define a Level type to represent the severity level for a log entry
type ShippingAddressVariety int8

// Initialize constants which represent a specific severity level. We use the iota
// keyword as a shortcut to assign successive integer values to the constants
const (
	ToNewAddress ShippingAddressVariety = iota
	ToExistingAddress
)

type ShippingAddress struct {
	Email         string `json:"email,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	Addresses     string `json:"addresses,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	ProvinceID    int    `json:"province_id,omitempty"`
	CityID        int    `json:"city_id,omitempty"`
	DistrictID    int    `json:"district_id,omitempty"`
	SubdistrictID int    `json:"subdistrict_id,omitempty"`
	Phone         string `json:"phone,omitempty"`
	userID        int    `json:"user_id,omitempty"`
}

type ShippingAddressModel struct {
	DB *sql.DB
}
