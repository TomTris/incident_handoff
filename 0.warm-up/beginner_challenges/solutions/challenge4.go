package main

import (
	"errors"
	"fmt"
)

type BankAccount struct {
	Owner   string
	Balance float64
}

func (b *BankAccount) Deposit(amount float64) {
	b.Balance += amount
}

func (b *BankAccount) Withdraw(amount float64) error {
	if b.Balance < amount {
		return errors.New("not enough balance")
	}
	b.Balance -= amount
	return nil
}

func (b *BankAccount) Summary() {
	fmt.Printf("Balance: %.2f\n", b.Balance)
}

func main() {
	accounts := []BankAccount{
		{Owner: "Tom", Balance: 1000},
		{Owner: "Anna", Balance: 50},
		{Owner: "Leo", Balance: 1000},
	}

	deposits := []float64{500, 100, 300}
	withdrawals := []float64{200, 900, 900}

	for i := range accounts {
		// Go doesn't have reference like in C++, only copy or pointer like in C
		a := &accounts[i]

		fmt.Printf("=== %s ===\n", a.Owner)
		a.Summary()

		a.Deposit(deposits[i])
		fmt.Printf("Deposit:  +%.2f\n", deposits[i])

		err := a.Withdraw(withdrawals[i])
		if err != nil {
			fmt.Printf("Withdraw: failed — %s\n", err)
		} else {
			fmt.Printf("Withdraw: -%.2f\n", withdrawals[i])
		}

		fmt.Println()
	}
	for i := range accounts {
		fmt.Printf("%s ", accounts[i].Owner)
		accounts[i].Summary()
	}
}
