package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"runtime"
	"strings"
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

	err = arch.buildSymCaches()
	if err != nil {
		return nil, err
	}

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

	err = arch.buildSymCaches()
	if err != nil {
		return nil, err
	}

	return arch, nil
}

func (a *Archive) objectCount() (int, error) {
	C.symbolic_err_clear()
	res := int(C.symbolic_archive_object_count(a.archive))

	err := checkErr()
	if (err != nil) {
		return 0, err
	}

	return res, nil
}

func (a *Archive) objects() ([]Object, error) {
	count, err := a.objectCount()
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


		obj, err := makeObject(cobj)
		if err != nil {
			return nil, err
		}
		s[i] = *obj
	}
	return s, nil
}

func (a *Archive) getObject(index int) (*Object, error) {
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

	return makeObject(obj)
}

func (a *Archive) buildSymCaches() error {
	a.symCaches = make(map[string]*SymCache)
	objects, err := a.objects()
	if (err != nil) {
		return err
	}

	for _,obj := range objects {
		symCache, err := NewSymCacheFromObject(&obj)
		if err != nil {
			return err
		}

		a.symCaches[symCache.debugId] = symCache
	}

	return nil
}

func freeArchive(a *Archive) {
	C.symbolic_archive_free(a.archive)
}

type Object struct {
	object *C.SymbolicObject
	arch string
	codeId string
	debugId string
	kind string
	fileFormat string
	features *ObjectFeatures
}

type ObjectFeatures struct {
	HasSymtab  bool
	HasDebug   bool
	HasUnwind  bool
	HasSources bool
}

func makeObject(cobj *C.SymbolicObject) (*Object, error) {
	arch, err := symbolicObjectGetArch(cobj)
	if err != nil {
		return nil, err
	}
	codeId, err := symbolicObjectGetCodeID(cobj)
	if err != nil {
		return nil, err
	}
	debugId, err := symbolicObjectGetDebugID(cobj)
	if err != nil {
		return nil, err
	}
	kind, err := symbolicObjectGetKind(cobj)
	if err != nil {
		return nil, err
	}
	fileFormat, err := symbolicObjectGetFileFormat(cobj)
	if err != nil {
		return nil, err
	}
	features, err := symbolicObjectGetFeatures(cobj)
	if err != nil {
		return nil, err
	}

	return &Object{
		object: cobj,
		arch: arch,
		codeId: codeId,
		debugId: debugId,
		kind: kind,
		fileFormat: fileFormat,
		features: features,
	}, nil
}


func (o *Object) Free() {
	C.symbolic_object_free(o.object)
}

func symbolicObjectGetArch(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()
	str := C.symbolic_object_get_arch(object)

	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symbolicObjectGetCodeID(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_code_id(object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symbolicObjectGetDebugID(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_debug_id(object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symbolicObjectGetKind(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_kind(object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symbolicObjectGetFileFormat(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_file_format(object)
	
	err := checkErr()
	if err != nil {
		return "", err
	}

	return decodeStr(&str), nil
}

func symbolicObjectGetFeatures(object *C.SymbolicObject) (*ObjectFeatures, error) {
	C.symbolic_err_clear()
	features := C.symbolic_object_get_features(object)

	err := checkErr()
	if err != nil {
		return nil, err
	}

	return &ObjectFeatures{
		HasSymtab: bool(features.symtab),
		HasDebug:  bool(features.debug),
		HasUnwind: bool(features.unwind),
		HasSources: bool(features.sources),
	}, nil
}

type SymCache struct {
	symcache *C.SymbolicSymCache
	arch string
	debugId string
	ipRegName string
}


func NewSymCacheFromObject(object *Object) (*SymCache, error) {
	C.symbolic_err_clear()
	sc := C.symbolic_symcache_from_object(object.object)
	err := checkErr()

	if err != nil {
		return nil, err
	}

	arch, err := symCacheGetArch(sc)
	if err != nil {
		return nil, err
	}

	debugId, err := symCacheGetDebugId(sc)
	if err != nil {
		return nil, err
	}

	ipRegName, err := archIPRegName(arch)
	if err != nil {
		return nil, err
	}

	symcache := &SymCache{
		symcache: sc,
		arch: arch,
		debugId: debugId,
		ipRegName: ipRegName,
	}

	runtime.SetFinalizer(symcache, func(s *SymCache) {
		s.freeSymCache()
	})

	return symcache, nil
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

	err := checkErr()

	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(&result,  func (obj *C.SymbolicLookupResult) {
		C.symbolic_lookup_result_free(obj)
	})

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

type Frame struct {
	Symbol *string
	SymbolLocation *uint64
	ImageOffset uint64
	ImageIndex int
}
type Thread struct {
	Frames []Frame
	ThreadState map[string]any
}
type Image struct {
	UUID string
	Base uint64
	Name string
}
type Termination struct {
	code uint32
}
type CrashReport struct {
	FaultingThread int
	Threads []Thread
	UsedImages []Image
	Termination Termination
}

type DSYMSymbolicator struct {
	Report CrashReport
	Archive Archive
}

func (symbolicator *DSYMSymbolicator) SymbolicateFrame(frame Frame, thread Thread, isCrashingFrame bool) ([]Frame, error) {
	imageOffset := frame.ImageOffset
	imageIndex := frame.ImageIndex
	image := symbolicator.Report.UsedImages[imageIndex]

	cache := symbolicator.Archive.symCaches[image.UUID]

	ipRegName, err := archIPRegName(cache.arch)
	if err != nil {
		return nil, err
	}
	ipRegState, found := thread.ThreadState[ipRegName]
	var ipRegValue uint64 = 0
	if found {
		ipMap := ipRegState.(map[string]any)
		ipRegValue = uint64(ipMap["value"].(float64))
	}
	addr, err := FindBestInstruction(imageOffset, ipRegValue, symbolicator.Report.Termination.code, cache.arch, isCrashingFrame)
	if err != nil {
		return nil, err
	}
	
	locations, err := cache.Lookup(addr)
	if err != nil {
		return nil, err
	}

	res := make([]Frame, len(locations))
	for idx,loc := range(locations) {
		symbol := strings.Clone(loc.Symbol)
		symAddr := loc.SymAddr

		res[idx] = Frame{
			Symbol: &symbol,
			SymbolLocation: &symAddr,
			ImageOffset: frame.ImageOffset,
			ImageIndex: frame.ImageIndex,
		}
	}

	return res, nil
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
