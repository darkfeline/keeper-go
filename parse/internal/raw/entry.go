package raw

import (
	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

type BalanceEntry struct {
	Date    civil.Date
	Account book.Account
	Amounts []Amount
}

type Amount struct {
	Number Decimal
	Unit   string
}

type UnitEntry struct {
	Symbol string
	Scale  Decimal
}

type TransactionEntry struct {
	Date        civil.Date
	Description string
	Splits      []Split
}

type Split struct {
	Account book.Account
	Amount  Amount
}
