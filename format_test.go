package mpv

import (
	"testing"
	"unsafe"
)

// newHeadless returns an initialized mpv instance with no audio/video output.
func newHeadless(t *testing.T) *Mpv {
	t.Helper()

	m := New()
	if err := m.SetPropertyString("vo", "null"); err != nil {
		t.Fatalf("SetPropertyString vo: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("SetOptionString ao: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(m.TerminateDestroy)

	return m
}

func TestPropertyRoundTrip(t *testing.T) {
	m := newHeadless(t)

	t.Run("flag", func(t *testing.T) {
		if err := m.SetProperty("pause", FormatFlag, true); err != nil {
			t.Fatal(err)
		}
		v, err := m.GetProperty("pause", FormatFlag)
		if err != nil {
			t.Fatal(err)
		}
		if v.(bool) != true {
			t.Fatalf("got %v, want true", v)
		}
	})

	t.Run("int64", func(t *testing.T) {
		if err := m.SetProperty("ab-loop-count", FormatInt64, int64(3)); err != nil {
			t.Fatal(err)
		}
		v, err := m.GetProperty("ab-loop-count", FormatInt64)
		if err != nil {
			t.Fatal(err)
		}
		if v.(int64) != 3 {
			t.Fatalf("got %v, want 3", v)
		}
	})

	t.Run("double", func(t *testing.T) {
		if err := m.SetProperty("speed", FormatDouble, 1.5); err != nil {
			t.Fatal(err)
		}
		v, err := m.GetProperty("speed", FormatDouble)
		if err != nil {
			t.Fatal(err)
		}
		if v.(float64) != 1.5 {
			t.Fatalf("got %v, want 1.5", v)
		}
	})

	t.Run("string", func(t *testing.T) {
		if err := m.SetProperty("force-media-title", FormatString, "héllo"); err != nil {
			t.Fatal(err)
		}
		v, err := m.GetProperty("force-media-title", FormatString)
		if err != nil {
			t.Fatal(err)
		}
		if v.(string) != "héllo" {
			t.Fatalf("got %q, want %q", v, "héllo")
		}
	})
}

func TestStringAccessors(t *testing.T) {
	m := New()
	if err := m.SetOption("force-media-title", FormatString, "via-option"); err != nil {
		t.Fatalf("SetOption FormatString: %v", err)
	}
	if err := m.SetPropertyString("vo", "null"); err != nil {
		t.Fatalf("SetPropertyString vo: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("SetOptionString ao: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	t.Cleanup(m.TerminateDestroy)

	if got := m.GetPropertyString("force-media-title"); got != "via-option" {
		t.Fatalf("GetPropertyString after SetOption = %q, want %q", got, "via-option")
	}

	if err := m.SetPropertyString("force-media-title", "via-string"); err != nil {
		t.Fatalf("SetPropertyString: %v", err)
	}
	if got := m.GetPropertyString("force-media-title"); got != "via-string" {
		t.Fatalf("GetPropertyString = %q, want %q", got, "via-string")
	}
	if got := m.GetPropertyOsdString("force-media-title"); got != "via-string" {
		t.Fatalf("GetPropertyOsdString = %q, want %q", got, "via-string")
	}

	if got := m.GetPropertyString("does-not-exist"); got != "" {
		t.Fatalf("GetPropertyString of missing property = %q, want empty", got)
	}
}

func TestToStr(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{"plain", []byte("hello\x00"), "hello"},
		{"stops at nul", []byte("hello\x00world\x00"), "hello"},
		{"preserves surrounding space", []byte("  hi  \x00"), "  hi  "},
		{"preserves newline", []byte("line\n\x00"), "line\n"},
		{"empty", []byte("\x00"), ""},
		{"utf8", []byte("héllo\x00"), "héllo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toStr(unsafe.Pointer(&tc.in[0]))
			if got != tc.want {
				t.Fatalf("toStr = %q, want %q", got, tc.want)
			}
		})
	}

	if got := toStr(nil); got != "" {
		t.Fatalf("toStr(nil) = %q, want empty", got)
	}
}
