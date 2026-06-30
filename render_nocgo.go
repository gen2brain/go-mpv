//go:build !cgo || nocgo

package mpv

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// mpv_render_param_type values.
const (
	renderParamAPIType          = 1
	renderParamOpenGLInitParams = 2
	renderParamOpenGLFBO        = 3
	renderParamFlipY            = 4
	renderParamSWSize           = 17
	renderParamSWFormat         = 18
	renderParamSWStride         = 19
	renderParamSWPointer        = 20
)

// cRenderParam mirrors C mpv_render_param.
type cRenderParam struct {
	typ  int32
	data unsafe.Pointer
}

// cOpenGLInitParams mirrors C mpv_opengl_init_params.
type cOpenGLInitParams struct {
	getProcAddress uintptr
	ctx            uintptr
}

// cOpenGLFBO mirrors C mpv_opengl_fbo.
type cOpenGLFBO struct {
	fbo, w, h, internalFormat int32
}

var renderContextCreate func(ctx unsafe.Pointer, mpv uintptr, params unsafe.Pointer) int
var renderContextRender func(ctx uintptr, params unsafe.Pointer) int
var renderContextUpdate func(ctx uintptr) uint64
var renderContextReportSwap func(ctx uintptr)
var renderContextSetUpdateCallback func(ctx, cb, cbCtx uintptr)
var renderContextFree func(ctx uintptr)

func init() {
	purego.RegisterLibFunc(&renderContextCreate, libmpv, "mpv_render_context_create")
	purego.RegisterLibFunc(&renderContextRender, libmpv, "mpv_render_context_render")
	purego.RegisterLibFunc(&renderContextUpdate, libmpv, "mpv_render_context_update")
	purego.RegisterLibFunc(&renderContextReportSwap, libmpv, "mpv_render_context_report_swap")
	purego.RegisterLibFunc(&renderContextSetUpdateCallback, libmpv, "mpv_render_context_set_update_callback")
	purego.RegisterLibFunc(&renderContextFree, libmpv, "mpv_render_context_free")
}

// Created once; a single trampoline per callback dispatches by the token.
var (
	renderCbOnce   sync.Once
	procAddrCb     uintptr
	renderUpdateCb uintptr
)

func ensureRenderCallbacks() {
	renderCbOnce.Do(func() {
		procAddrCb = purego.NewCallback(func(ctx unsafe.Pointer, name *byte) uintptr {
			return uintptr(dispatchProcAddress(uintptr(ctx), toStr(unsafe.Pointer(name))))
		})
		renderUpdateCb = purego.NewCallback(func(ctx unsafe.Pointer) uintptr {
			dispatchUpdate(uintptr(ctx))
			return 0
		})
	})
}

// RenderContext represents an mpv render context (the main video output).
type RenderContext struct {
	ctx  uintptr
	cbID uintptr
}

// NewRenderContextSW creates a software (CPU) render context. The mpv instance
// must have the "vo" option set to "libmpv".
func (m *Mpv) NewRenderContextSW() (*RenderContext, error) {
	id := registerRenderCallbacks()

	apiType := cStr(RenderAPITypeSW)
	params := []cRenderParam{
		{typ: renderParamAPIType, data: unsafe.Pointer(apiType)},
		{},
	}

	var ctx uintptr
	err := newError(renderContextCreate(unsafe.Pointer(&ctx), m.handle, unsafe.Pointer(&params[0])))
	if err != nil {
		unregisterRenderCallbacks(id)
		return nil, err
	}

	return &RenderContext{ctx: ctx, cbID: id}, nil
}

// NewRenderContextGL creates an OpenGL render context; getProcAddress resolves GL
// functions. Requires vo=libmpv and the GL context current on the calling thread.
func (m *Mpv) NewRenderContextGL(getProcAddress func(name string) unsafe.Pointer) (*RenderContext, error) {
	ensureRenderCallbacks()
	id := registerRenderCallbacks()
	setRenderProcAddress(id, getProcAddress)

	apiType := cStr(RenderAPITypeOpenGL)
	gl := cOpenGLInitParams{getProcAddress: procAddrCb, ctx: id}
	params := []cRenderParam{
		{typ: renderParamAPIType, data: unsafe.Pointer(apiType)},
		{typ: renderParamOpenGLInitParams, data: unsafe.Pointer(&gl)},
		{},
	}

	var ctx uintptr
	err := newError(renderContextCreate(unsafe.Pointer(&ctx), m.handle, unsafe.Pointer(&params[0])))
	if err != nil {
		unregisterRenderCallbacks(id)
		return nil, err
	}

	return &RenderContext{ctx: ctx, cbID: id}, nil
}

// RenderSW renders the current frame into buf, which must hold stride*height
// bytes. format is one of "rgb0", "bgr0", "0bgr", "0rgb".
func (rc *RenderContext) RenderSW(width, height, stride int, format string, buf []byte) error {
	size := [2]int32{int32(width), int32(height)}
	cformat := cStr(format)
	cstride := uintptr(stride)
	params := []cRenderParam{
		{typ: renderParamSWSize, data: unsafe.Pointer(&size[0])},
		{typ: renderParamSWFormat, data: unsafe.Pointer(cformat)},
		{typ: renderParamSWStride, data: unsafe.Pointer(&cstride)},
		{typ: renderParamSWPointer, data: unsafe.Pointer(&buf[0])},
		{},
	}

	return newError(renderContextRender(rc.ctx, unsafe.Pointer(&params[0])))
}

// RenderGL renders the current frame into the given OpenGL framebuffer object
// (0 for the default framebuffer). Set flipY for bottom-up coordinate systems.
func (rc *RenderContext) RenderGL(fbo, width, height int, flipY bool) error {
	gl := cOpenGLFBO{fbo: int32(fbo), w: int32(width), h: int32(height)}
	flip := int32(0)
	if flipY {
		flip = 1
	}
	params := []cRenderParam{
		{typ: renderParamOpenGLFBO, data: unsafe.Pointer(&gl)},
		{typ: renderParamFlipY, data: unsafe.Pointer(&flip)},
		{},
	}

	return newError(renderContextRender(rc.ctx, unsafe.Pointer(&params[0])))
}

// SetUpdateCallback sets fn to run when a new frame is ready. fn runs on an mpv
// thread and must only signal a redraw, not call mpv or render directly.
func (rc *RenderContext) SetUpdateCallback(fn func()) {
	ensureRenderCallbacks()
	setRenderUpdate(rc.cbID, fn)
	renderContextSetUpdateCallback(rc.ctx, renderUpdateCb, rc.cbID)
}

// Update returns the render update flags; check the result against RenderUpdateFrame.
func (rc *RenderContext) Update() uint64 {
	return renderContextUpdate(rc.ctx)
}

// ReportSwap tells mpv that a frame swap (vsync) occurred.
func (rc *RenderContext) ReportSwap() {
	renderContextReportSwap(rc.ctx)
}

// Free destroys the render context. Call it before the mpv core is destroyed.
func (rc *RenderContext) Free() {
	renderContextFree(rc.ctx)
	unregisterRenderCallbacks(rc.cbID)
}
