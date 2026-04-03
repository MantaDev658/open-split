package domain

import (
	"reflect"
	"testing"
)

func TestSimplifyDebts(t *testing.T) {
	tests := []struct {
		name     string
		balances map[UserID]int64
		expected []Transaction
	}{
		{
			name:     "Everyone is settled",
			balances: map[UserID]int64{"Alice": 0, "Bob": 0},
			expected: nil,
		},
		{
			name: "One debtor, one creditor",
			balances: map[UserID]int64{
				"Alice": 1500,  // owed $15
				"Bob":   -1500, // owes $15
			},
			expected: []Transaction{
				{From: "Bob", To: "Alice", Amount: 1500},
			},
		},
		{
			name: "The Middleman Chain (A owes B, B owes C)",
			// if A owes B 10, and B owes C 10...
			// net: A is -10, B is 0, C is +10.
			balances: map[UserID]int64{
				"Alice":   -1000,
				"Bob":     0,
				"Charlie": 1000,
			},
			expected: []Transaction{
				{From: "Alice", To: "Charlie", Amount: 1000},
			},
		},
		{
			name: "Multiple debtors to one creditor",
			balances: map[UserID]int64{
				"Alice":   2000,  // owed $20
				"Bob":     -1500, // owes $15
				"Charlie": -500,  // owes $5
			},
			expected: []Transaction{
				{From: "Bob", To: "Alice", Amount: 1500},
				{From: "Charlie", To: "Alice", Amount: 500},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SimplifyDebts(tt.balances)

			if len(got) == 0 && len(tt.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SimplifyDebts() = %v, want %v", got, tt.expected)
			}
		})
	}
}
