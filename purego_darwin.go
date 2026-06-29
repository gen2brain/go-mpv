//go:build (!cgo || nocgo) && darwin

package mpv

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// libnames are the libmpv dylib names, tried in order.
var libnames = []string{"libmpv.dylib", "libmpv.2.dylib"}

// Locale categories for SetLocale (BSD/macOS values).
const (
	LCAll     = 0
	LCNumeric = 4
)

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

// memLibrary returns the handle for malloc/free; libmpv's libc dependency exposes them.
func memLibrary() uintptr {
	return libmpv
}
