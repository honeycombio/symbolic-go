package symbolic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProguard(t *testing.T) {
	pm, err := NewProguardMapper("./proguard.txt")
	assert.NoError(t, err)

	assert.True(t, pm.HasLineInfo)
	assert.Equal(t, "a48ca62b-df26-544e-a8b9-2a5ce210d1d5", pm.UUID.String())
}
