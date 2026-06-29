//go:build (!cgo || nocgo) && darwin

package mpv

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// libnames are the libmpv dylib names, tried in order.
var libnames = []string{"libmpv.dylib", "libmpv.2.dylib"}

// loadLibrary loads the dylib and panics if none of the known names can be found.
func loadLibrary() uintptr {
	var err error
	for _, name := range libnames {
		var handle uintptr
		handle, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return handle
		}
	}

	panic(fmt.Errorf("cannot load libmpv (tried %v): %w", libnames, err))
}

// memLibrary returns the handle to resolve malloc/free from. libmpv pulls in libc,
// so its dependency chain exposes them.
func memLibrary() uintptr {
	return libmpv
}
