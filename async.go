package mpv

/*
#include <mpv/client.h>
#include <stdlib.h>
char** makeCharArray(int size);
void setStringArray(char** a, int i, char* s);
*/
import "C"

import (
	"unsafe"
)

// SetPropertyAsync .
func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	return NewError(C.mpv_set_property_async(m.handle, C.uint64_t(replyUserdata), C.CString(name), C.mpv_format(format), convertData(format, data)))
}

// GetPropertyAsync .
func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	return NewError(C.mpv_get_property_async(m.handle, C.uint64_t(replyUserdata), C.CString(name), C.mpv_format(format)))
}

// CommandAsync .
func (m *Mpv) CommandAsync(replyUserdata uint64, command []string) error {
	arr := C.makeCharArray(C.int(len(command) + 1))
	if arr == nil {
		return ERROR_NOMEM
	}
	defer C.free(unsafe.Pointer(arr))
	for i, s := range command {
		C.setStringArray(arr, C.int(i), C.CString(s))
	}
	return NewError(C.mpv_command_async(m.handle, C.uint64_t(replyUserdata), arr))
}

// CommandNodeAsync .
func (m *Mpv) CommandNodeAsync(replyUserdata uint64, args Node) error {
	return NewError(C.mpv_command_node_async(m.handle, C.uint64_t(replyUserdata), args.CNode()))
}
