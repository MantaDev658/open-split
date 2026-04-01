package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"opensplit/apps/backend/internal/expense/domain"
	"opensplit/apps/backend/internal/expense/infrastructure/csv"
)

func main() {
	// 1. Define CLI flags
	filePath := flag.String("file", "", "Path to the CSV file containing expenses")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: opensplit-cli -file=<path_to_csv>")
		os.Exit(1)
	}

	// 2. Parse the CSV into Domain Objects
	fmt.Printf("📂 Loading expenses from: %s\n", *filePath)
	expenses, err := csv.ParseExpenses(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing CSV: %v\n", err)
		os.Exit(1)
	}

	// 3. Calculate Balances using our Pure Domain Logic
	// This function lives in our 'internal/expense/domain' package
	netBalances := domain.CalculateNetBalances(expenses)

	// 4. Print the Results (The Ledger)
	fmt.Println("\n--- 📊 OpenSplit Ledger Summary ---")

	// Sort users by name for consistent output
	var users []domain.UserID
	for user := range netBalances {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i] < users[j]
	})

	for _, user := range users {
		balance := netBalances[user]
		formattedBalance := float64(balance) / 100.0 // Convert cents back to display units

		if balance > 0 {
			fmt.Printf("✅ %-10s is OWED:   $%7.2f\n", user, formattedBalance)
		} else if balance < 0 {
			// Multiply by -1 to show positive "Owes" amount
			fmt.Printf("🔴 %-10s OWES:       $%7.2f\n", user, -formattedBalance)
		} else {
			fmt.Printf("⚪ %-10s is SETTLED: $%7.2f\n", user, formattedBalance)
		}
	}
	fmt.Println("-----------------------------------")
}
