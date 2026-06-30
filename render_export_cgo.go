//go:build cgo && !nocgo

package mpv

/*
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

//export goMpvGetProcAddress
func goMpvGetProcAddress(ctx unsafe.Pointer, name *C.char) unsafe.Pointer {
	return dispatchProcAddress(uintptr(ctx), C.GoString(name))
}

//export goMpvRenderUpdate
func goMpvRenderUpdate(ctx unsafe.Pointer) {
	dispatchUpdate(uintptr(ctx))
}
