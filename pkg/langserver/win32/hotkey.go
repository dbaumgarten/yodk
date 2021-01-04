// +build windows

package win32

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"unsafe"
)

// most of the code is taken from: https://stackoverflow.com/questions/38646794/implement-a-global-hotkey-in-golang#38954281

// Modifiers for Hotkeys
const (
	ModAlt = 1 << iota
	ModCtrl
	ModShift
	ModWin
)

var (
	reghotkey = user32.MustFindProc("RegisterHotKey")
	getmsg    = user32.MustFindProc("GetMessageW")
)

// Hotkey represents a key-combination pressed by a user
type Hotkey struct {
	// Id, must be unique for each registered hotkey
	ID int // Unique id
	// Modifiers is a bitmask containing modifiers for the hotkey
	Modifiers int // Mask of modifiers
	// KeyCode is the keycode for the hotkey
	KeyCode int // Key code, e.g. 'A'
}

// String returns a human-friendly display name of the hotkey
// such as "Hotkey[Id: 1, Alt+Ctrl+O]"
func (h Hotkey) String() string {
	mod := &bytes.Buffer{}
	if h.Modifiers&ModAlt != 0 {
		mod.WriteString("Alt+")
	}
	if h.Modifiers&ModCtrl != 0 {
		mod.WriteString("Ctrl+")
	}
	if h.Modifiers&ModShift != 0 {
		mod.WriteString("Shift+")
	}
	if h.Modifiers&ModWin != 0 {
		mod.WriteString("Win+")
	}
	return fmt.Sprintf("Hotkey[Id: %d, %s%c]", h.ID, mod, h.KeyCode)
}

// HotkeyHandler is the callback for registered hotkeys
type HotkeyHandler func(Hotkey)

type msg struct {
	HWND   uintptr
	UINT   uintptr
	WPARAM int16
	LPARAM int64
	DWORD  int32
	POINT  struct{ X, Y int64 }
}

// ListenForHotkeys registers an listens for the given global Hotkeys. If a hotkey is pressed, the hendler function is executed
// This function blocks, so it shoue have it's own goroutine
func ListenForHotkeys(ctx context.Context, handler HotkeyHandler, hotkeys ...*Hotkey) error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hotkeymap := make(map[int16]*Hotkey)
	for _, v := range hotkeys {
		hotkeymap[int16(v.ID)] = v
		r1, _, err := reghotkey.Call(
			0, uintptr(v.ID), uintptr(v.Modifiers), uintptr(v.KeyCode))
		if r1 != 1 {
			return err
		}
	}

	for {
		if ctx != nil && ctx.Err() != nil {
			return nil
		}
		var msg = &msg{}
		getmsg.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)

		// Registered id is in the WPARAM field:
		if id := msg.WPARAM; id != 0 {
			hk, exists := hotkeymap[id]
			if exists {
				handler(*hk)
			}
		}
	}
}
