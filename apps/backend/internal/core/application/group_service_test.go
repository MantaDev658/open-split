package application

import (
	"context"
	"errors"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/apps/backend/internal/core/mocks"
	"opensplit/libs/shared/money"
)

func TestGroupService_CRUD(t *testing.T) {
	gRepo := &mocks.MockGroupRepo{}
	eRepo := &mocks.MockExpenseRepo{}
	aRepo := &mocks.MockAuditRepo{}
	service := NewGroupService(gRepo, eRepo, aRepo)

	t.Run("UpdateGroup fails on empty name", func(t *testing.T) {
		err := service.UpdateGroup(context.Background(), "g1", "", "u1")
		if !errors.Is(err, domain.ErrEmptyGroupName) {
			t.Errorf("expected ErrEmptyGroupName, got %v", err)
		}
	})

	t.Run("DeleteGroup succeeds", func(t *testing.T) {
		err := service.DeleteGroup(context.Background(), "g1", "u1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGroupService_RemoveMember_BalanceValidation(t *testing.T) {
	aRepo := &mocks.MockAuditRepo{}
	gRepo := &mocks.MockGroupRepo{}

	t.Run("Fails if user has an outstanding balance", func(t *testing.T) {
		// Mock an expense where UserA paid $30, split equally with UserB
		eRepo := &mocks.MockExpenseRepo{
			ListByGroupFunc: func(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error) {
				total, _ := money.New(3000)
				split, _ := money.New(1500)
				exp, _ := domain.NewExpense(
					"exp-1", nil, "Dinner", total, "UserA",
					[]domain.Split{{User: "UserA", Amount: split}, {User: "UserB", Amount: split}},
				)
				return []*domain.Expense{exp}, nil
			},
		}

		service := NewGroupService(gRepo, eRepo, aRepo)

		// UserB owes $15.00, they should NOT be allowed to leave.
		err := service.RemoveMember(context.Background(), "g1", "UserB", "a1")
		if !errors.Is(err, domain.ErrOutstandingBalance) {
			t.Errorf("expected ErrOutstandingBalance, got %v", err)
		}
	})

	t.Run("Succeeds if user balance is exactly zero", func(t *testing.T) {
		// Mock an empty ledger (no expenses = $0.00 balance)
		eRepo := &mocks.MockExpenseRepo{}
		service := NewGroupService(gRepo, eRepo, aRepo)

		err := service.RemoveMember(context.Background(), "g1", "UserC", "a1")
		if err != nil {
			t.Errorf("expected success, got: %v", err)
		}
	})
}
