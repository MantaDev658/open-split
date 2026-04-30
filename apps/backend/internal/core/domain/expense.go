package domain

import (
	"time"

	"opensplit/libs/shared/money"
)

type ExpenseID string
type UserID string

type Split struct {
	User   UserID
	Amount money.Money
}

type Expense struct {
	id          ExpenseID
	groupID     *GroupID
	description string
	total       money.Money
	payer       UserID
	splits      []Split
	createdAt   time.Time
}

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

func (e *Expense) ID() ExpenseID       { return e.id }
func (e *Expense) GroupID() *GroupID   { return e.groupID }
func (e *Expense) Description() string { return e.description }
func (e *Expense) Total() money.Money  { return e.total }
func (e *Expense) Payer() UserID       { return e.payer }
func (e *Expense) Splits() []Split     { return e.splits }

func CalculateNetBalances(expenses []*Expense) map[UserID]int64 {
	balances := make(map[UserID]int64)

	for _, exp := range expenses {
		// the payer's net position goes UP by the total amount they fronted
		balances[exp.Payer()] += exp.Total().Int64()

		// every user's net position goes DOWN by the exact amount of their split
		for _, split := range exp.Splits() {
			balances[split.User] -= split.Amount.Int64()
		}
	}

	return balances
}
