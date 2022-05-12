package main

import (
	"errors"
	"net/http"

	"github.com/hafizmfadli/hello-nerds-api/internal/data"
	"github.com/hafizmfadli/hello-nerds-api/internal/validator"
)

func (app *application) listBooksHandler (w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string
	var input struct {
		data.Filters
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the query params
	input.Searchword = app.readString(qs, "searchword", "")
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 24, v)

	// execute validation check on the Filters struct
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var books []*data.Book
	var metadata data.Metadata
	var err error

	// we need to separate basic search and advance filter bcz
	// they have different elasticsearch query logic
	if app.isQueryParamExists(qs, "author") || app.isQueryParamExists(qs, "extension") || app.isQueryParamExists(qs, "availability") {
		// advance filter
		input.Author = app.readString(qs, "author", "")
		input.Extension = app.readString(qs, "extension", "all")
		input.Availability = app.readInt(qs, "availability", 0, v)

		if data.ValidateAdvanceFilters(v, input.Filters); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		books, metadata, err = app.models.Books.AdvanceFilterBooks(input.Filters)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

	} else {
		// basic search
		books, metadata, err = app.models.Books.GetAll(input.Filters)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}	
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"books": books, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBookSuggestionsHandler (w http.ResponseWriter, r *http.Request) {
	// define struct to store query params from request
	var input struct {
		Typesearch string
		data.Filters
	}

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	input.Typesearch = app.readString(qs, "typesearch", "")

	books, err := app.models.Books.GetBookSuggestions(input.Typesearch, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"suggestions": books}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBookHandler (w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.models.Books.GetBook(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}