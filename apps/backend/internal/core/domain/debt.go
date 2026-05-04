package domain

import "sort"

// Transaction is a debt settlement instruction: From owes To the given Amount in cents.
type Transaction struct {
	From   UserID
	To     UserID
	Amount int64
}

// SimplifyDebts reduces a net-balance map to the minimum number of payments needed to settle all debts.
func SimplifyDebts(balances map[UserID]int64) []Transaction {
	type person struct {
		id      UserID
		balance int64
	}

	var debtors []person
	var creditors []person

	for id, bal := range balances {
		if bal < 0 {
			debtors = append(debtors, person{id, -bal})
		} else if bal > 0 {
			creditors = append(creditors, person{id, bal})
		}
	}

	sort.Slice(debtors, func(i, j int) bool { return debtors[i].balance > debtors[j].balance })
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].balance > creditors[j].balance })

	var transactions []Transaction
	i, j := 0, 0

	for i < len(debtors) && j < len(creditors) {
		debtor := &debtors[i]
		creditor := &creditors[j]

		settleAmount := min(creditor.balance, debtor.balance)

		transactions = append(transactions, Transaction{
			From:   debtor.id,
			To:     creditor.id,
			Amount: settleAmount,
		})

		debtor.balance -= settleAmount
		creditor.balance -= settleAmount

		if debtor.balance == 0 {
			i++
		}
		if creditor.balance == 0 {
			j++
		}
	}

	return transactions
}
