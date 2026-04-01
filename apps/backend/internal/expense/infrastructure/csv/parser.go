package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/google/uuid"
	"opensplit/apps/backend/internal/expense/domain"
	"opensplit/libs/go-core/money"
)

// ParseExpenses reads a CSV file and converts it into validated Domain Expenses
func ParseExpenses(filePath string) ([]*domain.Expense, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Allow variable number of columns per row (since the number of splitters varies)
	reader.FieldsPerRecord = -1

	var expenses []*domain.Expense
	lineNum := 0

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row %d: %w", lineNum, err)
		}

		// A valid row must have at least Payer, Amount, Description, and 1 Splitter
		if len(record) < 4 {
			return nil, fmt.Errorf("row %d has insufficient columns", lineNum)
		}

		payer := domain.UserID(record[0])
		cents, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount on row %d: %w", lineNum, err)
		}

		totalMoney, err := money.New(cents)
		if err != nil {
			return nil, fmt.Errorf("negative amount on row %d: %w", lineNum, err)
		}

		desc := record[2]
		splitterNames := record[3:]

		// Use our safe math library to distribute the pennies perfectly!
		splitAmounts := totalMoney.Distribute(len(splitterNames))

		var splits []domain.Split
		for i, name := range splitterNames {
			splits = append(splits, domain.Split{
				User:   domain.UserID(name),
				Amount: splitAmounts[i],
			})
		}

		expense, err := domain.NewExpense(
			domain.ExpenseID(uuid.NewString()),
			desc,
			totalMoney,
			payer,
			splits,
		)
		if err != nil {
			return nil, fmt.Errorf("domain validation failed on row %d: %w", lineNum, err)
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}
