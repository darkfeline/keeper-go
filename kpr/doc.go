// Copyright (C) 2021  Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package kpr documents the syntax of keeper files.  Keeper files are
used for recording transactions for bookkeeping.  The semantics of the
files is not explicitly documented here.

Tokens

Keeper files consist of some basic token types.

Strings are like in most languages:

 "some string"
 "escape \" quote"

Unit symbols describe currencies and commodities, and consist of
uppercase letters:

 USD
 BTC

Account names start with an uppercase letter and can contain
alphanumeric characters, underscores, or colons.  An account name
cannot contain only uppercase letters as that would make it a unit
symbol:

 Some:account_123

Decimal numbers use periods as the decimal separator.  This is to
match programming languages and is not configurable (sorry comma
users).  Commas are ignored and may be used freely for grouping.
Numbers can start with a dash:

 -1,234.56

Dates are in ISO 8601 format:

 2000-01-31

There are some keywords:

 tx
 end
 balance
 unit
 disable
 account
 treebal

Comments are supported:

 # This is a comment.

Entries

Keeper files are comprised of single or multi line entries.

Unit entries declare units and their lowest division.  The following
declares that the smallest unit of USD is 1/100.  Only decimal
divisions are supported.

 unit USD 100

Transactions declare double entry bookkeeping transactions.  Amounts
can be omitted from splits:

 tx 2020-01-01 "Initial balance"
 Assets:Cash 100 USD
 Equity:Capital
 end

Balance assertions assert the balance of an account.  They can be
multi line for accounts that contain multiple unit types.

 balance 2020-01-01 Some:account 5 USD
 balance 2020-01-01 Some:account
 5 USD
 10 BTC
 end

Tree balance assertions are like normal balance assertions:

 treebal 2020-01-01 Some:account 5 USD
 treebal 2020-01-01 Some:account
 5 USD
 10 BTC
 end

Disable account entries disable an account for use:

 disable 2020-01-01 Some:account

Account declarations provide account metadata:

 account Some:account
 meta "my key" "my value"
 end
*/
package kpr
