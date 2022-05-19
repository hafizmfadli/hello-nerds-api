package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Province struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type City struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	ProvID int    `json:"prov_id,omitempty"`
}

type District struct {
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	CityID int    `json:"city_id,omitempty"`
}

type SubDistrict struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	DistrictID int    `json:"district_id,omitempty"`
}

type PostalCode struct {
	ID            int    `json:"id,omitempty"`
	ProvID        int    `json:"prov_id,omitempty"`
	CityID        int    `json:"city_id,omitempty"`
	DistrictID    int    `json:"district_id,omitempty"`
	SubDistrictID int    `json:"sub_district_id,omitempty"`
	Code          int `json:"postal_code,omitempty"`
}

type IndonesiaModel struct {
	DB *sql.DB
}

// GetProvince return province data that match provided id
func (m IndonesiaModel) GetProvince(id int) (*Province, error) {
	query := `
		SELECT prov_id, prov_name FROM ec_provinces WHERE prov_id = ?`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var province Province

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&province.ID, &province.Name)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &province, nil
}

// GetProvinces return all provinces
func (m IndonesiaModel) GetProvinces() ([]*Province, error) {
	query := `
		SELECT prov_id, prov_name FROM ec_provinces`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	provinces := []*Province{}

	for rows.Next() {
		var province Province

		err = rows.Scan(&province.ID, &province.Name)
		if err != nil {
			return nil, err
		}
		provinces = append(provinces, &province)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return provinces, nil
}

// GetCitiesByProv return cities in particular province
func (m IndonesiaModel) GetCitiesByProv(provId int) ([]*City, error){
	query := `
		SELECT city_id, city_name, prov_id FROM ec_cities WHERE prov_id = ?`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var cities []*City

	rows, err := m.DB.QueryContext(ctx, query, provId)
	if err != nil {
		return nil, err
	}

	// Make sure resultset is closed before GetCitiesByProv() return
	defer rows.Close()

	cities = []*City{}

	// Iterate through the rows in resultset
	for rows.Next() {
		var city City

		err = rows.Scan(&city.ID, &city.Name, &city.ProvID)
		if err != nil {
			return nil, err
		}

		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}

// GetDistrictsByCity return districts in particular city
func (m IndonesiaModel) GetDistrictsByCity(cityId int) ([]*District, error) {
	query := `
		SELECT dis_id, dis_name, city_id FROM ec_districts WHERE city_id = ?`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, cityId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var districts []*District

	for rows.Next() {
		var district District

		err = rows.Scan(&district.ID, &district.Name, &district.CityID)
		if err != nil {
			return nil, err
		}

		districts = append(districts, &district)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return districts, nil
}

// GetSubDistrictsByDistrict return subdistricts in particular district
func (m IndonesiaModel) GetSubDistrictsByDistrict(districtId int) ([]*SubDistrict, error) {
	query := `
		SELECT subdis_id, subdis_name, dis_id FROM ec_subdistricts WHERE dis_id = ?`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, districtId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subDistricts := []*SubDistrict{}

	for rows.Next() {
		var subDistrict SubDistrict

		err = rows.Scan(&subDistrict.ID, &subDistrict.Name, &subDistrict.DistrictID)
		if err != nil {
			return nil, err
		}

		subDistricts = append(subDistricts, &subDistrict)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subDistricts, nil
}

// GetPostalCode return a postal code 
func (m IndonesiaModel) GetPostalCode (provId, cityId, districtId, subDistrictId int) (*PostalCode, error) {
	query := `
		SELECT postal_id, subdis_id, dis_id, city_id, prov_id, postal_code FROM ec_postalcode WHERE subdis_id = ?`
	
	args := []interface{}{subDistrictId}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var postalCode PostalCode

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&postalCode.ID, &postalCode.SubDistrictID,
	&postalCode.DistrictID, &postalCode.CityID, &postalCode.ProvID, &postalCode.Code)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &postalCode, nil
}