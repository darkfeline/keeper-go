package raw

import (
	"cloud.google.com/go/civil"
	"go.felesatra.moe/keeper/book"
)

type Balance struct {
	Date    civil.Date
	Account book.Account
	Amounts []Amount
}

type Amount struct {
	Number Decimal
	Unit   string
}
