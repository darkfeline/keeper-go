// Copyright (C) 2019  Allen Li
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

// Package parse implements parsing of keeper entries.
//
// keeper entries have a simple, regular grammar.
// Each entry starts with a keyword.
// Entries can be one or more lines long.
// Single line entries look like:
//
//   keyword [ARGS...]
//
// Multiline entries look like:
//
//    keyword [ARGS...]
//    [ARGS...]
//    .
//
// Comments start with # and go to the end of the line.
// Empty lines are ignored.
//
// Entry types
//
// Unit entries declare a unit type:
//
//   unit USD 100
//
// Units must be declared.
//
// The unit symbol must be all uppercase ASCII letters.
// The symbol is followed by the scale factor.
// The scale factor indicates what the smallest fraction is.
// In this example, the smallest fraction is 1/100 (one US cent).
// The scale should be a multiple of 10.
//
// Transaction entries are the main type of entry.
// They define a bookkeeping transaction:
//
//   tx 2001-02-03 "Some description"
//   Some:account -1.20 USD
//   Other:account 1.20 USD
//   .
//
// Transactions must balance.
// The amount can be elided for at most one of the splits,
// when there is only one currency that needs balancing.
//
// This works:
//
//   tx 2001-02-03 "Some description"
//   Some:account -1.20 USD
//   Other:account
//   .
//
// This does not work, since multiple splits are empty:
//
//   tx 2001-02-03 "Some description"
//   Some:account -1.20 USD
//   Other:account
//   Other:account2
//
// This does not work, since more than one unit is unbalanced:
//
//   tx 2001-02-03 "Some description"
//   Some:account -1.20 USD
//   Other:account -200 JPY
//   Other:account2
//   .
//
// Balance entries checks the balance for an account:
//
//   balance 2001-02-03 Some:account 100 USD
//
// A multiline format can be used for multiple unit types:
//
//   balance 2001-02-03 Some:account
//   100 USD
//   500 JPY
//   .
//
// bal can be used instead of balance for the entry keyword.
//
// Ordering
//
// Unit entries can be anywhere.
// They can come after any reference to the unit.
//
// All dated entries are sorted by date.
// Balance entries are checked after all transaction entries for the same date.
package parse
