package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type SymCache struct {
	symcache *C.SymbolicSymCache
	arch string
	debugId string
	ipRegName string
}

type SourceLocation struct {
	SymAddr   uint64
	InstrAddr uint64
	Line      uint32
	Lang      string
	Symbol    string
	FullPath  string
}

func (s *SymCache) Lookup(addr uint64) ([]SourceLocation, error) {
	C.symbolic_err_clear()
	result := C.symbolic_symcache_lookup(s.symcache, C.uint64_t(addr))
	defer C.symbolic_lookup_result_free(&result)

	err := checkErr()

	if err != nil {
		return nil, err
	}

	if result.items == nil || result.len == 0 {
		return []SourceLocation{}, nil
	}

	// Create a copy of all the data before freeing the C memory
	length := int(result.len)
	items := unsafe.Slice(result.items, length)
	sourceLocations := make([]SourceLocation, length)

	for i, item := range items {
		// Copy all values to our Go structs
		sourceLocations[i] = SourceLocation{
			SymAddr:   uint64(item.sym_addr),
			InstrAddr: uint64(item.instr_addr),
			Line:      uint32(item.line),
			Lang:      decodeStr(&item.lang),
			Symbol:    decodeStr(&item.symbol),
			FullPath:  decodeStr(&item.full_path),
		}
	}

	return sourceLocations, nil
}


func archIPRegName(arch string) (string, error) {
	C.symbolic_err_clear()
	encoded := encodeStr(arch)
	res := C.symbolic_arch_ip_reg_name(encoded)

	err := checkErr()

	if err != nil {
		return "", err
	}

	return decodeStr(&res), nil
}

func symCacheGetArch(symcache *C.SymbolicSymCache) (string, error) {
	C.symbolic_err_clear()
	str := C.symbolic_symcache_get_arch(symcache)

	err := checkErr()

	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symCacheGetDebugId(symcache *C.SymbolicSymCache) (string, error) {
	C.symbolic_err_clear()
	str := C.symbolic_symcache_get_debug_id(symcache)
	err := checkErr()

	if err != nil {
		return "", err
	}
	return decodeStr(&str), nil
}

func NewSymCacheFromObject(object *Object) (*SymCache, error) {
	C.symbolic_err_clear()
	sc := C.symbolic_symcache_from_object(object.object)
	err := checkErr()

	if err != nil {
		C.symbolic_symcache_free(sc)
		return nil, err
	}

	arch, err := symCacheGetArch(sc)
	if err != nil {
		C.symbolic_symcache_free(sc)
		return nil, err
	}

	debugId, err := symCacheGetDebugId(sc)
	if err != nil {
		C.symbolic_symcache_free(sc)
		return nil, err
	}

	ipRegName, err := archIPRegName(arch)
	if err != nil {
		C.symbolic_symcache_free(sc)
		return nil, err
	}

	symcache := &SymCache{
		symcache: sc,
		arch: arch,
		debugId: debugId,
		ipRegName: ipRegName,
	}
	runtime.SetFinalizer(symcache, func (s *SymCache) {
		C.symbolic_symcache_free(s.symcache)
	})

	return symcache, nil
}
