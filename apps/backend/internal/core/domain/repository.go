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
	Update(ctx context.Context, userID UserID, displayName string) error
	SoftDelete(ctx context.Context, userId UserID) error
}

type GroupRepository interface {
	Save(ctx context.Context, group *Group) error
	GetByID(ctx context.Context, id GroupID) (*Group, error)
	ListForUser(ctx context.Context, userID UserID) ([]*Group, error)
	UpdateName(ctx context.Context, id GroupID, name string) error
	Delete(ctx context.Context, id GroupID) error
	RemoveMember(ctx context.Context, id GroupID, userID UserID) error
}

type ExpenseRepository interface {
	Save(ctx context.Context, expense *Expense) error
	GetByID(ctx context.Context, id ExpenseID) (*Expense, error)
	ListAll(ctx context.Context) ([]*Expense, error)
	ListByGroup(ctx context.Context, groupID GroupID) ([]*Expense, error)
	ListNonGroupExpensesByUser(ctx context.Context, userID UserID) ([]*Expense, error)
	Update(ctx context.Context, expense *Expense) error
	Delete(ctx context.Context, id ExpenseID) error
}
