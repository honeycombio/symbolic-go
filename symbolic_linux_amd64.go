//go:build linux && amd64

package symbolic

// #cgo LDFLAGS: -lsymbolic_cabi
// #cgo LDFLAGS: -L${SRCDIR}/lib/linux_amd64
// #cgo LDFLAGS: -Wl,-rpath,$ORIGIN
// #cgo LDFLAGS: -Wl,-rpath,$ORIGIN/../lib
// #cgo LDFLAGS: -Wl,-rpath ${SRCDIR}/lib/linux_amd64
import "C"
