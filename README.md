# Open Split 💸

Open Split is a fast,lightweight, and privacy-respecting CLI utility for calculating group expenses and simplifying debts. It is a local-first, open-source alternative to Splitwise.

## 🚀 Features

* **Debt Simplification:** Uses a greedy algorithm to minimize the total number of transactions needed to settle a group's debts by default.
* **Penny-Perfect Math:** Custom `money` engine ensures that $10.00 split 3 ways never results in missing or fabricated cents.
* **Complex Split Strategies:** Supports splitting by Even distributions, Exact amounts, Percentages, and fractional Shares.

---

## 📂 Project Structure

Open Split is structured as a Go Workspace (`go.work`) monorepo to separate core domain libraries from the application logic.

## 🛠️ Quick Start & Installation
### Prerequisites

* Go 1.21+
* `make` (Standard on macOS/Linux)

### 1. Clone & Build

```Bash
git clone https://github.com/yourusername/opensplit.git
cd opensplit

# Sync the Go workspace and install linter tools
make sync
make setup

# Compile the single binary
make build
```

### 2. Run the App

```Bash
./bin/opensplit-cli -file=test_expenses.csv
```

## 📄 The CSV Ledger Format
Open Split reads expenses from a structured CSV file. The parser is strict to ensure mathematical accuracy.

**Important**: All currency amounts must be entered in **cents (pennies)** to prevent floating-point errors. For example, $50.00 must be written as `5000`.

### Column Schema
Every row must contain a minimum of 7 columns:

| Col | Name | Example | Description |
| :--- | :--- | :--- | :--- |
| **1** | **Date** | `2026-04-01` | The date of the transaction (YYYY-MM-DD). |
| **2** | **Description** | `Dinner` | A label for the expense. |
| **3** | **Category** | `Food` | A category for future analytics. |
| **4** | **TotalCents** | `9000` | The total cost in pennies (e.g., 9000 = $90.00). |
| **5** | **Payer** | `Alice` | The name of the user who paid the bill. |
| **6** | **Strategy** | `EVEN` | The split method (`EVEN`, `EXACT`, `PERCENT`, `SHARES`). |
| **7+** | **Participants** | `Alice, Bob` | Dynamic columns based on the chosen strategy. |

### Splitting Strategies

* `EVEN`: Splits the total evenly among all listed participants.
    * Example: `...,EVEN,Alice,Bob,Charlie` (Splits evenly between the three).

* `EXACT`: Define the exact amount (in cents) each person owes. The sum must equal `TotalCents`.

    * Syntax: `Name:Cents`

    * Example: `...,EXACT,Alice:2000,Bob:8000` (Alice owes $20, Bob owes $80).

* `PERCENT`: Calculates the split based on relative percentages.

    * Syntax: `Name:Percentage`

    * Example: `...,PERCENT,Alice:25,Bob:75` (Bob pays 75% of the total).

* `SHARES`: Splits cost based on proportional parts (e.g., "Alice had 2 drinks, Bob had 1").

    * Syntax: `Name:Shares`

    * Example: `...,SHARES,Alice:2,Bob:1` (Alice pays 2/3, Bob pays 1/3).

## 🧪 Example Usage
### 1. Create a file named `my_trip.csv`:

```Code snippet
2026-04-01,Dinner,Food,9000,Alice,EVEN,Alice,Bob,Charlie
2026-04-02,Groceries,Food,10000,Bob,EXACT,Alice:2000,Bob:8000
2026-04-03,Hotel,Lodging,30000,Charlie,PERCENT,Alice:25,Bob:25,Charlie:50
2026-04-04,Drinks,Entertainment,1000,Alice,SHARES,Alice:1,Bob:1,Charlie:1
```

### 2. Run the CLI:

```Bash
./bin/opensplit-cli -file=my_trip.csv
```

### 3. Expected Output:

```Plaintext
📂 Loading expenses from: my_trip.csv

--- 💸 Open Split: Settle Up Instructions ---
➡️  Bob        pays Charlie    $  23.33
➡️  Alice      pays Charlie    $  21.67
➡️  Alice      pays Bob        $  20.00
-------------------------------------------
```

To see raw balances without the simplification algorithm, use the `--no-simplify` flag.

## 💻 Development & Contributing
Open Split uses strict formatting and linting rules enforced by a Git pre-commit hook to maintain high code quality.

### Useful Make Commands:

* `make test`: Runs the test suite across all modules (Domain, Logic, Parser).

* `make lint`: Runs golangci-lint to check for code smells and complexity.

* `make all`: Full CI check (Formatting -> Linting -> Race Condition Testing).