//go:build (!cgo || nocgo) && unix && !darwin

package mpv

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// libnames are the libmpv shared object names, tried in order.
var libnames = []string{"libmpv.so", "libmpv.so.2"}

// Locale categories for SetLocale (glibc/musl values).
const (
	LCAll     = 6
	LCNumeric = 1
)

// loadLibrary loads the so and panics if none of the known names can be found.
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
