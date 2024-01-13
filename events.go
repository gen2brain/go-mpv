package mpv

import (
	"strings"
	"unsafe"
)

// EventID type.
type EventID uint32

// EventID constants.
const (
	EventNone             EventID = 0
	EventShutdown         EventID = 1
	EventLogMsg           EventID = 2
	EventGetPropertyReply EventID = 3
	EventSetPropertyReply EventID = 4
	EventCommandReply     EventID = 5
	EventStart            EventID = 6
	EventEnd              EventID = 7
	EventFileLoaded       EventID = 8
	EventClientMessage    EventID = 16
	EventVideoReconfig    EventID = 17
	EventAudioReconfig    EventID = 18
	EventSeek             EventID = 20
	EventPlaybackRestart  EventID = 21
	EventPropertyChange   EventID = 22
	EventQueueOverflow    EventID = 24
	EventHook             EventID = 25
)

var eventMap = map[EventID]string{
	EventNone:             "none",
	EventShutdown:         "shutdown",
	EventLogMsg:           "log-message",
	EventGetPropertyReply: "get-property-reply",
	EventSetPropertyReply: "set-property-reply",
	EventCommandReply:     "command-reply",
	EventStart:            "start-file",
	EventEnd:              "end-file",
	EventFileLoaded:       "file-loaded",
	EventClientMessage:    "client-message",
	EventVideoReconfig:    "video-reconfig",
	EventAudioReconfig:    "audio-reconfig",
	EventSeek:             "seek",
	EventPlaybackRestart:  "playback-restart",
	EventPropertyChange:   "property-change",
	EventQueueOverflow:    "event-queue-overflow",
	EventHook:             "hook",
}

// String .
func (e EventID) String() string {
	var str string

	_, ok := eventMap[e]
	if ok {
		str = eventMap[e]
	}

	return str
}

// Format is data format for options and properties.
type Format uint32

// Data formats.
const (
	FormatNone Format = iota
	FormatString
	FormatOsdString
	FormatFlag
	FormatInt64
	FormatDouble
)

// Reason is end file reason.
type Reason uint32

// End file reasons.
const (
	EndFileEOF      Reason = 0
	EndFileStop     Reason = 2
	EndFileQuit     Reason = 3
	EndFileError    Reason = 4
	EndFileRedirect Reason = 5
)

var endMap = map[Reason]string{
	EndFileEOF:      "eof",
	EndFileStop:     "stop",
	EndFileQuit:     "quit",
	EndFileError:    "error",
	EndFileRedirect: "redirect",
}

// String .
func (r Reason) String() string {
	var str string

	_, ok := endMap[r]
	if ok {
		str = endMap[r]
	}

	return str
}

// Event represents an mpv_event struct.
type Event struct {
	EventID       EventID
	Error         error
	ReplyUserdata uint64
	Data          unsafe.Pointer
}

type event struct {
	EventID       uint32
	Error         int32
	ReplyUserdata uint64
	Data          unsafe.Pointer
}

// LogMessage returns EventLogMessage.
func (e *Event) LogMessage() EventLogMessage {
	s := (*eventLogMessage)(e.Data)
	var elm EventLogMessage

	elm.LogLevel = s.LogLevel
	elm.Prefix = toStr(s.Prefix)
	elm.Level = toStr(s.Level)
	elm.Text = toStr(s.Text)

	return elm
}

// Property returns EventProperty.
func (e *Event) Property() EventProperty {
	s := (*eventProperty)(e.Data)
	var ep EventProperty

	ep.Name = toStr(s.Name)
	ep.Format = Format(s.Format)

	switch ep.Format {
	case FormatNone:
		ep.Data = nil
	case FormatString, FormatOsdString:
		ep.Data = toStr(s.Data)
	case FormatFlag:
		ep.Data = *(*int)(s.Data)
	case FormatInt64:
		ep.Data = *(*int64)(s.Data)
	case FormatDouble:
		ep.Data = *(*float64)(s.Data)
	default:
		ep.Data = nil
	}

	return ep
}

// StartFile returns EventStartFile.
func (e *Event) StartFile() EventStartFile {
	s := (*EventStartFile)(e.Data)
	var esf EventStartFile

	esf.EntryID = s.EntryID

	return esf
}

// EndFile returns EventEndFile.
func (e *Event) EndFile() EventEndFile {
	s := (*eventEndFile)(e.Data)
	var eef EventEndFile

	eef.Reason = Reason(s.Reason)
	eef.Error = newError(int(s.Error))
	eef.EntryID = s.EntryID
	eef.InsertID = s.InsertID
	eef.InsertNumEntries = s.InsertNumEntries

	return eef
}

// EventProperty type.
type EventProperty struct {
	Name   string
	Format Format
	Data   any
}

type eventProperty struct {
	Name   unsafe.Pointer
	Format uint32
	Data   unsafe.Pointer
}

// EventLogMessage type.
type EventLogMessage struct {
	Prefix   string
	Level    string
	Text     string
	LogLevel uint32
}

type eventLogMessage struct {
	Prefix   unsafe.Pointer
	Level    unsafe.Pointer
	Text     unsafe.Pointer
	LogLevel uint32
	_        [4]byte
}

// EventStartFile type.
type EventStartFile struct {
	EntryID int64
}

// EventEndFile type.
type EventEndFile struct {
	Reason           Reason
	Error            error
	EntryID          int64
	InsertID         int64
	InsertNumEntries int32
}

type eventEndFile struct {
	Reason           int32
	Error            int32
	EntryID          int64
	InsertID         int64
	InsertNumEntries int32
	_                [4]byte
}

func toStr(p unsafe.Pointer) string {
	str := unsafe.String((*byte)(p), 16*1024)
	idx := strings.Index(str, "\x00")
	if idx != -1 {
		str = str[:idx]
	}

	return strings.TrimSpace(str)
}
