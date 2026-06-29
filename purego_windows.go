//go:build (!cgo || nocgo) && windows

package mpv

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// libnames are the libmpv DLL names used by various distributions, tried in order.
var libnames = []string{"libmpv-2.dll", "libmpv.dll", "mpv-2.dll", "mpv-1.dll"}

// loadLibrary loads the dll and panics if none of the known names can be found.
func loadLibrary() uintptr {
	var err error
	for _, name := range libnames {
		var handle windows.Handle
		handle, err = windows.LoadLibrary(name)
		if err == nil {
			return uintptr(handle)
		}
	}

	panic(fmt.Errorf("cannot load libmpv (tried %v): %w", libnames, err))
}

// crtnames are the C runtime DLLs that export malloc/free, tried in order.
var crtnames = []string{"ucrtbase.dll", "msvcrt.dll"}

// memLibrary loads the C runtime to resolve malloc/free, which libmpv does not export.
func memLibrary() uintptr {
	var err error
	for _, name := range crtnames {
		var handle windows.Handle
		handle, err = windows.LoadLibrary(name)
		if err == nil {
			return uintptr(handle)
		}
	}

	panic(fmt.Errorf("cannot load C runtime (tried %v): %w", crtnames, err))
}
