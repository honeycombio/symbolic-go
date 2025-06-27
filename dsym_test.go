package symbolic

import (
	"os"
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
	count, err := archive.objectCount()
	assert.NoError(t, err)
	assert.Equal(t, count, 1, "Expected at least one object in the archive")

	// Get the first object
	obj, err := archive.getObject(0)
	assert.NoError(t, err, "Failed to get object")
	assert.NotNil(t, obj, "Object is nil")

	objects, err := archive.objects()
	assert.NoError(t, err)
	for _,obj := range objects {
		assert.Equal(t, obj.arch, "x86_64")
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
	count,err := archive.objectCount()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1, "Expected at least one object in the archive")

	// Get the first object
	obj, err := archive.getObject(0)
	assert.NoError(t, err, "Failed to get object")
	assert.NotNil(t, obj, "Object is nil")

	objects, err := archive.objects()
	assert.NoError(t, err)
	for _,obj := range objects {
		assert.NotEmpty(t, obj.arch)
		t.Logf("Object architecture: %s", obj.arch)
		assert.NotEmpty(t, obj.fileFormat)
		t.Logf("Object file format: %s", obj.fileFormat)
		assert.NotEmpty(t, obj.kind)
		t.Logf("Object kind: %s", obj.kind)
		assert.NotEmpty(t, obj.debugId)
		t.Logf("Object debug ID: %s", obj.debugId)

		features := obj.features
		assert.NotNil(t, obj.features)
		t.Logf("Has debug info: %v", features.HasDebug)
		t.Logf("Has symbols: %v", features.HasSymtab)

		// Create a symcache from the object
		symCache, err := NewSymCacheFromObject(&obj)
		assert.NoError(t, err, "Failed to create symcache")

		t.Logf("SymCache arch: %s", symCache.arch)
		t.Logf("SymCache debug ID: %s", symCache.debugId)


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