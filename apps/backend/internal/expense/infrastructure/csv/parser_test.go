package csv

import (
	"os"
	"testing"
)

func TestParseExpenses_Strategies(t *testing.T) {
	// 1. Create a temporary CSV file with our new schema
	tempFile, err := os.CreateTemp("", "test_expenses_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 2. Define our test data covering every strategy and edge case
	csvData := `2026-04-01,Dinner,Food,9000,Alice,EVEN,Alice,Bob,Charlie
2026-04-02,Groceries,Food,10000,Bob,EXACT,Alice:2000,Bob:8000
2026-04-03,Hotel,Lodging,30000,Charlie,PERCENT,Alice:25,Bob:25,Charlie:50
2026-04-04,Drinks,Entertainment,1000,Alice,SHARES,Alice:1,Bob:1,Charlie:1
`
	if _, err := tempFile.WriteString(csvData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	// 3. Run the parser
	expenses, err := ParseExpenses(tempFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(expenses) != 4 {
		t.Fatalf("expected 4 expenses, got %d", len(expenses))
	}

	// --- 4. Assertions ---

	t.Run("EVEN Strategy", func(t *testing.T) {
		exp := expenses[0]
		if exp.TotalAmount().Int64() != 9000 {
			t.Errorf("Expected total 9000, got %d", exp.TotalAmount().Int64())
		}
		splits := exp.Splits()
		if len(splits) != 3 {
			t.Fatalf("Expected 3 splits, got %d", len(splits))
		}
		// $90 evenly split 3 ways is $30 each
		if splits[0].Amount.Int64() != 3000 {
			t.Errorf("Expected Alice to pay 3000, got %d", splits[0].Amount.Int64())
		}
	})

	t.Run("EXACT Strategy", func(t *testing.T) {
		exp := expenses[1]
		splits := exp.Splits()
		if len(splits) != 2 {
			t.Fatalf("Expected 2 splits, got %d", len(splits))
		}
		// Assert exact amounts mapped correctly
		if splits[0].User != "Alice" || splits[0].Amount.Int64() != 2000 {
			t.Errorf("Expected Alice:2000, got %s:%d", splits[0].User, splits[0].Amount.Int64())
		}
		if splits[1].User != "Bob" || splits[1].Amount.Int64() != 8000 {
			t.Errorf("Expected Bob:8000, got %s:%d", splits[1].User, splits[1].Amount.Int64())
		}
	})

	t.Run("PERCENT Strategy", func(t *testing.T) {
		exp := expenses[2]
		splits := exp.Splits()
		// $300 total. 25% = $75. 50% = $150.
		if splits[0].Amount.Int64() != 7500 { // Alice 25%
			t.Errorf("Expected Alice to pay 7500, got %d", splits[0].Amount.Int64())
		}
		if splits[2].Amount.Int64() != 15000 { // Charlie 50%
			t.Errorf("Expected Charlie to pay 15000, got %d", splits[2].Amount.Int64())
		}
	})

	t.Run("SHARES Strategy with Remainder", func(t *testing.T) {
		exp := expenses[3]
		splits := exp.Splits()
		// $10.00 split 3 ways (1 share each).
		// Math: 1000 / 3 = 333 with a remainder of 1.
		// Our logic gives the remainder penny to the first person.
		if splits[0].Amount.Int64() != 334 { // Alice gets the extra penny
			t.Errorf("Expected Alice to pay 334, got %d", splits[0].Amount.Int64())
		}
		if splits[1].Amount.Int64() != 333 {
			t.Errorf("Expected Bob to pay 333, got %d", splits[1].Amount.Int64())
		}
		if splits[2].Amount.Int64() != 333 {
			t.Errorf("Expected Charlie to pay 333, got %d", splits[2].Amount.Int64())
		}
	})
}
