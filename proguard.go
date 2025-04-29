package symbolic

/*
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/google/uuid"
)

type ProguardMapper struct {
	cspm        *C.SymbolicProguardMapper
	UUID        *uuid.UUID
	HasLineInfo bool
}

func NewProguardMapper(path string) (*ProguardMapper, error) {
	p := C.CString(path)
	i := C._Bool(false)

	C.symbolic_err_clear()
	cspm := C.symbolic_proguardmapper_open(p, i)
	err := checkErr()

	if err != nil {
		return nil, err
	}

	pm := &ProguardMapper{
		cspm: cspm,
	}
	runtime.SetFinalizer(pm, freeProguardMapper)

	C.symbolic_err_clear()
	uuid := C.symbolic_proguardmapper_get_uuid(cspm)
	err = checkErr()

	if err != nil {
		return nil, err
	}

	id, err := toUUID(&uuid)

	if err != nil {
		return nil, err
	}

	pm.UUID = id

	C.symbolic_err_clear()
	hasLineInfo := C.symbolic_proguardmapper_has_line_info(cspm)
	err = checkErr()

	if err != nil {
		return nil, err
	}

	pm.HasLineInfo = bool(hasLineInfo)

	return pm, nil
}

func (p *ProguardMapper) RemapClass(class string) (string, error) {
	c := encodeStr(class)

	C.symbolic_err_clear()
	s := C.symbolic_proguardmapper_remap_class(p.cspm, c)
	err := checkErr()

	if err != nil {
		return "", err
	}

	return decodeStr(&s), nil
}

func freeProguardMapper(s *ProguardMapper) {
	C.symbolic_proguardmapper_free(s.cspm)
}

func toUUID(s *C.SymbolicUuid) (*uuid.UUID, error) {
	b := C.GoBytes(unsafe.Pointer(&s.data), 16)

	u, err := uuid.FromBytes(b)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
