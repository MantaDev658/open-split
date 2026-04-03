package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
)

type mockRepo struct {
	saveFunc func(ctx context.Context, expense *domain.Expense) error
}

func (m *mockRepo) Save(ctx context.Context, expense *domain.Expense) error {
	return m.saveFunc(ctx, expense)
}
func (m *mockRepo) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	return nil, nil // Not needed for this test
}
func (m *mockRepo) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	return nil, nil // Not needed for this test
}

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
			repo := &mockRepo{saveFunc: tt.mockSave}
			service := NewExpenseService(repo)

			err := service.AddExpense(context.Background(), tt.cmd)
			if (err != nil) != tt.expectError {
				t.Errorf("AddExpense() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
