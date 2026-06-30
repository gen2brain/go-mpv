package mpv

import (
	"sync"
	"unsafe"
)

// Render API type strings for MPV_RENDER_PARAM_API_TYPE.
const (
	RenderAPITypeOpenGL = "opengl"
	RenderAPITypeSW     = "sw"
)

// RenderUpdateFrame is set in the result of RenderContext.Update when a new
// video frame should be rendered.
const RenderUpdateFrame uint64 = 1 << 0

// Render callbacks live in a token-keyed registry; the token is passed to C as the
// callback context, so one trampoline dispatches without handing Go pointers to C.
var (
	renderCbMu  sync.Mutex
	renderCbs   = map[uintptr]*renderCallbacks{}
	renderCbSeq uintptr
)

type renderCallbacks struct {
	getProcAddress func(name string) unsafe.Pointer
	update         func()
}

func registerRenderCallbacks() uintptr {
	renderCbMu.Lock()
	defer renderCbMu.Unlock()

	renderCbSeq++
	renderCbs[renderCbSeq] = &renderCallbacks{}

	return renderCbSeq
}

func unregisterRenderCallbacks(id uintptr) {
	renderCbMu.Lock()
	defer renderCbMu.Unlock()

	delete(renderCbs, id)
}

func setRenderProcAddress(id uintptr, fn func(name string) unsafe.Pointer) {
	renderCbMu.Lock()
	defer renderCbMu.Unlock()

	if cb := renderCbs[id]; cb != nil {
		cb.getProcAddress = fn
	}
}

func setRenderUpdate(id uintptr, fn func()) {
	renderCbMu.Lock()
	defer renderCbMu.Unlock()

	if cb := renderCbs[id]; cb != nil {
		cb.update = fn
	}
}

// dispatchProcAddress and dispatchUpdate are called by the per-backend C
// trampolines with the token previously passed to C.
func dispatchProcAddress(id uintptr, name string) unsafe.Pointer {
	renderCbMu.Lock()
	var fn func(name string) unsafe.Pointer
	if cb := renderCbs[id]; cb != nil {
		fn = cb.getProcAddress
	}
	renderCbMu.Unlock()

	if fn == nil {
		return nil
	}

	return fn(name)
}

func dispatchUpdate(id uintptr) {
	renderCbMu.Lock()
	var fn func()
	if cb := renderCbs[id]; cb != nil {
		fn = cb.update
	}
	renderCbMu.Unlock()

	if fn != nil {
		fn()
	}
}
