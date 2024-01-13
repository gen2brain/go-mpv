package mpv_test

import (
	"fmt"
	"testing"

	"github.com/gen2brain/go-mpv"
)

func TestMPV(t *testing.T) {
	m := mpv.New()
	defer m.TerminateDestroy()

	err := m.RequestLogMessages("v")
	if err != nil {
		t.Errorf("RequestLogMessages: %v", err)
	}

	err = m.ObserveProperty(0, "pause", mpv.FormatFlag)
	if err != nil {
		t.Errorf("ObserveProperty: %v", err)
	}

	err = m.SetPropertyString("vo", "null")
	if err != nil {
		t.Errorf("SetPropertyString: %v", err)
	}

	err = m.SetOptionString("ao", "null")
	if err != nil {
		t.Errorf("SetOptionString: %v", err)
	}

	err = m.SetOption("cache", mpv.FormatFlag, true)
	if err != nil {
		t.Errorf("SetOption: %v", err)
	}

	err = m.Initialize()
	if err != nil {
		t.Errorf("Initialize: %v", err)
	}

	err = m.Command([]string{"loadfile", "testdata/test.mpg"})
	if err != nil {
		t.Errorf("Command: %v", err)
	}

loop:
	for {
		e := m.WaitEvent(10000)

		switch e.EventID {
		case mpv.EventPropertyChange:
			prop := e.Property()
			value := prop.Data.(int)
			fmt.Println("property:", prop.Name, value)
		case mpv.EventFileLoaded:
			p, err := m.GetProperty("media-title", mpv.FormatString)
			if err != nil {
				t.Errorf("GetProperty: %v", err)
			}
			fmt.Println("title:", p.(string))
		case mpv.EventLogMsg:
			msg := e.LogMessage()
			fmt.Println("message:", msg.Text)
		case mpv.EventStart:
			sf := e.StartFile()
			fmt.Println("start:", sf.EntryID)
		case mpv.EventEnd:
			ef := e.EndFile()
			fmt.Println("end:", ef.EntryID, ef.Reason)
			if ef.Reason == mpv.EndFileEOF {
				break loop
			} else if ef.Reason == mpv.EndFileError {
				t.Errorf("EventEnd: %v", err)
			}
		case mpv.EventShutdown:
			fmt.Println("shutdown:", e.EventID)
			break loop
		default:
			fmt.Println("event:", e.EventID)
		}

		if e.Error != nil {
			fmt.Println("error:", e.Error)
		}
	}
}
