# Challenge 4 — Build It 🏗️
# "The ATM Machine"
## Problem
A bank needs a simple ATM simulation. Multiple customers walk up to the ATM one by one and perform actions. The ATM must process each action and report the result.
## Scope

- Create a BankAccount struct with Owner (string) and Balance (float64)
with these methods:

  - Deposit(amount float64) — adds money, no validation needed
  - Withdraw(amount float64) error — fails if balance would go below 0, return an error
  - Summary() — prints the account status (see output below)


- Create a slice of 3 BankAccounts with different starting balances
- Loop through them using range and for each account:
  - Print the Summary() to show original balance
  - Deposit some amount
  - Withdraw some amount (use an amount that fails for at least one account)
- After deposite and withdraw already, loop through them using range and for each account and print the Summary()

Note: This step must be done separately!

## Expected output
=== Tom ===
Balance: 1000.00
Deposit:  +500.00
Withdraw: -200.00

=== Anna ===
Balance: 50.00
Deposit:  +100.00
Withdraw: failed — not enough balance

=== Leo ===
Balance: 1000.00
Deposit:  +300.00
Withdraw: -900.00

Tom Balance: 1300.00
Anna Balance: 150.00
Leo Balance: 400.00

## Hints if needed

errors.New("not enough balance") to return an error from Withdraw
fmt.Printf("%.2f", amount) to print float with 2 decimal places
import "errors" to use errors.New()
The error return type: func (b *BankAccount) Withdraw(...) error
Check withdraw result: if err != nil { fmt.Println("failed —", err) }