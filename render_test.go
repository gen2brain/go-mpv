package mpv

import (
	"sync/atomic"
	"testing"
)

func TestRenderSW(t *testing.T) {
	m := New()
	defer m.TerminateDestroy()

	if err := m.SetOptionString("vo", "libmpv"); err != nil {
		t.Fatalf("set vo=libmpv: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("set ao=null: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	rc, err := m.NewRenderContextSW()
	if err != nil {
		t.Fatalf("NewRenderContextSW: %v", err)
	}
	defer rc.Free()

	if err := m.Command([]string{"loadfile", "testdata/test.mpg"}); err != nil {
		t.Fatalf("loadfile: %v", err)
	}

	for {
		e := m.WaitEvent(10)
		if e.EventID == EventFileLoaded {
			break
		}
		if e.EventID == EventNone || e.EventID == EventShutdown {
			t.Fatal("file did not load")
		}
		if e.EventID == EventEnd {
			t.Fatalf("playback ended before load: %v", e.EndFile().Reason)
		}
	}

	const w, h, stride = 320, 240, 320 * 4
	buf := make([]byte, stride*h)

	// Render frames until one has non-black content (the first frames may be black).
	for i := 0; i < 200; i++ {
		if rc.Update()&RenderUpdateFrame != 0 {
			if err := rc.RenderSW(w, h, stride, "rgb0", buf); err != nil {
				t.Fatalf("RenderSW: %v", err)
			}
			for _, b := range buf {
				if b != 0 {
					return
				}
			}
		}
		m.WaitEvent(0.05)
	}

	t.Fatal("no non-black frame rendered")
}

// TestRenderUpdateCallback exercises the callback machinery (registry, C
// trampoline, dispatch) without needing an OpenGL context.
func TestRenderUpdateCallback(t *testing.T) {
	m := New()
	defer m.TerminateDestroy()

	if err := m.SetOptionString("vo", "libmpv"); err != nil {
		t.Fatalf("set vo=libmpv: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("set ao=null: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	rc, err := m.NewRenderContextSW()
	if err != nil {
		t.Fatalf("NewRenderContextSW: %v", err)
	}
	defer rc.Free()

	var called atomic.Int32
	rc.SetUpdateCallback(func() { called.Store(1) })

	if err := m.Command([]string{"loadfile", "testdata/test.mpg"}); err != nil {
		t.Fatalf("loadfile: %v", err)
	}

	buf := make([]byte, 320*240*4)
	for i := 0; i < 200; i++ {
		if called.Load() == 1 {
			return
		}
		if rc.Update()&RenderUpdateFrame != 0 {
			_ = rc.RenderSW(320, 240, 320*4, "rgb0", buf)
		}
		m.WaitEvent(0.05)
	}

	t.Fatal("update callback was not called")
}
