package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/apps/backend/internal/core/mocks"
	"opensplit/libs/shared/money"
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
		if !errors.Is(err, domain.ErrUserNotInGroup) {
			t.Errorf("expected ErrUserNotInGroup, got %v", err)
		}
	})
}

func TestExpenseService_SettleUp(t *testing.T) {
	eRepo := &mocks.MockExpenseRepo{
		SaveFunc: func(ctx context.Context, expense *domain.Expense) error {
			return nil
		},
	}
	gRepo := &mocks.MockGroupRepo{}
	service := NewExpenseService(eRepo, gRepo)

	t.Run("Fails if payer and receiver are the same", func(t *testing.T) {
		cmd := SettleUpCommand{
			PayerID:     "Alice",
			ReceiverID:  "Alice",
			AmountCents: 2000,
		}
		if err := service.SettleUp(context.Background(), cmd); err == nil {
			t.Error("expected error when payer equals receiver")
		}
	})

	t.Run("Fails if amount is zero", func(t *testing.T) {
		cmd := SettleUpCommand{
			PayerID:     "Alice",
			ReceiverID:  "Bob",
			AmountCents: 0,
		}
		if err := service.SettleUp(context.Background(), cmd); err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("Succeeds with valid parameters", func(t *testing.T) {
		cmd := SettleUpCommand{
			PayerID:     "Alice",
			ReceiverID:  "Bob",
			AmountCents: 1500,
		}
		if err := service.SettleUp(context.Background(), cmd); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestExpenseService_GetFriendBalances(t *testing.T) {
	eRepo := &mocks.MockExpenseRepo{
		ListNonGroupExpensesByUserFunc: func(ctx context.Context, userID domain.UserID) ([]*domain.Expense, error) {
			total, _ := money.New(3000)
			split, _ := money.New(1500)
			// Alice paid $30 for Alice and Bob
			exp, _ := domain.NewExpense("exp-1", nil, "Dinner", total, "Alice", []domain.Split{
				{User: "Alice", Amount: split}, {User: "Bob", Amount: split},
			})
			return []*domain.Expense{exp}, nil
		},
	}
	service := NewExpenseService(eRepo, &mocks.MockGroupRepo{})

	t.Run("Returns accurate settlements for user", func(t *testing.T) {
		settlements, err := service.GetFriendBalances(context.Background(), "Alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(settlements) != 1 {
			t.Fatalf("expected 1 settlement, got %d", len(settlements))
		}

		// Bob should owe Alice $15
		if string(settlements[0].From) != "Bob" || string(settlements[0].To) != "Alice" || settlements[0].Amount != 1500 {
			t.Errorf("incorrect settlement math: %+v", settlements[0])
		}
	})
}
