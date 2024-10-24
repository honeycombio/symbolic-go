//go:build linux && arm64

package symbolic

// #cgo LDFLAGS: -lsymbolic_cabi
// #cgo LDFLAGS: -L${SRCDIR}/lib/linux_arm64
// #cgo LDFLAGS: -Wl,-rpath,$ORIGIN
// #cgo LDFLAGS: -Wl,-rpath,$ORIGIN/../lib
// #cgo LDFLAGS: -Wl,-rpath ${SRCDIR}/lib/linux_arm64
import "C"
