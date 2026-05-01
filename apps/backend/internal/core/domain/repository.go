package domain

import (
	"context"
	"time"
)

type AuditLog struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	TargetID  string    `json:"target_id,omitempty"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           UserID
	DisplayName  string
	IsActive     bool
	PasswordHash string
}

type AuditRepository interface {
	Save(ctx context.Context, log AuditLog) error
	ListByGroup(ctx context.Context, groupID GroupID) ([]AuditLog, error)
}

type UserRepository interface {
	Save(ctx context.Context, user User) error
	GetByID(ctx context.Context, id UserID) (*User, error)
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
