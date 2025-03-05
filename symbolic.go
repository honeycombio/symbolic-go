package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"fmt"
)

func init() {
	C.symbolic_init()
}

func encodeStr(s string) *C.SymbolicStr {
	c := C.CString(s)
	return &C.SymbolicStr{
		data:  c,
		len:   C.size_t(len(s)),
		owned: true,
	}
}

func decodeStr(s *C.SymbolicStr) string {
	str := C.GoStringN(s.data, C.int(C.strnlen(s.data, C.size_t(s.len))))

	if s.owned {
		C.symbolic_str_free(s)
	}

	return str
}

func checkErr() error {
	err := C.symbolic_err_get_last_code()

	// no error
	if err == 0 {
		return nil
	}

	msg := C.symbolic_err_get_last_message()
	bt := C.symbolic_err_get_backtrace()

	return &SymbolicError{
		ErrorCode: int(err),
		Message:   decodeStr(&msg),
		Backtrace: decodeStr(&bt),
	}
}

type SymbolicError struct {
	ErrorCode int
	Message   string
	Backtrace string
}

func (e *SymbolicError) Error() string {
	return fmt.Sprintf("Symbolic error %d: %s\n%s", e.ErrorCode, e.Message, e.Backtrace)
}
