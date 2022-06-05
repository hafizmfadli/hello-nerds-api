package data

import (
	"math"

	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
)

type Filters struct {
	Searchword string
	Author string
	Extension string
	Availability int
	Page int
	PageSize int
	ISBN string
}


func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that the page and page_size parameters contain sensible values.
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 417, "page", "must be maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

}

func ValidateAdvanceFilters(v *validator.Validator, f Filters) {
	v.Check(f.Extension == "all" || f.Extension == "pdf" || f.Extension == "epub" || f.Extension == "djvu", "extension", "extension must be pdf, epub, djvu")
	
	// filter availability status 
	// 0 : (no filter)
	// 1 : in stock
	// 2 : currently unavailable
	v.Check(f.Availability >= 0, "availability", "availability status must be greater than zero")
	v.Check(f.Availability <= 2, "availability", "availability status must be less than two")
}

// Define a new Metadata struct for holding the pagination metadata.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// The calculateMetadata() function calculates the appropriate pagination metadata
// values given the total number of records, current page, and page size values. Note
// that the last page value is calculated using the math.Ceil() function, which rounds
// up a float to the nearest integer. So, for example, if there were 12 records in total
// and a page size of 5, the last page value would be math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Note that we return an empty Metadata struct if there are no records.
		return Metadata{}
	}

	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}