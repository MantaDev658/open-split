package domain

import (
	"errors"
	"time"

	"opensplit/libs/go-core/money"
)

type ExpenseID string
type UserID string

var (
	ErrSplitsDoNotEqualTotal = errors.New("the sum of splits must exactly equal the total expense amount")
	ErrMissingPayer          = errors.New("an expense must have a valid payer")
	ErrNoSplits              = errors.New("an expense must have at least one split")
)

type Split struct {
	User   UserID
	Amount money.Money
}

type Expense struct {
	id          ExpenseID
	description string
	totalAmount money.Money
	payer       UserID
	splits      []Split
	createdAt   time.Time
}

// NewExpense is the only way to create an Expense. It guarantees validity.
func NewExpense(id ExpenseID, desc string, total money.Money, payer UserID, splits []Split) (*Expense, error) {
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
		description: desc,
		totalAmount: total,
		payer:       payer,
		splits:      splits,
		createdAt:   time.Now().UTC(),
	}, nil
}

// Getters for infrastructure layer
func (e *Expense) ID() ExpenseID            { return e.id }
func (e *Expense) Description() string      { return e.description }
func (e *Expense) TotalAmount() money.Money { return e.totalAmount }
func (e *Expense) Payer() UserID            { return e.payer }
func (e *Expense) Splits() []Split          { return e.splits }

// CalculateNetBalances takes a list of expenses and returns a map of each user's net position.
// Positive = they are owed money. Negative = they owe money.
func CalculateNetBalances(expenses []*Expense) map[UserID]int64 {
	balances := make(map[UserID]int64)

	for _, exp := range expenses {
		// The payer's net position goes UP by the total amount they fronted
		balances[exp.Payer()] += exp.TotalAmount().Int64()

		// Every user's net position goes DOWN by the exact amount of their split
		for _, split := range exp.Splits() {
			balances[split.User] -= split.Amount.Int64()
		}
	}

	return balances
}
