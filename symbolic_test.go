package symbolic

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// the c-abi uses 1-based line and col numbers
// the rust sourcemapcache uses 0-based line and col numbers
// when doing a lookup add 1 to the line and col as per https://github.com/getsentry/symbolic/blob/master/symbolic-cabi/src/sourcemapcache.rs#L167
// when comparing the result line and col add 1 as per https://github.com/getsentry/symbolic/blob/master/symbolic-cabi/src/sourcemapcache.rs#L144-L145

func TestResolvesInlineFunction(t *testing.T) {
	minfied, err := os.ReadFile("symbolic/symbolic-testutils/fixtures/sourcemapcache/inlining/module.js")
	assert.NoError(t, err)
	sourceMap, err := os.ReadFile("symbolic/symbolic-testutils/fixtures/sourcemapcache/inlining/module.js.map")
	assert.NoError(t, err)

	smc, err := NewSourceMapCache(string(minfied), string(sourceMap))
	assert.NoError(t, err)

	token, err := smc.Lookup(1, 63, 0)
	assert.NoError(t, err)
	assert.Equal(t, "../src/app.js", token.Src)
	assert.Equal(t, 3, token.Line)
	assert.Equal(t, 30, token.Col)
	assert.Equal(t, "buttonCallback", token.FunctionName)

	token, err = smc.Lookup(1, 47, 0)
	assert.NoError(t, err)
	assert.Equal(t, "../src/bar.js", token.Src)
	assert.Equal(t, 4, token.Line)
	assert.Equal(t, 3, token.Col)
	assert.Equal(t, "bar", token.FunctionName)
	assert.Equal(t, "foo", token.Name)

	token, err = smc.Lookup(1, 34, 0)
	assert.NoError(t, err)
	assert.Equal(t, "../src/foo.js", token.Src)
	assert.Equal(t, 2, token.Line)
	assert.Equal(t, 9, token.Col)
	assert.Equal(t, "<anonymous>", token.FunctionName)
}

func TestWritesSimpleCache(t *testing.T) {
	minfied, err := os.ReadFile("symbolic/symbolic-testutils/fixtures/sourcemapcache/simple/minified.js")
	assert.NoError(t, err)
	sourceMap, err := os.ReadFile("symbolic/symbolic-testutils/fixtures/sourcemapcache/simple/minified.js.map")
	assert.NoError(t, err)

	smc, err := NewSourceMapCache(string(minfied), string(sourceMap))
	assert.NoError(t, err)

	token, err := smc.Lookup(1, 11, 0)
	assert.NoError(t, err)
	assert.Equal(t, "tests/fixtures/simple/original.js", token.Src)
	assert.Equal(t, 2, token.Line)
	assert.Equal(t, 10, token.Col)
	assert.Equal(t, "function abcd() {}\n", token.ContextLine)
}
