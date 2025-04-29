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

	class, err := pm.RemapClass("android.support.constraint.ConstraintLayout$a")
	assert.NoError(t, err)
	assert.Equal(t, "android.support.constraint.ConstraintLayout$LayoutParams", class)

	frames, err := pm.RemapMethod("android.support.constraint.a.b", "f")
	assert.NoError(t, err)
	assert.Len(t, frames, 1)
	assert.Equal(t, "android.support.constraint.solver.ArrayRow", frames[0].ClassName)
	assert.Equal(t, "pickRowVariable", frames[0].MethodName)
}
