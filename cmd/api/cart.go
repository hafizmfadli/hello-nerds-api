package main

import (
	"net/http"

	"github.com/hafizmfadli/hello-nerds-api/internal/data"
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
