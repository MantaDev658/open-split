package csv

import (
	"os"
	"testing"
)

func TestParseExpenses(t *testing.T) {
	// 1. Create a temporary CSV file
	tempFile, err := os.CreateTemp("", "test_expenses_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up after the test

	// 2. Write valid test data ($30 split evenly between Alice, Bob, Charlie)
	csvData := "Alice,3000,Dinner,Alice,Bob,Charlie\n"
	if _, err := tempFile.WriteString(csvData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	// 3. Run our parser
	expenses, err := ParseExpenses(tempFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 4. Validate the resulting Domain objects
	if len(expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(expenses))
	}

	exp := expenses[0]
	if exp.Payer() != "Alice" {
		t.Errorf("expected payer Alice, got %s", exp.Payer())
	}
	if exp.TotalAmount().Int64() != 3000 {
		t.Errorf("expected total 3000, got %d", exp.TotalAmount().Int64())
	}
	if len(exp.Splits()) != 3 {
		t.Errorf("expected 3 splits, got %d", len(exp.Splits()))
	}
}
