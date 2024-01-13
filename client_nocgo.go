//go:build !cgo || purego

// Package mpv provides purego bindings for libmpv.
package mpv

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	libmpv uintptr
)

var create func() uintptr
var apiVersion func() uint64
var name func(handle uintptr) string
var id func(handle uintptr) int64
var initialize func(handle uintptr) int
var terminateDestroy func(handle uintptr)
var loadConfigFile func(handle uintptr, fileName string) int
var timeUS func(handle uintptr) int64
var setOption func(handle uintptr, name string, format int, data unsafe.Pointer) int
var setOptionString func(handle uintptr, name, value string) int
var command func(handle uintptr, cmd **byte) int
var commandString func(handle uintptr, cmd string) int
var commandAsync func(handle uintptr, replyUserdata uint64, cmd **byte) int
var setProperty func(handle uintptr, name string, format int, data unsafe.Pointer) int
var setPropertyString func(handle uintptr, name, value string) int
var setPropertyAsync func(handle uintptr, replyUserdata uint64, name string, format int, data unsafe.Pointer) int
var getProperty func(handle uintptr, name string, format int, data unsafe.Pointer) int
var getPropertyString func(handle uintptr, name string) string
var getPropertyOsdString func(handle uintptr, name string) string
var getPropertyAsync func(handle uintptr, name string, replyUserdata uint64, format int) int
var observeProperty func(handle uintptr, replyUserdata uint64, name string, format int) int
var unobserveProperty func(handle uintptr, replyUserdata uint64) int
var requestEvent func(handle uintptr, event int, enable bool) int
var requestLogMessages func(handle uintptr, level string) int
var waitEvent func(handle uintptr, timeout float64) *event
var wakeup func(handle uintptr)
var waitAsyncRequests func(handle uintptr)

func init() {
	libmpv = loadLibrary()

	purego.RegisterLibFunc(&create, libmpv, "mpv_create")
	purego.RegisterLibFunc(&apiVersion, libmpv, "mpv_client_api_version")
	purego.RegisterLibFunc(&name, libmpv, "mpv_client_name")
	purego.RegisterLibFunc(&id, libmpv, "mpv_client_id")
	purego.RegisterLibFunc(&initialize, libmpv, "mpv_initialize")
	purego.RegisterLibFunc(&terminateDestroy, libmpv, "mpv_terminate_destroy")
	purego.RegisterLibFunc(&loadConfigFile, libmpv, "mpv_load_config_file")
	purego.RegisterLibFunc(&timeUS, libmpv, "mpv_get_time_us")
	purego.RegisterLibFunc(&setOption, libmpv, "mpv_set_option")
	purego.RegisterLibFunc(&setOptionString, libmpv, "mpv_set_option_string")
	purego.RegisterLibFunc(&command, libmpv, "mpv_command")
	purego.RegisterLibFunc(&commandString, libmpv, "mpv_command_string")
	purego.RegisterLibFunc(&commandAsync, libmpv, "mpv_command_async")
	purego.RegisterLibFunc(&setProperty, libmpv, "mpv_set_property")
	purego.RegisterLibFunc(&setPropertyString, libmpv, "mpv_set_property_string")
	purego.RegisterLibFunc(&setPropertyAsync, libmpv, "mpv_set_property_async")
	purego.RegisterLibFunc(&getProperty, libmpv, "mpv_get_property")
	purego.RegisterLibFunc(&getPropertyString, libmpv, "mpv_get_property_string")
	purego.RegisterLibFunc(&getPropertyOsdString, libmpv, "mpv_get_property_osd_string")
	purego.RegisterLibFunc(&getPropertyAsync, libmpv, "mpv_get_property_async")
	purego.RegisterLibFunc(&observeProperty, libmpv, "mpv_observe_property")
	purego.RegisterLibFunc(&unobserveProperty, libmpv, "mpv_unobserve_property")
	purego.RegisterLibFunc(&requestEvent, libmpv, "mpv_request_event")
	purego.RegisterLibFunc(&requestLogMessages, libmpv, "mpv_request_log_messages")
	purego.RegisterLibFunc(&waitEvent, libmpv, "mpv_wait_event")
	purego.RegisterLibFunc(&wakeup, libmpv, "mpv_wakeup")
	purego.RegisterLibFunc(&waitAsyncRequests, libmpv, "mpv_wait_async_requests")
}

// Mpv represents an mpv client.
type Mpv struct {
	handle uintptr
}

// New creates a new mpv instance and an associated client API handle.
func New() *Mpv {
	return &Mpv{create()}
}

// APIVersion returns the client api version the mpv source has been compiled with.
func (m *Mpv) APIVersion() uint64 {
	return apiVersion()
}

// Name returns the name of this client handle.
func (m *Mpv) Name() string {
	return name(m.handle)
}

// ID returns the id of this client handle.
func (m *Mpv) ID() int64 {
	return id(m.handle)
}

// Initialize initializes an uninitialized mpv instance.
func (m *Mpv) Initialize() error {
	return newError(initialize(m.handle))
}

// TerminateDestroy terminates mpv and destroys the client.
func (m *Mpv) TerminateDestroy() {
	terminateDestroy(m.handle)
}

// LoadConfigFile loads the given config file.
func (m *Mpv) LoadConfigFile(fileName string) error {
	return newError(loadConfigFile(m.handle, fileName))
}

// TimeUS returns the internal time in microseconds.
func (m *Mpv) TimeUS() int64 {
	return timeUS(m.handle)
}

// SetOption sets the given option according to the given format.
func (m *Mpv) SetOption(name string, format Format, data interface{}) error {
	return newError(setOption(m.handle, name, int(format), convertData(format, data)))
}

// SetOptionString sets the option to the given string.
func (m *Mpv) SetOptionString(name, value string) error {
	return newError(setOptionString(m.handle, name, value))
}

// Command runs the specified command, returning an error if something goes wrong.
func (m *Mpv) Command(cmd []string) error {
	cmds := make([]*byte, 0)
	for _, c := range cmd {
		cmds = append(cmds, cStr(c))
	}

	return newError(command(m.handle, unsafe.SliceData(cmds)))
}

// CommandString runs the given command string, this string is parsed internally by mpv.
func (m *Mpv) CommandString(cmd string) error {
	return newError(commandString(m.handle, cmd))
}

// CommandAsync runs the command asynchronously.
func (m *Mpv) CommandAsync(replyUserdata uint64, cmd []string) error {
	cmds := make([]*byte, 0)
	for _, c := range cmd {
		cmds = append(cmds, cStr(c))
	}

	return newError(commandAsync(m.handle, replyUserdata, unsafe.SliceData(cmds)))
}

// SetProperty sets the client property according to the given format.
func (m *Mpv) SetProperty(name string, format Format, data interface{}) error {
	return newError(setProperty(m.handle, name, int(format), convertData(format, data)))
}

// SetPropertyString sets the property to the given string.
func (m *Mpv) SetPropertyString(name, value string) error {
	return newError(setPropertyString(m.handle, name, value))
}

// SetPropertyAsync sets a property asynchronously.
func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	return newError(setPropertyAsync(m.handle, replyUserdata, name, int(format), convertData(format, data)))
}

// GetProperty returns the value of the property according to the given format.
func (m *Mpv) GetProperty(name string, format Format) (interface{}, error) {
	switch format {
	case FormatNone:
		err := newError(getProperty(m.handle, name, int(format), nil))
		if err != nil {
			return nil, err
		}
		return nil, nil
	case FormatString, FormatOsdString:
		var result *byte
		err := newError(getProperty(m.handle, name, int(format), unsafe.Pointer(&result)))
		if err != nil {
			return nil, err
		}
		return toStr(unsafe.Pointer(result)), nil
	case FormatFlag:
		var result int
		err := newError(getProperty(m.handle, name, int(format), unsafe.Pointer(&result)))
		if err != nil {
			return nil, err
		}
		return result == 1, nil
	case FormatInt64:
		var result int64
		err := newError(getProperty(m.handle, name, int(format), unsafe.Pointer(&result)))
		if err != nil {
			return nil, err
		}
		return int64(result), nil
	case FormatDouble:
		var result float64
		err := newError(getProperty(m.handle, name, int(format), unsafe.Pointer(&result)))
		if err != nil {
			return nil, err
		}
		return float64(result), nil
	default:
		return nil, ErrUnknownFormat
	}
}

// GetPropertyString returns the value of the property as a string.
// If the property is empty, an empty string is returned.
func (m *Mpv) GetPropertyString(name string) string {
	return getPropertyString(m.handle, name)
}

// GetPropertyOsdString returns the value of the property as a string formatted for on-screen display.
func (m *Mpv) GetPropertyOsdString(name string) string {
	return getPropertyOsdString(m.handle, name)
}

// GetPropertyAsync gets a property asynchronously.
func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	return newError(getPropertyAsync(m.handle, name, replyUserdata, int(format)))
}

// ObserveProperty gets a notification whenever the given property changes.
func (m *Mpv) ObserveProperty(replyUserdata uint64, name string, format Format) error {
	return newError(observeProperty(m.handle, replyUserdata, name, int(format)))
}

// UnobserveProperty will remove all observed properties for passed replyUserdata.
func (m *Mpv) UnobserveProperty(replyUserdata uint64) error {
	return newError(unobserveProperty(m.handle, replyUserdata))
}

// RequestEvent enables or disables the given event.
func (m *Mpv) RequestEvent(event EventID, enable bool) error {
	return newError(requestEvent(m.handle, int(event), enable))
}

// RequestLogMessages enables or disables receiving of log messages.
// Valid log levels: no fatal error warn info v debug trace.
func (m *Mpv) RequestLogMessages(level string) error {
	return newError(requestLogMessages(m.handle, level))
}

// WaitEvent calls mpv_wait_event and returns the result as an Event struct.
func (m *Mpv) WaitEvent(timeout float64) *Event {
	ev := waitEvent(m.handle, timeout)

	return &Event{
		EventID:       EventID(ev.EventID),
		Error:         newError(int(ev.Error)),
		ReplyUserdata: ev.ReplyUserdata,
		Data:          ev.Data,
	}
}

// Wakeup interrupts the current WaitEvent() call.
func (m *Mpv) Wakeup() {
	wakeup(m.handle)
}

// WaitAsyncRequests blocks until all asynchronous requests are done.
func (m *Mpv) WaitAsyncRequests() {
	waitAsyncRequests(m.handle)
}

// convertData converts the data according to the given format and returns an unsafe.Pointer
// for use in SetOption and SetProperty.
func convertData(format Format, data interface{}) unsafe.Pointer {
	switch format {
	case FormatNone:
		return nil
	case FormatString, FormatOsdString:
		return unsafe.Pointer(&[]byte(data.(string))[0])
	case FormatFlag:
		var val int
		if data.(bool) {
			val = 1
		} else {
			val = 0
		}
		return unsafe.Pointer(&val)
	case FormatInt64:
		val, ok := data.(int64)
		if !ok {
			val = int64(data.(int))
		}
		return unsafe.Pointer(&val)
	case FormatDouble:
		val := data.(float64)
		return unsafe.Pointer(&val)
	default:
		return nil
	}
}

func cStr(str string) *byte {
	bs := []byte(str)
	if len(bs) == 0 || bs[len(bs)-1] != 0 {
		bs = append(bs, 0)
	}

	return &bs[0]
}
