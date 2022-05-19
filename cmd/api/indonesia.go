package main

import (
	"net/http"

	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
)

func (app *application) listProvincesHandler (w http.ResponseWriter, r *http.Request) {
	provinces, err := app.models.Indonesia.GetProvinces()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	
	err = app.writeJSON(w, http.StatusOK, envelope{"provinces": provinces}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCitiesHandler (w http.ResponseWriter, r *http.Request) {
	// input struct to hold expected values from the request query string
	var input struct {
		ProvID int
	}

	v := validator.New()

	qs := r.URL.Query()

	// read query param "prov_id" and store the value at input.ProvID
	app.readPositiveInt(qs, "prov_id", &input.ProvID, v)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	cities, err := app.models.Indonesia.GetCitiesByProv(input.ProvID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"cities": cities}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listDistrictsHandler (w http.ResponseWriter, r *http.Request){
	var input struct {
		CityID int
	}

	v := validator.New()
	qs := r.URL.Query()

	app.readPositiveInt(qs, "city_id", &input.CityID, v)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	} 
	
	districts, err := app.models.Indonesia.GetDistrictsByCity(input.CityID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"districts": districts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSubdistrictsHandler (w http.ResponseWriter, r *http.Request) {
	var input struct {
		DistrictID int
	}

	v := validator.New()
	qs := r.URL.Query()

	app.readPositiveInt(qs, "district_id", &input.DistrictID, v)

	if !v.Valid(){
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	subdistricts, err := app.models.Indonesia.GetSubDistrictsByDistrict(input.DistrictID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"subdistricts": subdistricts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) selectPostalCodeHandler (w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProvID int
		CityID int
		DistrictID int
		Subdistrict int
	}

	v := validator.New()
	qs := r.URL.Query()

	app.readPositiveInt(qs, "prov_id", &input.ProvID, v)
	app.readPositiveInt(qs, "city_id", &input.CityID, v)
	app.readPositiveInt(qs, "district_id", &input.DistrictID, v)
	app.readPositiveInt(qs, "subdistrict_id", &input.Subdistrict, v)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	postalCode, err := app.models.Indonesia.GetPostalCode(input.ProvID, input.CityID, input.DistrictID,
	input.Subdistrict)
	
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"postal_code": postalCode}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}