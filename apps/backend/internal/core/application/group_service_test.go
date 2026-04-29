package application

import (
	"context"
	"strings"
	"testing"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/libs/shared/money"
)

func TestGroupService_CRUD(t *testing.T) {
	gRepo := &mockGroupRepo{}
	eRepo := &mockExpenseRepo{}
	service := NewGroupService(gRepo, eRepo)

	t.Run("UpdateGroup fails on empty name", func(t *testing.T) {
		err := service.UpdateGroup(context.Background(), "g1", "")
		if err == nil {
			t.Error("expected error for empty group name")
		}
	})

	t.Run("DeleteGroup succeeds", func(t *testing.T) {
		err := service.DeleteGroup(context.Background(), "g1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGroupService_RemoveMember_BalanceValidation(t *testing.T) {
	gRepo := &mockGroupRepo{}

	t.Run("Fails if user has an outstanding balance", func(t *testing.T) {
		// Mock an expense where UserA paid $30, split equally with UserB
		eRepo := &mockExpenseRepo{
			listByGroupFunc: func(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error) {
				total, _ := money.New(3000)
				split, _ := money.New(1500)
				exp, _ := domain.NewExpense(
					"exp-1", nil, "Dinner", total, "UserA",
					[]domain.Split{{User: "UserA", Amount: split}, {User: "UserB", Amount: split}},
				)
				return []*domain.Expense{exp}, nil
			},
		}

		service := NewGroupService(gRepo, eRepo)

		// UserB owes $15.00, they should NOT be allowed to leave.
		err := service.RemoveMember(context.Background(), "g1", "UserB")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "outstanding balance") {
			t.Errorf("expected balance error, got: %v", err)
		}
	})

	t.Run("Succeeds if user balance is exactly zero", func(t *testing.T) {
		// Mock an empty ledger (no expenses = $0.00 balance)
		eRepo := &mockExpenseRepo{}
		service := NewGroupService(gRepo, eRepo)

		err := service.RemoveMember(context.Background(), "g1", "UserC")
		if err != nil {
			t.Errorf("expected success, got: %v", err)
		}
	})
}
