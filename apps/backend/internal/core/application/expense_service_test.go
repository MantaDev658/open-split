package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/apps/backend/internal/core/mocks"
)

func TestExpenseService_AddExpense(t *testing.T) {
	tests := []struct {
		name        string
		cmd         CreateExpenseCommand
		mockSave    func(ctx context.Context, expense *domain.Expense) error
		expectError bool
	}{
		{
			name: "Success path",
			cmd: CreateExpenseCommand{
				Description: "Dinner",
				TotalCents:  3000,
				Payer:       "Alice",
				Splits:      map[string]int64{"Alice": 1500, "Bob": 1500},
			},
			mockSave: func(ctx context.Context, expense *domain.Expense) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "Domain validation failure (math doesn't add up)",
			cmd: CreateExpenseCommand{
				Description: "Dinner",
				TotalCents:  3000,
				Payer:       "Alice",
				Splits:      map[string]int64{"Alice": 1000, "Bob": 1000}, // only $20 out of $30
			},
			mockSave: func(ctx context.Context, expense *domain.Expense) error {
				t.Fatal("Save should never be called if domain validation fails")
				return nil
			},
			expectError: true,
		},
		{
			name: "Infrastructure failure (DB is down)",
			cmd: CreateExpenseCommand{
				Description: "Dinner",
				TotalCents:  3000,
				Payer:       "Alice",
				Splits:      map[string]int64{"Alice": 3000},
			},
			mockSave: func(ctx context.Context, expense *domain.Expense) error {
				return errors.New("database connection refused")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expenseRepo := &mocks.MockExpenseRepo{SaveFunc: tt.mockSave}
			groupRepo := &mocks.MockGroupRepo{}
			service := NewExpenseService(expenseRepo, groupRepo)

			err := service.AddExpense(context.Background(), tt.cmd)
			if (err != nil) != tt.expectError {
				t.Errorf("AddExpense() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestExpenseService_AddExpense_WithGroups(t *testing.T) {
	t.Run("Fails if payer is not in group", func(t *testing.T) {
		eRepo := &mocks.MockExpenseRepo{}
		gRepo := &mocks.MockGroupRepo{
			GetByIDFunc: func(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
				return &domain.Group{ID: id, Name: "Ski Trip", Members: []domain.UserID{"Bob"}}, nil
			},
		}
		service := NewExpenseService(eRepo, gRepo)

		cmd := CreateExpenseCommand{
			GroupID:     "some-uuid",
			Description: "Dinner",
			Payer:       "Alice", // Alice is NOT in the members list above
			Splits:      map[string]int64{"Alice": 1000},
		}

		err := service.AddExpense(context.Background(), cmd)
		if err == nil {
			t.Error("expected error when payer is not in group, got nil")
		}
	})
}
