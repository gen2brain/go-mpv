package mpv

import (
	"reflect"
	"testing"
)

func TestHook(t *testing.T) {
	m := newHeadless(t)

	if err := m.HookAdd(100, "on_load", 0); err != nil {
		t.Fatalf("HookAdd: %v", err)
	}
	if err := m.Command([]string{"loadfile", "testdata/test.mpg"}); err != nil {
		t.Fatalf("loadfile: %v", err)
	}

	gotHook := false
	for i := 0; i < 200; i++ {
		e := m.WaitEvent(10)
		switch e.EventID {
		case EventHook:
			h := e.Hook()
			if h.Name != "on_load" {
				t.Errorf("hook name = %q, want on_load", h.Name)
			}
			gotHook = true
			if err := m.HookContinue(h.ID); err != nil {
				t.Errorf("HookContinue: %v", err)
			}
		case EventFileLoaded:
			if !gotHook {
				t.Fatal("file loaded before the hook fired")
			}
			return
		case EventNone, EventShutdown:
			t.Fatal("no hook/file-loaded event")
		}
	}
	t.Fatal("timed out waiting for hook and file load")
}

func TestClientMessage(t *testing.T) {
	m := newHeadless(t)

	if err := m.CommandString("script-message hello world"); err != nil {
		t.Fatalf("script-message: %v", err)
	}

	for i := 0; i < 100; i++ {
		e := m.WaitEvent(10)
		if e.EventID == EventClientMessage {
			if got := e.ClientMessage(); !reflect.DeepEqual(got, []string{"hello", "world"}) {
				t.Fatalf("ClientMessage = %#v, want [hello world]", got)
			}
			return
		}
		if e.EventID == EventNone {
			break
		}
	}
	t.Fatal("no client-message event")
}
