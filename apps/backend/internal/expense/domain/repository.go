package domain

import (
	"context"
	"errors"
)

var (
	ErrExpenseNotFound = errors.New("expense not found")
	ErrUserNotFound    = errors.New("user not found")
)

type ExpenseRepository interface {
	// persists an Expense and all of its associated Splits.
	Save(ctx context.Context, expense *Expense) error

	// reconstructs a full Expense object from storage.
	GetByID(ctx context.Context, id ExpenseID) (*Expense, error)

	// retrieves all expenses in the system (useful for our ledger math).
	ListAll(ctx context.Context) ([]*Expense, error)
}
