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
	_ = x[DOT-4]
	_ = x[STRING-5]
	_ = x[IDENT-6]
	_ = x[ACCOUNT-7]
	_ = x[DECIMAL-8]
	_ = x[DATE-9]
}

const _Token_name = "ILLEGALEOFCOMMENTNEWLINEDOTSTRINGIDENTACCOUNTDECIMALDATE"

var _Token_index = [...]uint8{0, 7, 10, 17, 24, 27, 33, 38, 45, 52, 56}

func (i Token) String() string {
	if i < 0 || i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}