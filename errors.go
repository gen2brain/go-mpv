package mpv

import (
	"errors"
)

var ErrEventQueueFull = errors.New("event queue full")
var ErrNomem = errors.New("memory allocation failed")
var ErrUninitialized = errors.New("core not uninitialized")
var ErrInvalidParameter = errors.New("invalid parameter")
var ErrOptionNotFound = errors.New("option not found")
var ErrOptionFormat = errors.New("unsupported format for accessing option")
var ErrOptionError = errors.New("error setting option")
var ErrPropertyNotFound = errors.New("property not found")
var ErrPropertyFormat = errors.New("unsupported format for accessing property")
var ErrPropertyUnavailable = errors.New("property unavailable")
var ErrPropertyError = errors.New("error accessing property")
var ErrCommand = errors.New("error running command")
var ErrLoadingFailed = errors.New("loading failed")
var ErrAoInitFailed = errors.New("audio output initialization failed")
var ErrVoInitFailed = errors.New("video output initialization failed")
var ErrNothingToPlay = errors.New("no audio or video data played")
var ErrUnknownFormat = errors.New("unrecognized file format")
var ErrUnsupported = errors.New("not supported")
var ErrNotImplemented = errors.New("operation not implemented")
var ErrGeneric = errors.New("something happened")

var ErrUnknown = errors.New("unknown error")

// Error constants.
const (
	errorSuccess             = 0
	errorEventQueueFull      = -1
	errorNomem               = -2
	errorUninitialized       = -3
	errorInvalidParameter    = -4
	errorOptionNotFound      = -5
	errorOptionFormat        = -6
	errorOptionError         = -7
	errorPropertyNotFound    = -8
	errorPropertyFormat      = -9
	errorPropertyUnavailable = -10
	errorPropertyError       = -11
	errorCommand             = -12
	errorLoadingFailed       = -13
	errorAoInitFailed        = -14
	errorVoInitFailed        = -15
	errorNothingToPlay       = -16
	errorUnknownFormat       = -17
	errorUnsupported         = -18
	errorNotImplemented      = -19
	errorGeneric             = -20
)

var errorMap = map[int]error{
	errorSuccess:             nil,
	errorEventQueueFull:      ErrEventQueueFull,
	errorNomem:               ErrNomem,
	errorUninitialized:       ErrUninitialized,
	errorInvalidParameter:    ErrInvalidParameter,
	errorOptionNotFound:      ErrOptionNotFound,
	errorOptionFormat:        ErrOptionFormat,
	errorOptionError:         ErrOptionError,
	errorPropertyNotFound:    ErrPropertyNotFound,
	errorPropertyFormat:      ErrPropertyFormat,
	errorPropertyUnavailable: ErrPropertyUnavailable,
	errorPropertyError:       ErrPropertyError,
	errorCommand:             ErrCommand,
	errorLoadingFailed:       ErrLoadingFailed,
	errorAoInitFailed:        ErrAoInitFailed,
	errorVoInitFailed:        ErrVoInitFailed,
	errorNothingToPlay:       ErrNothingToPlay,
	errorUnknownFormat:       ErrUnknownFormat,
	errorUnsupported:         ErrUnsupported,
	errorNotImplemented:      ErrNotImplemented,
	errorGeneric:             ErrGeneric,
}

// newError turns an integer value into an error type.
func newError(e int) error {
	var err error
	var ok bool

	err, ok = errorMap[e]
	if !ok {
		err = ErrUnknown
	}

	return err
}
