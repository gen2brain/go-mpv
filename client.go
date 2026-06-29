//go:build cgo && !nocgo

// Package mpv provides cgo bindings for libmpv.
package mpv

/*
#include <mpv/client.h>
#include <stdlib.h>

static char** makeCharArray(int size) {
    return calloc(sizeof(char*), size);
}

static void setStringArray(char** a, int i, char* s) {
    a[i] = s;
}

#cgo !pkgconfig LDFLAGS: -lmpv
#cgo pkgconfig,!static pkg-config: mpv
#cgo pkgconfig,static pkg-config: --static mpv
*/
import "C"

import (
	"unsafe"
)

func init() {
	cAlloc = func(size int) unsafe.Pointer { return C.malloc(C.size_t(size)) }
	cFree = func(p unsafe.Pointer) { C.free(p) }
}

// Mpv represents an mpv client.
type Mpv struct {
	handle *C.mpv_handle
}

// New creates a new mpv instance and an associated client API handle.
func New() *Mpv {
	return &Mpv{C.mpv_create()}
}

// APIVersion returns the client api version the mpv source has been compiled with.
func (m *Mpv) APIVersion() uint64 {
	return uint64(C.mpv_client_api_version())
}

// Name returns the name of this client handle.
func (m *Mpv) Name() string {
	return C.GoString(C.mpv_client_name(m.handle))
}

// ID returns the ID of this client handle.
func (m *Mpv) ID() int64 {
	return int64(C.mpv_client_id(m.handle))
}

// Initialize initializes an uninitialized mpv instance.
func (m *Mpv) Initialize() error {
	return newError(int(C.mpv_initialize(m.handle)))
}

// TerminateDestroy terminates mpv and destroys the client.
func (m *Mpv) TerminateDestroy() {
	C.mpv_terminate_destroy(m.handle)
}

// Destroy disconnects and destroys this client handle without terminating mpv.
func (m *Mpv) Destroy() {
	C.mpv_destroy(m.handle)
}

// LoadConfigFile loads the given config file.
func (m *Mpv) LoadConfigFile(fileName string) error {
	cfileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cfileName))

	return newError(int(C.mpv_load_config_file(m.handle, cfileName)))
}

// TimeUS returns the internal time in microseconds.
func (m *Mpv) TimeUS() int64 {
	return int64(C.mpv_get_time_us(m.handle))
}

// TimeNS returns the internal time in nanoseconds.
func (m *Mpv) TimeNS() int64 {
	return int64(C.mpv_get_time_ns(m.handle))
}

// SetOption sets the given option according to the given format.
func (m *Mpv) SetOption(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cdata, cleanup := convertData(format, data)
	defer cleanup()

	return newError(int(C.mpv_set_option(m.handle, cname, C.mpv_format(format), cdata)))
}

// SetOptionString sets the option to the given string.
func (m *Mpv) SetOptionString(name, value string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	return newError(int(C.mpv_set_option_string(m.handle, cname, cvalue)))
}

// Command runs the specified command, returning an error if something goes wrong.
func (m *Mpv) Command(cmd []string) error {
	arr := C.makeCharArray(C.int(len(cmd) + 1))
	if arr == nil {
		return ErrNomem
	}
	defer C.free(unsafe.Pointer(arr))

	for i, s := range cmd {
		cs := C.CString(s)
		defer C.free(unsafe.Pointer(cs))
		C.setStringArray(arr, C.int(i), cs)
	}

	return newError(int(C.mpv_command(m.handle, arr)))
}

// CommandString runs the given command string, this string is parsed internally by mpv.
func (m *Mpv) CommandString(cmd string) error {
	ccmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(ccmd))

	return newError(int(C.mpv_command_string(m.handle, ccmd)))
}

// CommandAsync runs the command asynchronously.
func (m *Mpv) CommandAsync(replyUserdata uint64, cmd []string) error {
	arr := C.makeCharArray(C.int(len(cmd) + 1))
	if arr == nil {
		return ErrNomem
	}
	defer C.free(unsafe.Pointer(arr))

	for i, s := range cmd {
		cs := C.CString(s)
		defer C.free(unsafe.Pointer(cs))
		C.setStringArray(arr, C.int(i), cs)
	}

	return newError(int(C.mpv_command_async(m.handle, C.uint64_t(replyUserdata), arr)))
}

// CommandNode runs a command given as a []any or map[string]any and returns its result.
func (m *Mpv) CommandNode(args interface{}) (interface{}, error) {
	cargs, cleanup := goToNode(args)
	defer cleanup()

	var result C.mpv_node
	err := newError(int(C.mpv_command_node(m.handle, (*C.mpv_node)(cargs), &result)))
	if err != nil {
		return nil, err
	}
	defer C.mpv_free_node_contents(&result)

	return nodeToGo(unsafe.Pointer(&result)), nil
}

// CommandNodeAsync runs a structured command asynchronously.
func (m *Mpv) CommandNodeAsync(replyUserdata uint64, args interface{}) error {
	cargs, cleanup := goToNode(args)
	defer cleanup()

	return newError(int(C.mpv_command_node_async(m.handle, C.uint64_t(replyUserdata), (*C.mpv_node)(cargs))))
}

// AbortAsyncCommand aborts an outstanding asynchronous command with the given reply userdata.
func (m *Mpv) AbortAsyncCommand(replyUserdata uint64) {
	C.mpv_abort_async_command(m.handle, C.uint64_t(replyUserdata))
}

// SetProperty sets the client property according to the given format.
func (m *Mpv) SetProperty(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cdata, cleanup := convertData(format, data)
	defer cleanup()

	return newError(int(C.mpv_set_property(m.handle, cname, C.mpv_format(format), cdata)))
}

// SetPropertyString sets the property to the given string.
func (m *Mpv) SetPropertyString(name, value string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	return newError(int(C.mpv_set_property_string(m.handle, cname, cvalue)))
}

// DelProperty deletes the given property.
func (m *Mpv) DelProperty(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return newError(int(C.mpv_del_property(m.handle, cname)))
}

// SetPropertyAsync sets a property asynchronously.
func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cdata, cleanup := convertData(format, data)
	defer cleanup()

	return newError(int(C.mpv_set_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format), cdata)))
}

// GetProperty returns the value of the property according to the given format.
func (m *Mpv) GetProperty(name string, format Format) (interface{}, error) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	switch format {
	case FormatNone:
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), nil)))
		if err != nil {
			return nil, err
		}
		return nil, nil
	case FormatString, FormatOsdString:
		var result *C.char
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), unsafe.Pointer(&result))))
		if err != nil {
			return nil, err
		}
		defer C.mpv_free(unsafe.Pointer(result))
		return C.GoString(result), nil
	case FormatFlag:
		var result C.int
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), unsafe.Pointer(&result))))
		if err != nil {
			return nil, err
		}
		return result == 1, nil
	case FormatInt64:
		var result C.int64_t
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), unsafe.Pointer(&result))))
		if err != nil {
			return nil, err
		}
		return int64(result), nil
	case FormatDouble:
		var result C.double
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), unsafe.Pointer(&result))))
		if err != nil {
			return nil, err
		}
		return float64(result), nil
	case FormatNode:
		var result C.mpv_node
		err := newError(int(C.mpv_get_property(m.handle, n, C.mpv_format(format), unsafe.Pointer(&result))))
		if err != nil {
			return nil, err
		}
		defer C.mpv_free_node_contents(&result)
		return nodeToGo(unsafe.Pointer(&result)), nil
	default:
		return nil, ErrUnknownFormat
	}
}

// GetPropertyString returns the value of the property as a string.
// If the property is empty, an empty string is returned.
func (m *Mpv) GetPropertyString(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	str := C.mpv_get_property_string(m.handle, cname)
	defer C.mpv_free(unsafe.Pointer(str))

	if str != nil {
		return C.GoString(str)
	}

	return ""
}

// GetPropertyOsdString returns the value of the property as a string formatted for on-screen display.
func (m *Mpv) GetPropertyOsdString(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	str := C.mpv_get_property_osd_string(m.handle, cname)
	defer C.mpv_free(unsafe.Pointer(str))

	if str != nil {
		return C.GoString(str)
	}

	return ""
}

// GetPropertyAsync gets a property asynchronously.
func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return newError(int(C.mpv_get_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format))))
}

// ObserveProperty gets a notification whenever the given property changes.
func (m *Mpv) ObserveProperty(replyUserdata uint64, name string, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return newError(int(C.mpv_observe_property(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format))))
}

// UnobserveProperty will remove all observed properties for passed replyUserdata.
func (m *Mpv) UnobserveProperty(replyUserdata uint64) error {
	return newError(int(C.mpv_unobserve_property(m.handle, C.uint64_t(replyUserdata))))
}

// RequestEvent enables or disables the given event.
func (m *Mpv) RequestEvent(event EventID, enable bool) error {
	var enable_ C.int
	if enable {
		enable_ = 1
	}

	return newError(int(C.mpv_request_event(m.handle, C.mpv_event_id(event), enable_)))
}

// RequestLogMessages enables or disables receiving of log messages.
// Valid log levels: no fatal error warn info v debug trace.
func (m *Mpv) RequestLogMessages(level string) error {
	clevel := C.CString(level)
	defer C.free(unsafe.Pointer(clevel))

	return newError(int(C.mpv_request_log_messages(m.handle, clevel)))
}

// WaitEvent calls mpv_wait_event and returns the result as an Event struct.
func (m *Mpv) WaitEvent(timeout float64) *Event {
	ev := C.mpv_wait_event(m.handle, C.double(timeout))

	return &Event{
		EventID:       EventID(ev.event_id),
		Data:          unsafe.Pointer(ev.data),
		ReplyUserdata: uint64(ev.reply_userdata),
		Error:         newError(int(ev.error)),
	}
}

// Wakeup interrupts the current mpv_wait_event() call.
func (m *Mpv) Wakeup() {
	C.mpv_wakeup(m.handle)
}

// WakeupPipe returns the read end of a pipe that signals new events, or -1 on error.
func (m *Mpv) WakeupPipe() int {
	return int(C.mpv_get_wakeup_pipe(m.handle))
}

// WaitAsyncRequests blocks until all asynchronous requests are done.
func (m *Mpv) WaitAsyncRequests() {
	C.mpv_wait_async_requests(m.handle)
}

// convertData converts data for the given format into a pointer for SetOption/SetProperty,
// and a cleanup function that must be called once the pointer is no longer needed.
func convertData(format Format, data interface{}) (unsafe.Pointer, func()) {
	switch format {
	case FormatNone:
		return nil, func() {}
	case FormatString, FormatOsdString:
		cstr := C.CString(data.(string))
		return unsafe.Pointer(&cstr), func() { C.free(unsafe.Pointer(cstr)) }
	case FormatFlag:
		var val C.int
		if data.(bool) {
			val = 1
		} else {
			val = 0
		}
		return unsafe.Pointer(&val), func() {}
	case FormatInt64:
		i, ok := data.(int64)
		if !ok {
			i = int64(data.(int))
		}
		val := C.int64_t(i)
		return unsafe.Pointer(&val), func() {}
	case FormatDouble:
		val := C.double(data.(float64))
		return unsafe.Pointer(&val), func() {}
	case FormatNode:
		return goToNode(data)
	default:
		return nil, func() {}
	}
}
