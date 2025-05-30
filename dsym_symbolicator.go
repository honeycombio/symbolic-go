package symbolic

/*
#include <stdlib.h>
#include <string.h>
#include "include/symbolic.h"
*/
import "C"
import (
	"strings"
)

type Frame struct {
	Symbol *string
	SymbolLocation *uint64
	ImageOffset uint64
	ImageIndex int
}
type Thread struct {
	Frames []*Frame
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
	Threads []*Thread
	UsedImages []*Image
	Termination *Termination
}

type DSYMSymbolicator struct {
	Report *CrashReport
	Archive *Archive
}

func (symbolicator *DSYMSymbolicator) SymbolicateFrame(frame *Frame, thread *Thread, isCrashingFrame bool) ([]Frame, error) {
	imageOffset := frame.ImageOffset
	imageIndex := frame.ImageIndex
	image := symbolicator.Report.UsedImages[imageIndex]

	cache := symbolicator.Archive.SymCaches[image.UUID]

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
	addr, err := findBestInstruction(imageOffset, ipRegValue, symbolicator.Report.Termination.code, cache.arch, isCrashingFrame)
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

		symbol = demangle(symbol)

		res[idx] = Frame{
			Symbol: &symbol,
			SymbolLocation: &symAddr,
			ImageOffset: frame.ImageOffset,
			ImageIndex: frame.ImageIndex,
		}
	}

	return res, nil
}

var langSymbolicStr = encodeStr("Swift")
func demangle(symbol string) string {
	symbolSymbolicStr := encodeStr(symbol)

	demangledSymbol := C.symbolic_demangle(symbolSymbolicStr, langSymbolicStr)

	return decodeStr(&demangledSymbol)
}

func findBestInstruction(addr, ipRegValue uint64, signal uint32, arch string, crashingFrame bool) (uint64, error) {
	siiptr := C.malloc(C.sizeof_SymbolicInstructionInfo)
	sii := (*C.SymbolicInstructionInfo)(siiptr)
	defer C.free(siiptr)

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
