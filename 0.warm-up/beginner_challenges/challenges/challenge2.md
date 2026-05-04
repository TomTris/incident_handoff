# "The Beer Eligibility Checker"

## Problem
A bar needs a small program to check a list of customers and decide who can buy beer.

## Scope
- Create a Customer struct with Name (string) and Age (int)
- Add a method CanBuyBeer() that returns true if age >= 18
- Create a slice of at least 4 customers (mix of ages)
- Loop through them using range and print the result

## Expected Output
- Tom (20) → can buy beer: true
- Anna (16) → can buy beer: false
- Leo (18) → can buy beer: true
- Kim (15) → can buy beer: false

## Hints
Use fmt.Printf for formatted output
%s = string, %d = int, %v = bool