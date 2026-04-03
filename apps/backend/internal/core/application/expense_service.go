package application

import (
	"context"
	"fmt"
	"time"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/libs/shared/money"

	"github.com/google/uuid"
)

type ExpenseService struct {
	expenseRepo domain.ExpenseRepository
	groupRepo   domain.GroupRepository
}

func NewExpenseService(eRepo domain.ExpenseRepository, gRepo domain.GroupRepository) *ExpenseService {
	return &ExpenseService{
		expenseRepo: eRepo,
		groupRepo:   gRepo,
	}
}

type CreateExpenseCommand struct {
	GroupID     string           `json:"group_id,omitempty"`
	Description string           `json:"description"`
	TotalCents  int64            `json:"total_cents"`
	Payer       string           `json:"payer"`
	Splits      map[string]int64 `json:"splits"`
}

func (s *ExpenseService) AddExpense(ctx context.Context, cmd CreateExpenseCommand) error {
	totalMoney, err := money.New(cmd.TotalCents)
	if err != nil {
		return fmt.Errorf("invalid total amount: %w", err)
	}

	var splits []domain.Split
	for user, cents := range cmd.Splits {
		splitMoney, _ := money.New(cents)
		splits = append(splits, domain.Split{
			User:   domain.UserID(user),
			Amount: splitMoney,
		})
	}
	var groupIDPtr *domain.GroupID

	if cmd.GroupID != "" {
		gID := domain.GroupID(cmd.GroupID)
		groupIDPtr = &gID

		group, err := s.groupRepo.GetByID(ctx, gID)
		if err != nil {
			return fmt.Errorf("failed to validate group: %w", err)
		}

		if !group.HasMember(domain.UserID(cmd.Payer)) {
			return fmt.Errorf("payer %s is not a member of group %s", cmd.Payer, group.Name)
		}

		for _, split := range splits {
			if !group.HasMember(split.User) {
				return fmt.Errorf("split participant %s is not a member of group %s", split.User, group.Name)
			}
		}
	}

	expense, err := domain.NewExpense(
		domain.ExpenseID(uuid.NewString()),
		groupIDPtr,
		cmd.Description,
		totalMoney,
		domain.UserID(cmd.Payer),
		splits,
	)
	if err != nil {
		return fmt.Errorf("business rule violation: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.expenseRepo.Save(dbCtx, expense); err != nil {
		return fmt.Errorf("infrastructure failure: %w", err)
	}

	return nil
}

func (s *ExpenseService) ListAllExpenses(ctx context.Context) ([]*domain.Expense, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	expenses, err := s.expenseRepo.ListAll(dbCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}

	return expenses, nil
}

func (s *ExpenseService) ListExpensesByGroup(ctx context.Context, groupID string) ([]*domain.Expense, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	expenses, err := s.expenseRepo.ListByGroup(dbCtx, domain.GroupID(groupID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group expenses: %w", err)
	}
	return expenses, nil
}
