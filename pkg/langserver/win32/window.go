// +build windows

package win32

import (
	"context"
	"syscall"
	"time"
	"unsafe"
)

var (
	forgrWindow   = user32.MustFindProc("GetForegroundWindow")
	windowText    = user32.MustFindProc("GetWindowTextW")
	windowTextLen = user32.MustFindProc("GetWindowTextLengthW")
)

// GetForegroundWindow returns the name of the window that currently has the user-focus
func GetForegroundWindow() string {
	hwnd, _, _ := forgrWindow.Call()

	if hwnd != 0 {
		textlen, _, _ := windowTextLen.Call(hwnd)
		buf := make([]uint16, textlen+1)
		windowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(textlen+1))
		return syscall.UTF16ToString(buf)
	}

	return ""
}

// WaitForWindowChange blocks until the window with the user-focus changes (or ctx is cancelled). Returns the title of the new active window
func WaitForWindowChange(ctx context.Context) string {
	oldWindow := GetForegroundWindow()
	for {
		time.Sleep(200 * time.Millisecond)
		newWindow := GetForegroundWindow()
		if newWindow != oldWindow {
			return newWindow
		}
		if ctx != nil && ctx.Err() != nil {
			return oldWindow
		}

	}
}
