package symbolic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeStr(t *testing.T) {
	original := "hello"

	encoded := encodeStr(original)

	decoded := decodeStr(encoded)

	assert.Equal(t, original, decoded)
}
