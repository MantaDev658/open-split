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
	repo domain.ExpenseRepository
}

func NewExpenseService(repo domain.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

type CreateExpenseCommand struct {
	Description string
	TotalCents  int64
	Payer       string
	Splits      map[string]int64
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

	expense, err := domain.NewExpense(
		domain.ExpenseID(uuid.NewString()),
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

	if err := s.repo.Save(dbCtx, expense); err != nil {
		return fmt.Errorf("infrastructure failure: %w", err)
	}

	return nil
}

func (s *ExpenseService) ListAllExpenses(ctx context.Context) ([]*domain.Expense, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	expenses, err := s.repo.ListAll(dbCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}

	return expenses, nil
}
