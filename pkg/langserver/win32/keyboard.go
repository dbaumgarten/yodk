// +build windows

package win32

import (
	"reflect"
	"syscall"
	"unsafe"
)

var (
	user32        = syscall.MustLoadDLL("user32.dll")
	sendInputProc = user32.MustFindProc("SendInput")
)

// KeyboardInput describes a keyboard-input event
type KeyboardInput struct {
	VirtualKeyCode uint16
	ScanCode       uint16
	Flags          uint32
	Time           uint32
	ExtraInfo      uint64
}

// Input describes an input-event
type Input struct {
	InputType     uint32
	KeyboardInput KeyboardInput
	Padding       uint64
}

// Keyboard-Input flags
const (
	KeyeventfExtendedkey = 0x0001
	KeyeventfKeyup       = 0x0002
	KeyeventfUnicode     = 0x0004
	KeyeventfScancode    = 0x0008
)

// Constants for special keycodes
const (
	KeycodeCtrl      = 0x11
	KeycodeAlt       = 0x12
	KeycodeReturn    = 0x0D
	KeycodeShift     = 0x10
	KeycodeBackspace = 0x08
	KeycodeDown      = 0x28
	KeycodeUp        = 0x26
	KeycodeRight     = 0x27
	KeycodeLeft      = 0x25
	KeycodeDelete    = 0x2E
	KeycodeEnd       = 0x23
)

// SendInput is a go-wrapper for the win32 SendInput function. It sends input events to the currently active window
func SendInput(inputEvents ...Input) error {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&inputEvents))
	data := unsafe.Pointer(hdr.Data)
	ret, _, err := sendInputProc.Call(
		uintptr(len(inputEvents)),
		uintptr(data),
		uintptr(unsafe.Sizeof(inputEvents[0])),
	)
	if int(ret) != len(inputEvents) {
		return err
	}
	return nil
}

// SendString sends input-events to the OS, that simulate typing the given string
func SendString(s string) {
	for _, r := range s {
		SendInput(UnicodeKeyDownInput(uint16(r)), UnicodeKeyUpInput(uint16(r)))
	}
}

// KeyDownInput returns the input-struct for pressing the given key (by virtual-keycode)
func KeyDownInput(keycode uint16) Input {
	return Input{
		InputType: 1,
		KeyboardInput: KeyboardInput{
			VirtualKeyCode: keycode,
		},
	}
}

// KeyUpInput returns the input-struct for releasing the given key (by virtual-keycode)
func KeyUpInput(keycode uint16) Input {
	return Input{
		InputType: 1,
		KeyboardInput: KeyboardInput{
			VirtualKeyCode: keycode,
			Flags:          KeyeventfKeyup,
		},
	}
}

// UnicodeKeyDownInput returns the input-struct for pressing the given key (by unicode codepoint)
func UnicodeKeyDownInput(keycode uint16) Input {
	return Input{
		InputType: 1,
		KeyboardInput: KeyboardInput{
			VirtualKeyCode: 0,
			ScanCode:       keycode,
			Flags:          KeyeventfUnicode,
		},
	}
}

// UnicodeKeyUpInput returns the input-struct for releasing the given key (by unicode codepoint)
func UnicodeKeyUpInput(keycode uint16) Input {
	return Input{
		InputType: 1,
		KeyboardInput: KeyboardInput{
			VirtualKeyCode: 0,
			ScanCode:       keycode,
			Flags:          KeyeventfKeyup | KeyeventfUnicode,
		},
	}
}
