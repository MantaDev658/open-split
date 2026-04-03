package domain

import (
	"testing"

	"opensplit/libs/go-core/money"
)

func TestNewExpense_Validation(t *testing.T) {
	total, _ := money.New(3000)
	split10, _ := money.New(1000)
	split20, _ := money.New(2000)

	tests := []struct {
		name        string
		payer       UserID
		splits      []Split
		expectError error
	}{
		{
			name:  "Valid expense",
			payer: "Alice",
			splits: []Split{
				{User: "Alice", Amount: split10},
				{User: "Bob", Amount: split20},
			},
			expectError: nil,
		},
		{
			name:  "Missing payer",
			payer: "",
			splits: []Split{
				{User: "Bob", Amount: total},
			},
			expectError: ErrMissingPayer,
		},
		{
			name:  "Splits don't add up",
			payer: "Alice",
			splits: []Split{
				{User: "Alice", Amount: split10},
				{User: "Bob", Amount: split10},
			},
			expectError: ErrSplitsDoNotEqualTotal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewExpense("exp-1", "Test", total, tt.payer, tt.splits)
			if err != tt.expectError {
				t.Errorf("got error %v, want %v", err, tt.expectError)
			}
		})
	}
}

func TestCalculateNetBalances(t *testing.T) {
	// Alice pays $30. Alice and Bob split it ($15 each).
	// Alice should be +15 (owed), Bob should be -15 (owes)
	total, _ := money.New(3000)
	split15, _ := money.New(1500)

	exp, _ := NewExpense("exp-1", "Dinner", total, "Alice", []Split{
		{User: "Alice", Amount: split15},
		{User: "Bob", Amount: split15},
	})

	balances := CalculateNetBalances([]*Expense{exp})

	if balances["Alice"] != 1500 {
		t.Errorf("Alice balance: got %d, want 1500", balances["Alice"])
	}
	if balances["Bob"] != -1500 {
		t.Errorf("Bob balance: got %d, want -1500", balances["Bob"])
	}
}
