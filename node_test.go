package mpv

import (
	"reflect"
	"testing"
)

func TestNodeRoundTrip(t *testing.T) {
	cases := []any{
		nil,
		"hello",
		"héllo",
		true,
		false,
		int64(42),
		int64(-7),
		3.14,
		[]byte{1, 2, 3, 0, 255},
		[]byte{},
		[]any{},
		[]any{"a", int64(1), true, 2.5},
		map[string]any{},
		map[string]any{"k1": "v1", "k2": int64(7)},
		[]any{[]any{int64(1), int64(2)}, map[string]any{"nested": "yes"}},
	}

	for _, c := range cases {
		p, cleanup := goToNode(c)
		got := nodeToGo(p)
		cleanup()

		if !reflect.DeepEqual(got, c) {
			t.Errorf("round-trip of %#v = %#v", c, got)
		}
	}
}

func TestNodeTrackList(t *testing.T) {
	m := newHeadless(t)

	if err := m.Command([]string{"loadfile", "testdata/test.mpg"}); err != nil {
		t.Fatalf("loadfile: %v", err)
	}

	for {
		e := m.WaitEvent(10)
		if e.EventID == EventFileLoaded {
			break
		}
		if e.EventID == EventShutdown || e.EventID == EventNone {
			t.Fatal("file did not load")
		}
		if e.EventID == EventEnd {
			t.Fatalf("playback ended before load: %v", e.EndFile().Reason)
		}
	}

	v, err := m.GetProperty("track-list", FormatNode)
	if err != nil {
		t.Fatalf("GetProperty track-list NODE: %v", err)
	}

	tracks, ok := v.([]any)
	if !ok {
		t.Fatalf("track-list = %T, want []any", v)
	}
	if len(tracks) == 0 {
		t.Fatal("track-list is empty")
	}

	first, ok := tracks[0].(map[string]any)
	if !ok {
		t.Fatalf("track[0] = %T, want map[string]any", tracks[0])
	}
	if _, ok := first["type"]; !ok {
		t.Fatalf("track[0] missing \"type\" key: %#v", first)
	}
}
