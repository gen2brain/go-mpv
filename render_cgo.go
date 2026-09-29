//go:build cgo && !nocgo

package mpv

/*
#include <mpv/client.h>
#include <mpv/render.h>
#include <mpv/render_gl.h>
#include <stdlib.h>
#include <stdint.h>

void *goMpvGetProcAddress(void *ctx, const char *name);
void goMpvRenderUpdate(void *ctx);

static int render_create_sw(mpv_render_context **ctx, mpv_handle *mpv, int advanced) {
    mpv_render_param params[] = {
        {MPV_RENDER_PARAM_API_TYPE, (void *)MPV_RENDER_API_TYPE_SW},
        {MPV_RENDER_PARAM_ADVANCED_CONTROL, &advanced},
        {0},
    };
    return mpv_render_context_create(ctx, mpv, params);
}

static int render_sw(mpv_render_context *ctx, int w, int h, const char *format, size_t stride, void *ptr, int block) {
    int size[2] = {w, h};
    mpv_render_param params[] = {
        {MPV_RENDER_PARAM_SW_SIZE, &size[0]},
        {MPV_RENDER_PARAM_SW_FORMAT, (void *)format},
        {MPV_RENDER_PARAM_SW_STRIDE, &stride},
        {MPV_RENDER_PARAM_SW_POINTER, ptr},
        {MPV_RENDER_PARAM_BLOCK_FOR_TARGET_TIME, &block},
        {0},
    };
    return mpv_render_context_render(ctx, params);
}

static int render_create_gl(mpv_render_context **ctx, mpv_handle *mpv, uintptr_t cb_ctx, int advanced) {
    mpv_opengl_init_params gl = {
        .get_proc_address = goMpvGetProcAddress,
        .get_proc_address_ctx = (void *)cb_ctx,
    };
    mpv_render_param params[] = {
        {MPV_RENDER_PARAM_API_TYPE, (void *)MPV_RENDER_API_TYPE_OPENGL},
        {MPV_RENDER_PARAM_OPENGL_INIT_PARAMS, &gl},
        {MPV_RENDER_PARAM_ADVANCED_CONTROL, &advanced},
        {0},
    };
    return mpv_render_context_create(ctx, mpv, params);
}

static int render_gl(mpv_render_context *ctx, int fbo, int w, int h, int flip_y, int block) {
    mpv_opengl_fbo gl_fbo = {.fbo = fbo, .w = w, .h = h};
    mpv_render_param params[] = {
        {MPV_RENDER_PARAM_OPENGL_FBO, &gl_fbo},
        {MPV_RENDER_PARAM_FLIP_Y, &flip_y},
        {MPV_RENDER_PARAM_BLOCK_FOR_TARGET_TIME, &block},
        {0},
    };
    return mpv_render_context_render(ctx, params);
}

static void render_set_update_callback(mpv_render_context *ctx, uintptr_t cb_ctx) {
    mpv_render_context_set_update_callback(ctx, goMpvRenderUpdate, (void *)cb_ctx);
}
*/
import "C"

import (
	"unsafe"
)

// RenderContext represents an mpv render context (the main video output).
type RenderContext struct {
	ctx  *C.mpv_render_context
	cbID uintptr
}

// NewRenderContextSW creates a software (CPU) render context. The mpv instance
// must have the "vo" option set to "libmpv".
func (m *Mpv) NewRenderContextSW(opts ...RenderContextOption) (*RenderContext, error) {
	o := newRenderContextOptions(opts)
	id := registerRenderCallbacks()

	var ctx *C.mpv_render_context
	err := newError(int(C.render_create_sw(&ctx, m.handle, C.int(cFlag(o.advancedControl)))))
	if err != nil {
		unregisterRenderCallbacks(id)
		return nil, err
	}

	return &RenderContext{ctx: ctx, cbID: id}, nil
}

// NewRenderContextGL creates an OpenGL render context; getProcAddress resolves GL
// functions. Requires vo=libmpv and the GL context current on the calling thread.
func (m *Mpv) NewRenderContextGL(getProcAddress func(name string) unsafe.Pointer, opts ...RenderContextOption) (*RenderContext, error) {
	o := newRenderContextOptions(opts)
	id := registerRenderCallbacks()
	setRenderProcAddress(id, getProcAddress)

	var ctx *C.mpv_render_context
	err := newError(int(C.render_create_gl(&ctx, m.handle, C.uintptr_t(id), C.int(cFlag(o.advancedControl)))))
	if err != nil {
		unregisterRenderCallbacks(id)
		return nil, err
	}

	return &RenderContext{ctx: ctx, cbID: id}, nil
}

// RenderSW renders the current frame into buf, which must hold stride*height
// bytes. format is one of "rgb0", "bgr0", "0bgr", "0rgb".
func (rc *RenderContext) RenderSW(width, height, stride int, format string, buf []byte, opts ...RenderOption) error {
	o := newRenderOptions(opts)
	cformat := C.CString(format)
	defer C.free(unsafe.Pointer(cformat))

	return newError(int(C.render_sw(rc.ctx, C.int(width), C.int(height), cformat, C.size_t(stride), unsafe.Pointer(&buf[0]), C.int(cFlag(o.blockForTargetTime)))))
}

// RenderGL renders the current frame into the given OpenGL framebuffer object
// (0 for the default framebuffer). Set flipY for bottom-up coordinate systems.
func (rc *RenderContext) RenderGL(fbo, width, height int, flipY bool, opts ...RenderOption) error {
	o := newRenderOptions(opts)

	return newError(int(C.render_gl(rc.ctx, C.int(fbo), C.int(width), C.int(height), C.int(cFlag(flipY)), C.int(cFlag(o.blockForTargetTime)))))
}

// SetUpdateCallback sets fn to run when a new frame is ready. fn runs on an mpv
// thread and must only signal a redraw, not call mpv or render directly.
func (rc *RenderContext) SetUpdateCallback(fn func()) {
	setRenderUpdate(rc.cbID, fn)
	C.render_set_update_callback(rc.ctx, C.uintptr_t(rc.cbID))
}

// Update returns the render update flags; check the result against RenderUpdateFrame.
func (rc *RenderContext) Update() uint64 {
	return uint64(C.mpv_render_context_update(rc.ctx))
}

// ReportSwap tells mpv that a frame swap (vsync) occurred.
func (rc *RenderContext) ReportSwap() {
	C.mpv_render_context_report_swap(rc.ctx)
}

// Free destroys the render context. Call it before the mpv core is destroyed.
func (rc *RenderContext) Free() {
	C.mpv_render_context_free(rc.ctx)
	unregisterRenderCallbacks(rc.cbID)
}
