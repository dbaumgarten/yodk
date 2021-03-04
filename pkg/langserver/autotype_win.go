// +build windows

package langserver

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dbaumgarten/yodk/pkg/langserver/win32"
)

var (
	// AutotypeHotkey is the hotkey to trigger auto-typing
	AutotypeHotkey = &win32.Hotkey{
		ID:        1,
		Modifiers: win32.ModCtrl,
		KeyCode:   'I',
	}
	// AutotypeSSCHotkey is the hotkey to trigger auto-typing when in the SSC
	AutotypeSSCHotkey = &win32.Hotkey{
		ID:        1,
		Modifiers: win32.ModCtrl,
		KeyCode:   'U',
	}
	// AutodeleteHotkey is the hotkey to trigger auto-deletion
	AutodeleteHotkey = &win32.Hotkey{
		ID:        2,
		Modifiers: win32.ModCtrl,
		KeyCode:   'P',
	}
	// AutooverwriteHotkey is the hotkey to overwrite the current line with new ones
	AutooverwriteHotkey = &win32.Hotkey{
		ID:        3,
		Modifiers: win32.ModCtrl,
		KeyCode:   'O',
	}
)

const typeDelay = 40 * time.Millisecond

// ListenForHotkeys listens for global hotkeys and dispatches the registered actions
func (ls *LangServer) ListenForHotkeys() {
	go func() {
		currentWindow := win32.GetForegroundWindow()
		wg := sync.WaitGroup{}
		var cancelHotkeyListening context.CancelFunc
		var hotkeysRegistered = false
		for {
			if isStarbaseWindow(currentWindow) && !hotkeysRegistered {
				ctx := context.Background()
				ctx, cancelHotkeyListening = context.WithCancel(ctx)
				hotkeysRegistered = true
				go func() {
					wg.Add(1)
					err := win32.ListenForHotkeys(ctx, ls.hotkeyHandler, AutotypeHotkey, AutodeleteHotkey, AutooverwriteHotkey)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error when registering hotkeys: %s", err)
					}
					wg.Done()
				}()
			} else if hotkeysRegistered {
				cancelHotkeyListening()
				wg.Wait()
				hotkeysRegistered = false
			}
			currentWindow = win32.WaitForWindowChange(nil)
		}
	}()
}

func isStarbaseWindow(name string) bool {
	return name == "Starbase"
}

func (ls *LangServer) hotkeyHandler(hk win32.Hotkey) {
	win32.SendInput(win32.KeyUpInput(win32.KeycodeCtrl))
	win32.SendInput(win32.KeyUpInput(uint16(hk.KeyCode)))
	switch hk.ID {
	case AutotypeHotkey.ID:
		if code := ls.getLastOpenedCode(); code == code {
			typeYololCode(code)
		}
	case AutotypeSSCHotkey.ID:
		if code := ls.getLastOpenedCode(); code == code {
			typeYololSSCCode(code)
		}
	case AutodeleteHotkey.ID:
		deleteAllLines()
	case AutooverwriteHotkey.ID:
		if code := ls.getLastOpenedCode(); code == code {
			overwriteYololCode(code)
		}
	}
}

func (ls *LangServer) getLastOpenedCode() string {
	ls.cache.Lock.Lock()
	lastOpened := ls.cache.LastOpenedYololFile
	ls.cache.Lock.Unlock()

	if lastOpened != "" {
		code, err := ls.cache.Get(lastOpened)
		if err == nil {
			return code
		}
	}
	return ""
}

func typeYololCode(code string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		win32.SendString(line)
		time.Sleep(typeDelay)
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
}

func typeYololSSCCode(code string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		win32.SendString(line)
		time.Sleep(typeDelay)
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
}

func overwriteYololCode(code string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		deleteLine()
		win32.SendString(line)
		time.Sleep(typeDelay)
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
}

func deleteAllLines() {
	for i := 0; i < 20; i++ {
		deleteLine()
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
}

func deleteLine() {
	win32.SendInput(win32.KeyDownInput(win32.KeycodeCtrl), win32.KeyDownInput('A'))
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyUpInput('A'), win32.KeyUpInput(win32.KeycodeCtrl))
	win32.SendInput(win32.KeyDownInput(win32.KeycodeBackspace), win32.KeyUpInput(win32.KeycodeBackspace))
	time.Sleep(typeDelay)
}
