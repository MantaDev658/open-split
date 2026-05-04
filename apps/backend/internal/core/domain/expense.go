package domain

import (
	"time"

	"opensplit/libs/shared/money"
)

// ExpenseID is the unique identifier for an Expense.
type ExpenseID string

// UserID is the unique identifier for a User.
type UserID string

// Split is a single participant's share of an expense.
type Split struct {
	User   UserID
	Amount money.Money
}

// Expense is the aggregate root representing a shared cost.
type Expense struct {
	id          ExpenseID
	groupID     *GroupID
	description string
	total       money.Money
	payer       UserID
	splits      []Split
	createdAt   time.Time
}

// NewExpense validates inputs and constructs an Expense.
func NewExpense(id ExpenseID, groupID *GroupID, desc string, total money.Money, payer UserID, splits []Split) (*Expense, error) {
	if payer == "" {
		return nil, ErrMissingPayer
	}
	if len(splits) == 0 {
		return nil, ErrNoSplits
	}

	var calculatedTotal money.Money = 0
	for _, split := range splits {
		calculatedTotal = calculatedTotal.Add(split.Amount)
	}

	if calculatedTotal != total {
		return nil, ErrSplitsDoNotEqualTotal
	}

	return &Expense{
		id:          id,
		groupID:     groupID,
		description: desc,
		total:       total,
		payer:       payer,
		splits:      splits,
		createdAt:   time.Now().UTC(),
	}, nil
}

// NewExpenseFromDB reconstitutes an expense from persistent storage.
// It trusts that the data is already valid and sets createdAt from the database value.
func NewExpenseFromDB(id ExpenseID, groupID *GroupID, desc string, total money.Money, payer UserID, splits []Split, createdAt time.Time) (*Expense, error) {
	e, err := NewExpense(id, groupID, desc, total, payer, splits)
	if err != nil {
		return nil, err
	}
	e.createdAt = createdAt
	return e, nil
}

// ID returns the expense's unique identifier.
func (e *Expense) ID() ExpenseID { return e.id }

// GroupID returns the group this expense belongs to, or nil for non-group expenses.
func (e *Expense) GroupID() *GroupID { return e.groupID }

// Description returns the human-readable label for this expense.
func (e *Expense) Description() string { return e.description }

// Total returns the full amount paid.
func (e *Expense) Total() money.Money { return e.total }

// Payer returns the user who paid upfront.
func (e *Expense) Payer() UserID { return e.payer }

// Splits returns each participant's allocated share.
func (e *Expense) Splits() []Split { return e.splits }

// CreatedAt returns when the expense was recorded.
func (e *Expense) CreatedAt() time.Time { return e.createdAt }

// CalculateNetBalances returns each user's net position across all expenses.
// A positive value means the user is owed money; negative means they owe money.
func CalculateNetBalances(expenses []*Expense) map[UserID]int64 {
	balances := make(map[UserID]int64)

	for _, exp := range expenses {
		balances[exp.Payer()] += exp.Total().Int64()
		for _, split := range exp.Splits() {
			balances[split.User] -= split.Amount.Int64()
		}
	}

	return balances
}
