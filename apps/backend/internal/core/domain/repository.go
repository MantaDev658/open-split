package domain

import (
	"context"
	"errors"
)

var (
	ErrExpenseNotFound = errors.New("expense not found")
	ErrUserNotFound    = errors.New("user not found")
)

type User struct {
	ID          UserID
	DisplayName string
}

type UserRepository interface {
	Save(ctx context.Context, user User) error
	ListAll(ctx context.Context) ([]User, error)
}

type ExpenseRepository interface {
	Save(ctx context.Context, expense *Expense) error
	GetByID(ctx context.Context, id ExpenseID) (*Expense, error)
	ListAll(ctx context.Context) ([]*Expense, error)
}
