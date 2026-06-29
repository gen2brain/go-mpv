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
	FormatNode
	FormatNodeArray
	FormatNodeMap
	FormatByteArray
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
	elm.Text = strings.TrimSuffix(toStr(s.Text), "\n")

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
		ep.Data = int(*(*int32)(s.Data))
	case FormatInt64:
		ep.Data = *(*int64)(s.Data)
	case FormatDouble:
		ep.Data = *(*float64)(s.Data)
	case FormatNode:
		ep.Data = nodeToGo(s.Data)
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

// ClientMessage returns the arguments of a client message event.
func (e *Event) ClientMessage() []string {
	s := (*eventClientMessage)(e.Data)
	out := make([]string, s.NumArgs)
	if s.NumArgs > 0 {
		args := unsafe.Slice((*unsafe.Pointer)(s.Args), int(s.NumArgs))
		for i := range args {
			out[i] = toStr(args[i])
		}
	}

	return out
}

// Hook returns the hook event. Its ID must be passed to HookContinue.
func (e *Event) Hook() Hook {
	s := (*eventHook)(e.Data)

	return Hook{Name: toStr(s.Name), ID: s.ID}
}

// CommandReply returns the result of an asynchronous command.
func (e *Event) CommandReply() any {
	return nodeToGo(e.Data)
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

type eventClientMessage struct {
	NumArgs int32
	Args    unsafe.Pointer
}

// Hook is the payload of a hook event.
type Hook struct {
	Name string
	ID   uint64
}

type eventHook struct {
	Name unsafe.Pointer
	ID   uint64
}

// toStr copies a NUL-terminated C string into a Go-owned string.
func toStr(p unsafe.Pointer) string {
	if p == nil {
		return ""
	}

	var n int
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}

	return string(unsafe.Slice((*byte)(p), n))
}
