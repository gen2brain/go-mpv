package mpv

//#include <mpv/client.h>
import "C"

import (
	"unsafe"
)

// EventId type.
type EventId int

// Events .
const (
	EVENT_NONE                  EventId = C.MPV_EVENT_NONE
	EVENT_SHUTDOWN              EventId = C.MPV_EVENT_SHUTDOWN
	EVENT_LOG_MESSAGE           EventId = C.MPV_EVENT_LOG_MESSAGE
	EVENT_GET_PROPERTY_REPLY    EventId = C.MPV_EVENT_GET_PROPERTY_REPLY
	EVENT_SET_PROPERTY_REPLY    EventId = C.MPV_EVENT_SET_PROPERTY_REPLY
	EVENT_COMMAND_REPLY         EventId = C.MPV_EVENT_COMMAND_REPLY
	EVENT_START_FILE            EventId = C.MPV_EVENT_START_FILE
	EVENT_END_FILE              EventId = C.MPV_EVENT_END_FILE
	EVENT_FILE_LOADED           EventId = C.MPV_EVENT_FILE_LOADED
	EVENT_IDLE                  EventId = C.MPV_EVENT_IDLE
	EVENT_TICK                  EventId = C.MPV_EVENT_TICK
	EVENT_CLIENT_MESSAGE        EventId = C.MPV_EVENT_CLIENT_MESSAGE
	EVENT_VIDEO_RECONFIG        EventId = C.MPV_EVENT_VIDEO_RECONFIG
	EVENT_AUDIO_RECONFIG        EventId = C.MPV_EVENT_AUDIO_RECONFIG
	EVENT_SEEK                  EventId = C.MPV_EVENT_SEEK
	EVENT_PLAYBACK_RESTART      EventId = C.MPV_EVENT_PLAYBACK_RESTART
	EVENT_PROPERTY_CHANGE       EventId = C.MPV_EVENT_PROPERTY_CHANGE
	EVENT_QUEUE_OVERFLOW        EventId = C.MPV_EVENT_QUEUE_OVERFLOW
)

// String .
func (e EventId) String() string {
	return C.GoString(C.mpv_event_name(C.mpv_event_id(e)))
}

// Event represents an mpv_event struct.
type Event struct {
	Event_Id       EventId
	Error          error
	Reply_Userdata uint64
	Data           unsafe.Pointer
}

// Message .
func (e *Event) Message() string {
	s := (*C.struct_mpv_event_log_message)(e.Data)
	return C.GoString((*C.char)(s.text))
}
