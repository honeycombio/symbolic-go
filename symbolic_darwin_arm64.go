//go:build darwin && arm64

package symbolic

// #cgo LDFLAGS: -lsymbolic_cabi
// #cgo LDFLAGS: -L${SRCDIR}/lib/darwin_arm64
// #cgo LDFLAGS: -Wl,-rpath,./
// #cgo LDFLAGS: -Wl,-rpath,./../lib
// #cgo LDFLAGS: -Wl,-rpath ${SRCDIR}/lib/darwin_arm64
import "C"
