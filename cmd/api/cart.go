package main

import (
	"errors"
	"net/http"

	"github.com/hafizmfadli/hello-nerds-api/internal/data"
	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
)

func (app *application) insertCartHandler(w http.ResponseWriter, r *http.Request) {
	// create an anoymous struct to hold the expected data from the request body
	var input struct {
		Quantity        int64 `json:"quantity"`
		UserID          int64 `json:"user_id"`
		UpdatedEditedID int64 `json:"updated_edited_id"`
	}

	// parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy the data from the request body into a new Cart struct
	cart := &data.Cart{
		Quantity:        input.Quantity,
		UserID:          input.UserID,
		UpdatedEditedID: input.UpdatedEditedID,
	}

	// insert cart data into the database
	err = app.models.Carts.Insert(cart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed
	err = app.writeJSON(w, http.StatusAccepted, envelope{"cart": cart}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCartHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	cartDetails, err := app.models.Carts.GetByUserID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"carts": cartDetails}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateQuantityCartHandler(w http.ResponseWriter, r *http.Request) {

	// create an anoymous struct to hold the expected data from the request body
	var input struct {
		UserID          int64 `json:"user_id"`
		UpdatedEditedID int64 `json:"updated_edited_id"`
		Quantity        int64 `json:"quantity"`
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the query params
	input.UserID = app.readInt64(qs, "user_id", 1, v)
	input.UpdatedEditedID = app.readInt64(qs, "updated_edited_id", 1, v)

	// parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy the data from the request body into a new Cart struct
	cart := &data.Cart{
		UserID:          input.UserID,
		UpdatedEditedID: input.UpdatedEditedID,
		Quantity:        input.Quantity,
	}

	// insert cart data into the database
	err = app.models.Carts.UpdateQuantity(cart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed
	err = app.writeJSON(w, http.StatusAccepted, envelope{"cart": cart}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCartHandler(w http.ResponseWriter, r *http.Request) {

	// create an anoymous struct to hold the expected data from the request body
	var input struct {
		UserID          int64 `json:"user_id"`
		UpdatedEditedID int64 `json:"updated_edited_id"`
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the query params
	input.UserID = app.readInt64(qs, "user_id", 1, v)
	input.UpdatedEditedID = app.readInt64(qs, "updated_edited_id", 1, v)

	// copy the data from the request body into a new Cart struct
	cart := &data.Cart{
		UserID:          input.UserID,
		UpdatedEditedID: input.UpdatedEditedID,
	}

	// insert cart data into the database
	err := app.models.Carts.Delete(cart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed
	err = app.writeJSON(w, http.StatusAccepted, envelope{"cart": cart}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
