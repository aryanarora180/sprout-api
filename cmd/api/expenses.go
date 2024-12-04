package main

import (
	"errors"
	"fmt"
	"net/http"
	"sprout.example.com/internal/data"
	"sprout.example.com/internal/validator"
	"time"
)

func (app *application) showExpenseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	expense, err := app.models.Expenses.Get(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"expense": expense}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createExpenseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Date            data.DateOnly `json:"date"`
		SpentAt         string        `json:"spent_at"`
		Notes           string        `json:"notes"`
		Category        string        `json:"category"`
		PaymentMethod   string        `json:"payment_method"`
		ISOCurrencyCode string        `json:"iso_currency_code"`
		Amount          float64       `json:"amount"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	expense := &data.Expense{
		Date:            input.Date,
		SpentAt:         input.SpentAt,
		Notes:           input.Notes,
		Category:        input.Category,
		PaymentMethod:   input.PaymentMethod,
		ISOCurrencyCode: input.ISOCurrencyCode,
		Amount:          input.Amount,
	}

	v := validator.New()

	if data.ValidateExpense(v, expense); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Expenses.Insert(user.ID, expense)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", expense.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"expense": expense}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	expense, err := app.models.Expenses.Get(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Date            *data.DateOnly `json:"date"`
		SpentAt         *string        `json:"spent_at"`
		Notes           *string        `json:"notes"`
		Category        *string        `json:"category"`
		PaymentMethod   *string        `json:"payment_method"`
		ISOCurrencyCode *string        `json:"iso_currency_code"`
		Amount          *float64       `json:"amount"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Date != nil {
		expense.Date = *input.Date
	}

	if input.SpentAt != nil {
		expense.SpentAt = *input.SpentAt
	}

	if input.Notes != nil {
		expense.Notes = *input.Notes
	}

	if input.Category != nil {
		expense.Category = *input.Category
	}

	if input.PaymentMethod != nil {
		expense.PaymentMethod = *input.PaymentMethod
	}

	if input.ISOCurrencyCode != nil {
		expense.ISOCurrencyCode = *input.ISOCurrencyCode
	}

	if input.Amount != nil {
		expense.Amount = *input.Amount
	}

	v := validator.New()
	if data.ValidateExpense(v, expense); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Expenses.Update(user.ID, expense)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"expense": expense}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Expenses.Delete(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "expense deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listExpenseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.MonthYearFilter
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Month = app.readInt(qs, "month", int(time.Now().Month()), v)
	input.Year = app.readInt(qs, "year", time.Now().Year(), v)
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 50, v)
	input.Sort = app.readString(qs, "sort", "date")

	input.Filters.SortSafeList = []string{"date", "-date", "spent_at", "-spent_at", "category", "-category", "payment_method", "-payment_method", "amount", "-amount"}

	data.ValidateMonthYearFilter(v, input.MonthYearFilter)
	data.ValidateFilters(v, input.Filters)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.contextGetUser(r)

	expenses, metadata, err := app.models.Expenses.GetAll(user.ID, input.MonthYearFilter, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"expenses": expenses, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
