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

// Archive represents a potential multi arch object archive (like a dSYM)
type Archive struct {
	archive *C.SymbolicArchive
	symCaches map[string]*SymCache
}

// NewArchiveFromPath loads an archive from a given file path
func NewArchiveFromPath(path string) (*Archive, error) {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	C.symbolic_err_clear()
	a := C.symbolic_archive_open(c_path)
	err := checkErr()

	if err != nil {
		return nil, err
	}

	arch := &Archive{
		archive: a,
	}

	runtime.SetFinalizer(arch, freeArchive)

	return arch, nil
}

// NewArchiveFromBytes creates an archive from a byte buffer
func NewArchiveFromBytes(data []byte) (*Archive, error) {
	C.symbolic_err_clear()
	a := C.symbolic_archive_from_bytes((*C.uint8_t)(unsafe.Pointer(&data[0])), C.uintptr_t(len(data)))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	arch := &Archive{
		archive: a,
	}

	runtime.SetFinalizer(arch, freeArchive)

	return arch, nil
}

func ArchIPRegName(arch string) string {
	encoded := encodeStr(arch)
	res := C.symbolic_arch_ip_reg_name(encoded)
	return decodeStr(&res)
}

func FindBestInstruction(addr, ipRegValue uint64, signal uint32, arch string, crashingFrame bool) (uint64, error) {
	sii := (*C.SymbolicInstructionInfo)(C.malloc(C.sizeof_SymbolicInstructionInfo))
	defer C.free(unsafe.Pointer(sii))

	sii.addr = C.uint64_t(addr) 
	sii.arch = encodeStr(arch)
	sii.crashing_frame = C.bool(crashingFrame)
	sii.signal = C.uint32_t(signal)
	sii.ip_reg = C.uint64_t(ipRegValue)

	C.symbolic_err_clear()
	res := C.symbolic_find_best_instruction(sii)

	err := checkErr()
	if err != nil {
		return 0, err
	}

	return uint64(res), nil
}

func (a *Archive) ObjectCount() (int, error) {
	C.symbolic_err_clear()
	res := int(C.symbolic_archive_object_count(a.archive))

	err := checkErr()
	if (err != nil) {
		return 0, err
	}

	return res, nil
}

func (a *Archive) Objects() ([]Object, error) {
	count, err := a.ObjectCount()
	if err != nil {
		return nil, err
	}

	s := make([]Object, count)

	for i:= 0; i<count; i++ {
		C.symbolic_err_clear()
		cobj := C.symbolic_archive_get_object(a.archive, C.uintptr_t(i))

		err := checkErr()
		if (err != nil) {
			return nil, err
		}

		runtime.SetFinalizer(cobj, func (obj *C.SymbolicObject) {
			C.symbolic_object_free(obj)
		})


		s[i] = Object{ object: cobj }
	}
	return s, nil
}

// GetObject returns the n-th object, or nil if the object does not exist
func (a *Archive) GetObject(index int) (*Object, error) {
	C.symbolic_err_clear()
	obj := C.symbolic_archive_get_object(a.archive, C.uintptr_t(index))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(obj, func (obj *C.SymbolicObject) {
		C.symbolic_object_free(obj)
	})

	if obj == nil {
		return nil, nil
	}

	return &Object{object: obj}, nil
}

func (a *Archive) BuildSymCaches() error {
	a.symCaches = make(map[string]*SymCache)
	objects, err := a.Objects()
	if (err != nil) {
		return err
	}

	for _,obj := range objects {
		symCache, err := NewSymCacheFromObject(&obj)
		if err != nil {
			return err
		}

		a.symCaches[symCache.DebugID()] = symCache
	}

	return nil
}

func freeArchive(a *Archive) {
	C.symbolic_archive_free(a.archive)
}

// Object represents a single arch debug object
type Object struct {
	object *C.SymbolicObject
}

type ObjectFeatures struct {
	HasSymtab  bool
	HasDebug   bool
	HasUnwind  bool
	HasSources bool
}


func (o *Object) Free() {
	C.symbolic_object_free(o.object)
}

func (o *Object) Arch() (string, error) {
	C.symbolic_err_clear()
	str := C.symbolic_object_get_arch(o.object)

	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func (o *Object) CodeID() (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_code_id(o.object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func (o *Object) DebugID() (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_debug_id(o.object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func (o *Object) Kind() (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_kind(o.object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func (o *Object) FileFormat() (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_file_format(o.object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func (o *Object) Features() ObjectFeatures {
	features := C.symbolic_object_get_features(o.object)
	return ObjectFeatures{
		HasSymtab: bool(features.symtab),
		HasDebug:  bool(features.debug),
		HasUnwind: bool(features.unwind),
		HasSources: bool(features.sources),
	}
}

// SymCache represents a symbolic sym cache for fast symbol lookups
type SymCache struct {
	symcache *C.SymbolicSymCache
}

// SourceLocation represents a single symbol after lookup
type SourceLocation struct {
	SymAddr   uint64
	InstrAddr uint64
	Line      uint32
	Lang      string
	Symbol    string
	FullPath  string
}

// NewSymCacheFromObject creates a symcache from a given object
func NewSymCacheFromObject(object *Object) (*SymCache, error) {
	C.symbolic_err_clear()
	sc := C.symbolic_symcache_from_object(object.object)
	err := checkErr()

	if err != nil {
		return nil, err
	}

	symcache := &SymCache{
		symcache: sc,
	}

	runtime.SetFinalizer(symcache, func(s *SymCache) {
		s.freeSymCache()
	})

	return symcache, nil
}

// Arch returns the architecture of the symcache
func (s *SymCache) Arch() string {
	str := C.symbolic_symcache_get_arch(s.symcache)
	return decodeStr(&str)
}

// DebugID returns the debug identifier of the symcache
func (s *SymCache) DebugID() string {
	str := C.symbolic_symcache_get_debug_id(s.symcache)
	return decodeStr(&str)
}

// Version returns the version of the cache file
func (s *SymCache) Version() uint32 {
	return uint32(C.symbolic_symcache_get_version(s.symcache))
}

// Lookup looks up a single symbol at the given address
func (s *SymCache) Lookup(addr uint64) ([]SourceLocation, error) {
	C.symbolic_err_clear()
	result := C.symbolic_symcache_lookup(s.symcache, C.uint64_t(addr))

	err := checkErr()

	if err != nil {
		return nil, err
	}

	// todo: memory management
	// defer C.symbolic_lookup_result_free(&result)

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

func (s *SymCache) freeSymCache() {
	C.symbolic_symcache_free(s.symcache)
}