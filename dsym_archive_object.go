package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import "runtime"

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

func symbolicObjectGetFileFormat(object *C.SymbolicObject) (string, error) {
	C.symbolic_err_clear()

	str := C.symbolic_object_get_file_format(object)
	
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

func makeObject(cobj *C.SymbolicObject) (*Object, error) {
	goObj, err := makeObjectReqsFree(cobj)

	if err != nil {
		C.symbolic_object_free(cobj)
		return nil, err
	}

	runtime.SetFinalizer(goObj, func (o *Object) {
		C.symbolic_object_free(goObj.object)
	})

	return goObj, nil
}

func makeObjectReqsFree(cobj *C.SymbolicObject) (*Object, error) {
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

	goObj := &Object{
		object: cobj,
		arch: arch,
		codeId: codeId,
		debugId: debugId,
		kind: kind,
		fileFormat: fileFormat,
		features: features,
	}

	return goObj, nil
}
