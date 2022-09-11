# keeper

The keeper module provides accounting and bookkeeping tools for
programmers.

Keeper uses double entry accounting and stores numbers using a custom
absolute precision floating point scheme.  Specifically, all numbers
are stored as arbitrary sized integers.  Commodity/currency units have
an int64 scaling factor to represent fractional amounts.  For example,
1.23 USD is represented as 123, and the unit USD has a scale of 100.

The keeper module works with keeper files (with the `.kpr` extension).
Keeper files make it easy to record transactions and calculate and
reconcile account balances.

Keeper is designed for programmers.  Keeper files have a simple syntax
which can be generated from other sources, and the library returns the
data in a convenient format for further processing, as well as some
basic report generators.  Additionally, the library exposes the parser
and AST for keeper files.

Basic overview of usage:

1. Enter transactions into keeper files.
2. Reconcile balances.
3. Generate ledgers.
4. Generate trial balance.
5. Generate financial statements: income statement, capital statement,
   balance sheet, cash flow statement
