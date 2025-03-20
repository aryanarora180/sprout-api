package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	router.HandlerFunc(http.MethodPost, "/v1/expenses", app.requireActivatedUser(app.createExpenseHandler))
	router.HandlerFunc(http.MethodGet, "/v1/expenses", app.requireActivatedUser(app.listExpenseHandler))
	router.HandlerFunc(http.MethodGet, "/v1/expenses/:id", app.requireActivatedUser(app.showExpenseHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/expenses/:id", app.requireActivatedUser(app.updateExpenseHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/expenses/:id", app.requireActivatedUser(app.deleteExpenseHandler))

	router.HandlerFunc(http.MethodGet, "/v1/ai/classify-receipt", app.requireActivatedUser(app.classifyReceiptHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
