package symbolic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProguard(t *testing.T) {
	pm, err := NewProguardMapper("./proguard.txt")
	assert.NoError(t, err)

	assert.True(t, pm.HasLineInfo)
	assert.Equal(t, "a48ca62b-df26-544e-a8b9-2a5ce210d1d5", pm.UUID)

	class, err := pm.RemapClass("android.support.constraint.ConstraintLayout$a")
	assert.NoError(t, err)
	assert.Equal(t, "android.support.constraint.ConstraintLayout$LayoutParams", class)

	frames, err := pm.RemapMethod("android.support.constraint.a.b", "f")
	assert.NoError(t, err)
	assert.Len(t, frames, 1)
	assert.Equal(t, "android.support.constraint.solver.ArrayRow", frames[0].ClassName)
	assert.Equal(t, "pickRowVariable", frames[0].MethodName)

	frames, err = pm.RemapFrame("android.support.constraint.a.b", "a", 116)
	assert.NoError(t, err)
	assert.Len(t, frames, 1)

	assert.Equal(t, "android.support.constraint.solver.ArrayRow", frames[0].ClassName)
	assert.Equal(t, "createRowDefinition", frames[0].MethodName)
	assert.Equal(t, 116, frames[0].LineNumber)

	frames, err = pm.RemapFrame("io.sentry.sample.MainActivity", "a", 1)
	assert.NoError(t, err)
	assert.Len(t, frames, 3)

	assert.Equal(t, "bar", frames[0].MethodName)
	assert.Equal(t, 54, frames[0].LineNumber)
	assert.Equal(t, "foo", frames[1].MethodName)
	assert.Equal(t, 44, frames[1].LineNumber)
	assert.Equal(t, "onClickHandler", frames[2].MethodName)
	assert.Equal(t, 40, frames[2].LineNumber)
}
