// Code generated by "stringer -type=Token"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ILLEGAL-0]
	_ = x[EOF-1]
	_ = x[COMMENT-2]
	_ = x[NEWLINE-3]
	_ = x[STRING-4]
	_ = x[USYMBOL-5]
	_ = x[ACCTNAME-6]
	_ = x[DECIMAL-7]
	_ = x[DATE-8]
	_ = x[TX-9]
	_ = x[END-10]
	_ = x[BALANCE-11]
	_ = x[UNIT-12]
	_ = x[DISABLE-13]
	_ = x[ACCOUNT-14]
	_ = x[TREEBAL-15]
	_ = x[META-16]
}

const _Token_name = "ILLEGALEOFCOMMENTNEWLINESTRINGUSYMBOLACCTNAMEDECIMALDATETXENDBALANCEUNITDISABLEACCOUNTTREEBALMETA"

var _Token_index = [...]uint8{0, 7, 10, 17, 24, 30, 37, 45, 52, 56, 58, 61, 68, 72, 79, 86, 93, 97}

func (i Token) String() string {
	if i < 0 || i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}
