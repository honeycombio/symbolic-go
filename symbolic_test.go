package symbolic

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// the c-abi uses 1-based line and col numbers
// the rust sourcemapcache uses 0-based line and col numbers
// when doing a lookup add 1 to the line and col as per https://github.com/getsentry/symbolic/blob/master/symbolic-cabi/src/sourcemapcache.rs#L167
// when comparing the result line and col add 1 as per https://github.com/getsentry/symbolic/blob/master/symbolic-cabi/src/sourcemapcache.rs#L144-L145

func TestMin(t *testing.T) {
	_, err := NewSourceMapCache("{}", "{}")
	assert.NoError(t, err)
}

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

type TestCase struct {
	Name             string
	Description      string
	BaseFile         string
	SourceMapFile    string
	SourceMapIsValid bool
	TestActions      []struct {
		ActionType       string
		GeneratedLine    int
		GeneratedColumn  int
		OriginalSource   string
		OriginalLine     int
		OriginalColumn   int
		MappedName       string
		IntermediateMaps []string
	}
}

func TestSourceMaps(t *testing.T) {
	f, err := os.ReadFile("source-map-tests/source-map-spec-tests.json")
	assert.NoError(t, err)

	var specs map[string][]*TestCase

	err = json.Unmarshal(f, &specs)
	assert.NoError(t, err)

	for _, cases := range specs {
		for _, tc := range cases {
			if tc.Name == "validMappingLargeVLQ" {
				// skip this test as it is not supported by the current implementation
				continue
			}

			t.Run(tc.Name, func(t *testing.T) {
				base, err := os.ReadFile("source-map-tests/resources/" + tc.BaseFile)
				assert.NoError(t, err)
				sourceMap, err := os.ReadFile("source-map-tests/resources/" + tc.SourceMapFile)
				assert.NoError(t, err)

				smc, err := NewSourceMapCache(string(base), string(sourceMap))

				if tc.SourceMapIsValid {
					assert.NoError(t, err)
				}

				for i, action := range tc.TestActions {
					if tc.Name == "vlqValidNegativeDigit" && i == 0 {
						// skip this test as it is not supported by the current implementation
						continue
					}

					if tc.Name == "mappingSemanticsSingleFieldSegment" && i == 1 {
						// skip this test as it is not supported by the current implementation
						continue
					}

					t.Run(fmt.Sprintf("%s%d", action.ActionType, i), func(t *testing.T) {
						switch action.ActionType {
						case "checkMapping":
							token, err := smc.Lookup(uint32(action.GeneratedLine+1), uint32(action.GeneratedColumn+1), 0)
							assert.NoError(t, err)
							assert.Equal(t, action.OriginalColumn, token.Col-1)
							assert.Equal(t, action.OriginalLine, token.Line-1)
							assert.Equal(t, action.OriginalSource, token.Src)
							assert.Equal(t, action.MappedName, token.Name)
						case "checkMappingTransitive":
							token, err := smc.Lookup(uint32(action.GeneratedLine+1), uint32(action.GeneratedColumn+1), 0)
							assert.NoError(t, err)
							for _, m := range action.IntermediateMaps {
								f, err = os.ReadFile("source-map-tests/resources/" + m)
								assert.NoError(t, err)
								ismc, err := NewSourceMapCache(string(base), string(f))
								assert.NoError(t, err)
								token, err = ismc.Lookup(uint32(token.Line), uint32(token.Col), 0)
								assert.NoError(t, err)
							}

							assert.Equal(t, action.OriginalColumn, token.Col-1)
							assert.Equal(t, action.OriginalLine, token.Line-1)
							assert.Equal(t, action.OriginalSource, token.Src)
							assert.Equal(t, action.MappedName, token.Name)
						}
					})
				}
			})
		}
	}
}
