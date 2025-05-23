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
	SymCaches map[string]*SymCache
}

func (a *Archive) buildSymCaches() error {
	a.SymCaches = make(map[string]*SymCache)
	objects, err := a.objects()
	if (err != nil) {
		return err
	}

	for _,obj := range objects {
		symCache, err := NewSymCacheFromObject(&obj)
		if err != nil {
			return err
		}

		a.SymCaches[symCache.debugId] = symCache
	}

	return nil
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

func freeArchive(a *Archive) {
	C.symbolic_archive_free(a.archive)
}
