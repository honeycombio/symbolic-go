package symbolic

/*
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type ProguardMapper struct {
	cspm        *C.SymbolicProguardMapper
	UUID        string
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

	C.symbolic_err_clear()
	uuidStr := C.symbolic_uuid_to_str(&uuid)
	err = checkErr()

	if err != nil {
		return nil, err
	}

	pm.UUID = decodeStr(&uuidStr)

	C.symbolic_err_clear()
	hasLineInfo := C.symbolic_proguardmapper_has_line_info(cspm)
	err = checkErr()

	if err != nil {
		return nil, err
	}

	pm.HasLineInfo = bool(hasLineInfo)

	return pm, nil
}

func (p *ProguardMapper) RemapFrame(class, method string, line int) ([]*SymbolicJavaStackFrame, error) {
	c := encodeStr(class)
	m := encodeStr(method)
	l := C.uintptr_t(line)
	params := encodeStr("")

	C.symbolic_err_clear()
	s := C.symbolic_proguardmapper_remap_frame(p.cspm, c, m, l, params, C._Bool(false))
	err := checkErr()

	if err != nil {
		return nil, err
	}

	frames := toSymbolicJavaStackFrames(&s)

	C.symbolic_proguardmapper_result_free(&s)

	return frames, nil
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

func (p *ProguardMapper) RemapMethod(class, method string) ([]*SymbolicJavaStackFrame, error) {
	c := encodeStr(class)
	m := encodeStr(method)

	C.symbolic_err_clear()
	s := C.symbolic_proguardmapper_remap_method(p.cspm, c, m)
	err := checkErr()

	if err != nil {
		return nil, err
	}

	r := toSymbolicJavaStackFrames(&s)

	C.symbolic_proguardmapper_result_free(&s)

	return r, nil
}

func freeProguardMapper(s *ProguardMapper) {
	C.symbolic_proguardmapper_free(s.cspm)
}

type SymbolicJavaStackFrame struct {
	ClassName      string
	MethodName     string
	LineNumber     int
	SourceFile     string
	ParameterNames string
}

func toSymbolicJavaStackFrames(s *C.SymbolicProguardRemapResult) []*SymbolicJavaStackFrame {
	frames := make([]*SymbolicJavaStackFrame, s.len)

	ptr := unsafe.Pointer(s.frames)

	for i := 0; i < int(s.len); i++ {
		frame := (*C.SymbolicJavaStackFrame)(ptr)
		frames[i] = &SymbolicJavaStackFrame{
			ClassName:      decodeStr(&frame.class_name),
			MethodName:     decodeStr(&frame.method),
			LineNumber:     int(frame.line),
			SourceFile:     decodeStr(&frame.file),
			ParameterNames: decodeStr(&frame.parameters),
		}
		ptr = unsafe.Add(ptr, C.sizeof_SymbolicJavaStackFrame)
	}

	return frames
}
