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

type SplitDetail struct {
	UserID string  `json:"user_id"`
	Value  float64 `json:"value"`
}

type CreateExpenseCommand struct {
	GroupID     string        `json:"group_id,omitempty"`
	Description string        `json:"description"`
	TotalCents  int64         `json:"total_cents"`
	Payer       string        `json:"payer"`
	SplitType   string        `json:"split_type"`
	Splits      []SplitDetail `json:"splits"`
}

type UpdateExpenseCommand struct {
	ID          string        `json:"id"`
	GroupID     string        `json:"group_id,omitempty"`
	Description string        `json:"description"`
	TotalCents  int64         `json:"total_cents"`
	Payer       string        `json:"payer"`
	SplitType   string        `json:"split_type"`
	Splits      []SplitDetail `json:"splits"`
}

type SettleUpCommand struct {
	GroupID     string `json:"group_id,omitempty"`
	PayerID     string `json:"payer_id"`
	ReceiverID  string `json:"receiver_id"`
	AmountCents int64  `json:"amount_cents"`
}

func (s *ExpenseService) buildAndValidateExpense(ctx context.Context, id string, groupID string, desc string, totalCents int64, payer string, splitType string, inputSplits []SplitDetail) (*domain.Expense, error) {
	var domainInputs []domain.AllocationInput
	for _, split := range inputSplits {
		domainInputs = append(domainInputs, domain.AllocationInput{
			UserID: domain.UserID(split.UserID),
			Value:  split.Value,
		})
	}

	splits, err := domain.Allocate(domain.AllocationType(splitType), totalCents, domainInputs)
	if err != nil {
		return nil, fmt.Errorf("allocation math error: %w", err)
	}

	totalMoney, err := money.New(totalCents)
	if err != nil {
		return nil, domain.ErrInvalidTotal
	}

	var groupIDPtr *domain.GroupID
	if groupID != "" {
		gID := domain.GroupID(groupID)
		groupIDPtr = &gID

		group, groupErr := s.groupRepo.GetByID(ctx, gID)
		if groupErr != nil {
			return nil, fmt.Errorf("failed to validate group: %w", groupErr)
		}

		if !group.HasMember(domain.UserID(payer)) {
			return nil, fmt.Errorf("%w: payer %s is not a member of group %s", domain.ErrUserNotInGroup, payer, groupID)
		}

		for _, split := range splits {
			if !group.HasMember(split.User) {
				return nil, fmt.Errorf("split participant %s is not a member of group %s", split.User, group.Name)
			}
		}
	}

	return domain.NewExpense(domain.ExpenseID(id), groupIDPtr, desc, totalMoney, domain.UserID(payer), splits)
}

func (s *ExpenseService) AddExpense(ctx context.Context, cmd CreateExpenseCommand) error {
	expense, err := s.buildAndValidateExpense(ctx, uuid.NewString(), cmd.GroupID, cmd.Description, cmd.TotalCents, cmd.Payer, cmd.SplitType, cmd.Splits)
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

func (s *ExpenseService) GetFriendBalances(ctx context.Context, userID string) ([]domain.Transaction, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	expenses, err := s.expenseRepo.ListNonGroupExpensesByUser(dbCtx, domain.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch friend expenses: %w", err)
	}

	balances := domain.CalculateNetBalances(expenses)
	suggestions := domain.SimplifyDebts(balances)

	var userDebts []domain.Transaction
	for _, debt := range suggestions {
		if string(debt.From) == userID || string(debt.To) == userID {
			userDebts = append(userDebts, debt)
		}
	}

	if userDebts == nil {
		return []domain.Transaction{}, nil
	}

	return userDebts, nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, cmd UpdateExpenseCommand) error {
	expense, err := s.buildAndValidateExpense(ctx, cmd.ID, cmd.GroupID, cmd.Description, cmd.TotalCents, cmd.Payer, cmd.SplitType, cmd.Splits)
	if err != nil {
		return fmt.Errorf("business rule violation: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.expenseRepo.Update(dbCtx, expense)
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.expenseRepo.Delete(dbCtx, domain.ExpenseID(id)); err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	return nil
}

func (s *ExpenseService) SettleUp(ctx context.Context, cmd SettleUpCommand) error {
	if cmd.PayerID == cmd.ReceiverID {
		return domain.ErrSamePayerReceiver
	}
	if cmd.AmountCents <= 0 {
		return domain.ErrInvalidSettlementAmount
	}

	createCmd := CreateExpenseCommand{
		GroupID:     cmd.GroupID,
		Description: "Payment",
		TotalCents:  cmd.AmountCents,
		Payer:       cmd.PayerID,
		SplitType:   string(domain.AllocationTypeExact),
		Splits: []SplitDetail{
			{UserID: cmd.ReceiverID, Value: float64(cmd.AmountCents)},
		},
	}

	return s.AddExpense(ctx, createCmd)
}
