package symbolic

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDSymArchiveElectron(t *testing.T) {
	binaryPath := "symbolic/py/tests/res/electron/1.8.1/Electron/CB63147AC9DC308B8CA1EE92A5042E8E0/Electron.app.dSYM/Contents/Resources/DWARF/Electron"

	// Verify the file exists
	_, err := os.Stat(binaryPath)
	assert.NoError(t, err, "DWARF binary file not found")

	archive, err := NewArchiveFromPath(binaryPath)
	assert.NoError(t, err, "Failed to load DWARF binary")

	// Get basic information about the archive
	count := archive.ObjectCount()
	assert.Equal(t, count, 1, "Expected at least one object in the archive")

	// Get the first object
	obj, err := archive.GetObject(0)
	assert.NoError(t, err, "Failed to get object")
	assert.NotNil(t, obj, "Object is nil")

	
	for _,obj := range archive.Objects() {
		// Check object properties
		assert.Equal(t, obj.Arch(), "x86_64")

		// Create a symcache from the object
		symCache, err := NewSymCacheFromObject(&obj)
		assert.NoError(t, err, "Failed to create symcache")

		// Verify a known symbol
		locations, err := symCache.Lookup(0x107BB9F25 - 0x107BB9000)
		assert.NoError(t, err, "Symbol lookup failed")
		
		symbol := locations[0]
		assert.Equal(t, symbol.Symbol, "main")
		assert.Equal(t, symbol.Lang, "cpp")
		assert.Equal(t, symbol.Line, uint32(186))
	}
}

func TestDSymArchive(t *testing.T) {
	dsymPath := "crashcrashcrash.app.dSYM"
	
	// Find the actual DWARF file within the dSYM
	dwarfBinaryPath := dsymPath + "/Contents/Resources/DWARF/crashcrashcrash"
	
	// Verify the file exists
	_, err := os.Stat(dwarfBinaryPath)
	assert.NoError(t, err, "DWARF binary file not found")

	archive, err := NewArchiveFromPath(dwarfBinaryPath)
	assert.NoError(t, err, "Failed to load DWARF binary")

	// Get basic information about the archive
	count := archive.ObjectCount()
	assert.GreaterOrEqual(t, count, 1, "Expected at least one object in the archive")

	// Get the first object
	obj, err := archive.GetObject(0)
	assert.NoError(t, err, "Failed to get object")
	assert.NotNil(t, obj, "Object is nil")

	for _,obj := range archive.Objects() {
		// Check object properties
		t.Logf("Object architecture: %s", obj.Arch())
		t.Logf("Object file format: %s", obj.FileFormat())
		t.Logf("Object kind: %s", obj.Kind())
		t.Logf("Object debug ID: %s", obj.DebugID())

		features := obj.Features()
		t.Logf("Has debug info: %v", features.HasDebug)
		t.Logf("Has symbols: %v", features.HasSymtab)

		// Create a symcache from the object
		symCache, err := NewSymCacheFromObject(&obj)
		assert.NoError(t, err, "Failed to create symcache")


		// Log some symcache info
		t.Logf("SymCache arch: %s", symCache.Arch())
		t.Logf("SymCache debug ID: %s", symCache.DebugID())
		t.Logf("SymCache version: %d", symCache.Version())

		// Try looking up a symbol at a specific address
		// Using 0x10000 as a more likely address to find something in a real binary
		locations, err := symCache.Lookup(0x10000)
		assert.NoError(t, err, "Symbol lookup failed")
		
		// Log found symbols (there might not be any at this specific address)
		t.Logf("Found %d locations at address 0x10000", len(locations))
		if len(locations) > 0 {
			for i, loc := range locations {
				t.Logf("Symbol %d: %s at %s:%d", i, loc.Symbol, loc.FullPath, loc.Line)
			}
		}
	}
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
type CrashReport struct {
	FaultingThread int
	Threads []Thread
	UsedImages []Image
}

func TestSymbolicateWithDSym(t *testing.T) {
	// Test using the specific dSYM file available in the repo
	dsymPath := "crashcrashcrash.app.dSYM"
	
	// Find the actual DWARF file within the dSYM
	dwarfBinaryPath := dsymPath + "/Contents/Resources/DWARF/crashcrashcrash"
	
	// Verify the file exists
	_, err := os.Stat(dwarfBinaryPath)
	assert.NoError(t, err, "DWARF binary file not found")

	archive, err := NewArchiveFromPath(dwarfBinaryPath)
	assert.NoError(t, err, "Failed to load DWARF binary")

	err = archive.BuildSymCaches()
	assert.NoError(t, err, "Failed to build SymCaches")

	// somehow open the crash file
	f, err := os.ReadFile("crashcrashcrash.json")
	assert.NoError(t, err)

	var report CrashReport

	err = json.Unmarshal(f, &report)
	assert.NoError(t, err)

	// thread 0 is the crashing thread
	assert.Equal(t, 0, report.FaultingThread)

	thread := report.Threads[0]

	// frames 1 and 2 need to be symbolicated
	frame1 := thread.Frames[0]
	frame2 := thread.Frames[1]
	assert.Nil(t, frame1.Symbol)
	assert.Nil(t, frame2.Symbol)

	symbolicated, err := symbolicateFrame(frame1, thread, report, *archive, true)
	assert.NoError(t, err)

	// frame 1 symbolicates to 2 frames O.o
	assert.Len(t, symbolicated, 2)

	sframe1 := symbolicated[0]
	assert.Equal(t, "Swift runtime failure: Unexpectedly found nil while unwrapping an Optional value", *sframe1.Symbol)
	assert.Equal(t, uint64(4294967295), *sframe1.SymbolLocation)
	sframe2 := symbolicated[1]
	assert.Equal(t, "$s15crashcrashcrash4loopyyF", *sframe2.Symbol)
	assert.Equal(t, uint64(4084), *sframe2.SymbolLocation)
	
	// frame 2 symbolicates to just 1
	symbolicated, err = symbolicateFrame(frame2, thread, report, *archive, false)
	assert.NoError(t, err)
	assert.Len(t, symbolicated, 1)
	
	sframe1 = symbolicated[0]
	assert.Equal(t, "$s15crashcrashcrash11crashTheAppyyF", *sframe1.Symbol)
	assert.Equal(t, uint64(4072), *sframe1.SymbolLocation)
}

func TestFindBestInstruction(t *testing.T) {
	dsymPath := "crashcrashcrash.app.dSYM"
	dwarfBinaryPath := dsymPath + "/Contents/Resources/DWARF/crashcrashcrash"
	archive, err := NewArchiveFromPath(dwarfBinaryPath)
	assert.NoError(t, err, "Failed to load DWARF binary")

	err = archive.BuildSymCaches()
	assert.NoError(t, err, "Failed to build SymCaches")

	f, err := os.ReadFile("crashcrashcrash.json")
	assert.NoError(t, err)

	var report CrashReport
	err = json.Unmarshal(f, &report)
	assert.NoError(t, err)

	thread := report.Threads[0]
	frame1 := thread.Frames[0]

	// frame 1
	imageOffset := frame1.ImageOffset
	imageIndex := frame1.ImageIndex
	image := report.UsedImages[imageIndex]

	cache := archive.symCaches[image.UUID]

	ipRegName := ArchIPRegName(cache.Arch())
	ipRegState, found := thread.ThreadState[ipRegName]
	var ipRegValue uint64 = 0
	if found {
		ipMap := ipRegState.(map[string]any)
		ipRegValue = uint64(ipMap["value"].(float64))
	}
	addr := FindBestInstruction(imageOffset, cache.Arch(), true, ipRegValue)
	assert.Equal(t, uint64(4196), addr)

	// frame 2
	frame2 := thread.Frames[1]
	imageOffset = frame2.ImageOffset
	imageIndex = frame2.ImageIndex
	image = report.UsedImages[imageIndex]

	cache = archive.symCaches[image.UUID]

	ipRegName = ArchIPRegName(cache.Arch())
	ipRegState, found = thread.ThreadState[ipRegName]
	ipRegValue = uint64(0)
	if found {
		ipMap := ipRegState.(map[string]any)
		ipRegValue = uint64(ipMap["value"].(float64))
	}
	addr = FindBestInstruction(imageOffset, cache.Arch(), true, ipRegValue)
	assert.Equal(t, uint64(4084), addr)
}

func symbolicateFrame(frame Frame, thread Thread, report CrashReport, archive Archive, isCrashingFrame bool) ([]Frame, error) {
	imageOffset := frame.ImageOffset
	imageIndex := frame.ImageIndex
	image := report.UsedImages[imageIndex]

	cache := archive.symCaches[image.UUID]

	ipRegName := ArchIPRegName(cache.Arch())
	ipRegState, found := thread.ThreadState[ipRegName]
	var ipRegValue uint64 = 0
	if found {
		ipMap := ipRegState.(map[string]any)
		ipRegValue = uint64(ipMap["value"].(float64))
	}
	addr := FindBestInstruction(imageOffset, cache.Arch(), isCrashingFrame, ipRegValue)
	
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
