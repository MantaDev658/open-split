package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
)

// Expense
type mockExpenseRepo struct {
	saveFunc        func(ctx context.Context, expense *domain.Expense) error
	listByGroupFunc func(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error)
}

func (m *mockExpenseRepo) Save(ctx context.Context, expense *domain.Expense) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, expense)
	}
	return nil
}
func (m *mockExpenseRepo) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	return nil, domain.ErrExpenseNotFound
}
func (m *mockExpenseRepo) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	return []*domain.Expense{}, nil
}
func (m *mockExpenseRepo) ListByGroup(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error) {
	if m.listByGroupFunc != nil {
		return m.listByGroupFunc(ctx, groupID)
	}
	return nil, nil
}
func (m *mockExpenseRepo) Update(ctx context.Context, expense *domain.Expense) error { return nil }
func (m *mockExpenseRepo) Delete(ctx context.Context, id domain.ExpenseID) error     { return nil }

// Group
type mockGroupRepo struct {
	getByIDFunc func(id domain.GroupID) (*domain.Group, error)
}

func (m *mockGroupRepo) Save(ctx context.Context, g *domain.Group) error { return nil }
func (m *mockGroupRepo) ListForUser(ctx context.Context, u domain.UserID) ([]*domain.Group, error) {
	return []*domain.Group{}, nil
}
func (m *mockGroupRepo) GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	return m.getByIDFunc(id)
}
func (m *mockGroupRepo) UpdateName(ctx context.Context, id domain.GroupID, name string) error {
	return nil
}
func (m *mockGroupRepo) Delete(ctx context.Context, id domain.GroupID) error { return nil }
func (m *mockGroupRepo) RemoveMember(ctx context.Context, id domain.GroupID, userID domain.UserID) error {
	return nil
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
			expenseRepo := &mockExpenseRepo{saveFunc: tt.mockSave}
			groupRepo := &mockGroupRepo{}
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
		eRepo := &mockExpenseRepo{}
		gRepo := &mockGroupRepo{
			getByIDFunc: func(id domain.GroupID) (*domain.Group, error) {
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
