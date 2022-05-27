package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter instance
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Likewise, convert the methodNotAllowedResponse() helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandlerFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the strings "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books", app.listBooksHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/suggest", app.listBookSuggestionsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/detail/:id", app.showBookHandler)

	// router.HandlerFunc(http.MethodGet, "/v1/books/detail/:id", app.requirePermission("books:read", app.showBookHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/provinces", app.listProvincesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cities", app.listCitiesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/districts", app.listDistrictsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/subdistricts", app.listSubdistrictsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/postalcode", app.selectPostalCodeHandler)

	router.HandlerFunc(http.MethodPost, "/v1/carts/add-to-cart", app.insertCartHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/cart", app.showCartHandler)
	router.HandlerFunc(http.MethodPut, "/v1/carts/setQuantity", app.updateQuantityCartHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/carts/delete", app.deleteCartHandler)

	router.HandlerFunc(http.MethodPost, "/v1/checkout", app.checkoutHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/logout", app.requireAuthenticatedUser(app.removeAuthenticationTokenHandler))

	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
