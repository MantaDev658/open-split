package domain

import "sort"

type Transaction struct {
	From   UserID
	To     UserID
	Amount int64 // Stored in cents
}

func SimplifyDebts(balances map[UserID]int64) []Transaction {
	type person struct {
		id      UserID
		balance int64
	}

	var debtors []person
	var creditors []person

	for id, bal := range balances {
		if bal < 0 {
			// store debt as a positive number to make the math easier
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
