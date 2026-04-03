package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"
	"opensplit/libs/shared/money"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Save(ctx context.Context, expense *domain.Expense) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// ignore the rollback error, as it will naturally fail if we already committed.
	defer func() {
		_ = tx.Rollback()
	}()

	insertExpenseQuery := `
		INSERT INTO expenses (id, description, total_cents, payer_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, insertExpenseQuery,
		expense.ID(),
		expense.Description(),
		expense.TotalAmount().Int64(),
		expense.Payer(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert expense: %w", err)
	}

	insertSplitQuery := `
		INSERT INTO splits (expense_id, user_id, amount_cents)
		VALUES ($1, $2, $3)
	`
	for _, split := range expense.Splits() {
		_, err = tx.ExecContext(ctx, insertSplitQuery,
			expense.ID(),
			split.User,
			split.Amount.Int64(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert split for user %s: %w", split.User, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ExpenseRepository) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	// 1. Fetch the main Expense row
	var desc, payer string
	var totalCents int64
	err := r.db.QueryRowContext(ctx, "SELECT description, total_cents, payer_id FROM expenses WHERE id = $1", id).
		Scan(&desc, &totalCents, &payer)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrExpenseNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to query expense: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, "SELECT user_id, amount_cents FROM splits WHERE expense_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to query splits: %w", err)
	}
	defer rows.Close()

	var splits []domain.Split
	for rows.Next() {
		var userID string
		var amountCents int64
		if err := rows.Scan(&userID, &amountCents); err != nil {
			return nil, fmt.Errorf("failed to scan split: %w", err)
		}

		amount, _ := money.New(amountCents)
		splits = append(splits, domain.Split{
			User:   domain.UserID(userID),
			Amount: amount,
		})
	}

	totalMoney, _ := money.New(totalCents)
	return domain.NewExpense(id, desc, totalMoney, domain.UserID(payer), splits)
}

func (r *ExpenseRepository) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	query := `
		SELECT e.id, e.description, e.total_cents, e.payer_id, s.user_id, s.amount_cents
		FROM expenses e
		JOIN splits s ON e.id = s.expense_id
		ORDER BY e.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	type rawExpense struct {
		id          string
		description string
		totalCents  int64
		payer       string
		splits      []domain.Split
	}
	expenseMap := make(map[string]*rawExpense)
	var orderedIDs []string

	for rows.Next() {
		var expID, desc, payer, splitUser string
		var totalCents, splitCents int64

		if err := rows.Scan(&expID, &desc, &totalCents, &payer, &splitUser, &splitCents); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, exists := expenseMap[expID]; !exists {
			expenseMap[expID] = &rawExpense{
				id:          expID,
				description: desc,
				totalCents:  totalCents,
				payer:       payer,
			}
			orderedIDs = append(orderedIDs, expID)
		}

		splitAmt, _ := money.New(splitCents)
		expenseMap[expID].splits = append(expenseMap[expID].splits, domain.Split{
			User:   domain.UserID(splitUser),
			Amount: splitAmt,
		})
	}

	var results []*domain.Expense
	for _, id := range orderedIDs {
		raw := expenseMap[id]
		totalMoney, _ := money.New(raw.totalCents)
		exp, err := domain.NewExpense(
			domain.ExpenseID(raw.id),
			raw.description,
			totalMoney,
			domain.UserID(raw.payer),
			raw.splits,
		)
		if err != nil {
			return nil, fmt.Errorf("corrupted data in db for expense %s: %w", id, err)
		}
		results = append(results, exp)
	}

	return results, nil
}
