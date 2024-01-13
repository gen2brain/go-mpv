//go:build (!cgo || nocgo) && windows

package mpv

import (
	"fmt"

	"golang.org/x/sys/windows"
)

const (
	libname = "libmpv-2.dll"
)

// loadLibrary loads the dll and panics on error.
func loadLibrary() uintptr {
	handle, err := windows.LoadLibrary(libname)
	if err != nil {
		panic(fmt.Errorf("cannot load library %s: %w", libname, err))
	}

	return uintptr(handle)
}
