package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Expenses ExpenseModel
	Users    UserModel
	Tokens   TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Expenses: ExpenseModel{DB: db},
		Users:    UserModel{DB: db},
		Tokens:   TokenModel{DB: db},
	}
}
