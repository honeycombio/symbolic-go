package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

func init() {
	C.symbolic_init()
}

type SourceMapCache struct {
	ssmc *C.SymbolicSourceMapCache
}

func NewSourceMapCache(source, sourceMap string) (*SourceMapCache, error) {
	cs := C.CString(source)
	csm := C.CString(sourceMap)

	C.symbolic_err_clear()
	ssmc := C.symbolic_sourcemapcache_from_bytes(cs, C.strlen(cs), csm, C.strlen(csm))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	s := &SourceMapCache{
		ssmc: ssmc,
	}

	runtime.SetFinalizer(s, free)

	return s, nil
}

func (s *SourceMapCache) Lookup(line, col, contextLines uint32) (*SourceMapCacheToken, error) {
	C.symbolic_err_clear()
	match := C.symbolic_sourcemapcache_lookup_token(s.ssmc, C.uint32_t(line), C.uint32_t(col), C.uint32_t(contextLines))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	defer C.symbolic_sourcemapcache_token_match_free(match)

	smct := newSourceMapCacheToken(match)

	return smct, nil
}

func free(s *SourceMapCache) {
	C.symbolic_sourcemapcache_free(s.ssmc)
}

type SourceMapCacheToken struct {
	Line         int
	Col          int
	Src          string
	Name         string
	FunctionName string
	ContextLine  string
	PreContext   []string
	PostContext  []string
}

func newSourceMapCacheToken(match *C.SymbolicSmTokenMatch) *SourceMapCacheToken {
	pre := make([]string, match.pre_context.len)
	for i, s := range unsafe.Slice(match.pre_context.strs, match.pre_context.len) {
		pre[i] = decodeStr(&s)
	}

	post := make([]string, match.post_context.len)
	for i, s := range unsafe.Slice(match.post_context.strs, match.post_context.len) {
		post[i] = decodeStr(&s)
	}

	return &SourceMapCacheToken{
		Line:         int(match.line),
		Col:          int(match.col),
		Src:          decodeStr(&match.src),
		Name:         decodeStr(&match.name),
		FunctionName: decodeStr(&match.function_name),
		ContextLine:  decodeStr(&match.context_line),
		PreContext:   pre,
		PostContext:  post,
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
