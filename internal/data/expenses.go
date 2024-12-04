package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sprout.example.com/internal/validator"
	"time"
)

type Expense struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"-"`
	Date            DateOnly  `json:"date"`
	SpentAt         string    `json:"spent_at"`
	Notes           string    `json:"notes"`
	Category        string    `json:"category"`
	PaymentMethod   string    `json:"payment_method"`
	ISOCurrencyCode string    `json:"iso_currency_code"`
	Amount          float64   `json:"amount"`
	Version         int32     `json:"version"`
}

func ValidateExpense(v *validator.Validator, expense *Expense) {
	v.Check(expense.SpentAt != "", "spent_at", "must be provided")
	v.Check(len(expense.SpentAt) <= 50, "spent_at", "must not be more than 50 bytes long")

	v.Check(len(expense.Notes) <= 100, "notes", "must not be more than 100 bytes long")

	// TODO: Make sure category is a valid category
	v.Check(expense.Category != "", "category", "must be provided")
	v.Check(len(expense.Category) <= 50, "category", "must not be more than 50 bytes long")

	// TODO: Make sure payment method is a valid one
	v.Check(expense.PaymentMethod != "", "payment_method", "must be provided")
	v.Check(len(expense.PaymentMethod) <= 50, "payment_method", "must not be more than 50 bytes long")

	// TODO: Make sure currency code is a valid one
	v.Check(expense.ISOCurrencyCode != "", "iso_currency_code", "must be provided")
	v.Check(len(expense.ISOCurrencyCode) <= 3, "iso_currency_code", "must not be more than 3 bytes long")

	v.Check(expense.Amount != 0, "amount", "must be provided")
	v.Check(expense.Amount > 0, "amount", "must be greater than 0")
	v.Check(expense.Amount < 1_000_000, "amount", "must be less than 1 million")

}

type ExpenseModel struct {
	DB *sql.DB
}

func (m ExpenseModel) Insert(userID int64, expense *Expense) error {
	query := `
		INSERT INTO expenses (user_id, date, spent_at, notes, category, payment_method, iso_currency_code, amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, version`

	args := []interface{}{userID, expense.Date.String(), expense.SpentAt, expense.Notes, expense.Category, expense.PaymentMethod, expense.ISOCurrencyCode, expense.Amount}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&expense.ID, &expense.CreatedAt, &expense.Version)
}

func (m ExpenseModel) Get(userID int64, id int64) (*Expense, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, created_at, date, spent_at, notes, category, payment_method, iso_currency_code, amount, version
		FROM expenses
		WHERE user_id = $1 AND id = $2`

	var expense Expense

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID, id).Scan(
		&expense.ID,
		&expense.CreatedAt,
		&expense.Date,
		&expense.SpentAt,
		&expense.Notes,
		&expense.Category,
		&expense.PaymentMethod,
		&expense.ISOCurrencyCode,
		&expense.Amount,
		&expense.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &expense, nil
}

func (m ExpenseModel) Update(userID int64, expense *Expense) error {
	query := `UPDATE expenses
		SET date = $1, spent_at = $2, notes = $3, category = $4, payment_method = $5, iso_currency_code = $6, amount = $7, version = version + 1
		WHERE user_id=$8 AND id = $9 AND version = $10
		RETURNING version`

	args := []interface{}{expense.Date.String(), expense.SpentAt, expense.Notes, expense.Category, expense.PaymentMethod, expense.ISOCurrencyCode, expense.Amount, userID, expense.ID, expense.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&expense.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m ExpenseModel) Delete(userID int64, id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM expenses
		WHERE user_id = $1 AND id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID, id)
	if err != nil {
		return nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m ExpenseModel) GetAll(userID int64, myf MonthYearFilter, filters Filters) ([]*Expense, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, date, spent_at, notes, category, payment_method, iso_currency_code, amount, version
		FROM expenses
		WHERE user_id = $1 AND date >= $2 AND date < $3
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	startDate, endDate := myf.monthRange()
	args := []interface{}{userID, startDate, endDate, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	var expenses []*Expense

	for rows.Next() {
		var expense Expense

		err := rows.Scan(
			&totalRecords,
			&expense.ID,
			&expense.Date,
			&expense.SpentAt,
			&expense.Notes,
			&expense.Category,
			&expense.PaymentMethod,
			&expense.ISOCurrencyCode,
			&expense.Amount,
			&expense.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		expenses = append(expenses, &expense)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return expenses, metadata, nil
}
