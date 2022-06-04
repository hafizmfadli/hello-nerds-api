package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/hafizmfadli/hello-nerds-api/internal/data"
	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
)

func (app *application) registerUserHandler (w http.ResponseWriter, r *http.Request) {
	// create an anoymous struct to hold the expected data from the request body
	var input struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	// parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// copy the data from the request body into a new User struct
	user := &data.User{
		FirstName: input.FirstName,
		LastName: input.LastName,
		Email: input.Email,
		Activated: false,
	}

	// generate and store the hashed and plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// validate password and confirm password
	if data.ValidateConfirmPassword(v, input.Password, input.ConfirmPassword); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// validate the user struct
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert user data into the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
			// if we get a ErrDuplicateEmail error, use the v.AddError() method to manually
			// add a message to the validator instance
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Add the "books:read" permission for the new user.
	// (This is just for testing purpose, In the future we wanna give user permission 
	// related to cart).
	err = app.models.Permissions.AddForUser(user.ID, "books:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// After the user record has been created in the database, generate a new activation
	// token for the user.
	token, err := app.models.Tokens.New(user.ID, 3 * 24 * time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Use the background helper to execute an anonymous function that sends the welcome
	// email.
	app.background(func() {
		// As there are now multiple pieces of data that we want to pass to our email
		// templates, we create a map to act as a 'holding structure' for the data. This
		// contains the plaintext version of the activation token for the user, along
		// with their ID
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID": user.ID,
		} 

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler (w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by the client
	v := validator.New()

	if data.ValidateTokenPlainText(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve the details of the user associated with the token using the
	// GetForToken() method. If no mathcing record
	// is found, then we let the client know that the token they provided is not valid.
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user's activation status
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conlflicts
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConlictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If everything went successfully, then we delete all activation tokens for the
	// user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated user details to the client in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) checkoutHandler (w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token                     string                      `json:"token"`
		Carts                     []*data.Cart                `json:"carts"`
		OrderShippingAddress      data.ShippingAddress        `json:"shipping_address"`
		AddressVariety            data.ShippingAddressVariety `json:"address_variety"`
		CheckoutType              data.CheckoutVariety        `json:"checkout_type"`
		ExistingShippingAddressId int                         `json:"existing_shipping_address_id"`
	}

	var err error

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateCheckoutVariety(v, input.CheckoutType)
	data.ValidateShippingVariety(v, input.AddressVariety)
	data.ValidateCheckoutAndAddressVarietyPair(v, input.CheckoutType, input.AddressVariety)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var userID interface{}
	if input.CheckoutType == data.MemberCheckout {
		// Validate token only when is member checkout
		if data.ValidateTokenPlainText(v, input.Token); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, input.Token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("token", "invalid (has been expiry or tempered)")
				app.failedValidationResponse(w, r, v.Errors)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		userID = user.ID
	}

	err = app.models.Users.CheckoutV2(&input.OrderShippingAddress, input.AddressVariety, input.CheckoutType, int64(input.ExistingShippingAddressId), 
	input.Carts, userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNotEnoughStock):
			v.AddError("available_stock", "not enough")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("book", "not found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "order created"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

