// +build windows

package langserver

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/dbaumgarten/yodk/pkg/langserver/win32"
)

var (
	// AutotypeHotkey is the hotkey to trigger auto-typing
	AutotypeHotkey = &win32.Hotkey{
		ID:        1,
		Modifiers: win32.ModCtrl,
		KeyCode:   'I',
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
	// SSCAutotypeHotkey is the hotkey to trigger auto-typing (in SSC-Mode)
	SSCAutotypeHotkey = &win32.Hotkey{
		ID:        4,
		Modifiers: win32.ModCtrl | win32.ModAlt,
		KeyCode:   'I',
	}
	// SSCAutodeleteHotkey is the hotkey to trigger auto-deletion (in SSC-Mode)
	SSCAutodeleteHotkey = &win32.Hotkey{
		ID:        5,
		Modifiers: win32.ModCtrl | win32.ModAlt,
		KeyCode:   'P',
	}
	// SSCAutooverwriteHotkey is the hotkey to overwrite the current line with new ones (in SSC-Mode)
	SSCAutooverwriteHotkey = &win32.Hotkey{
		ID:        6,
		Modifiers: win32.ModCtrl | win32.ModAlt,
		KeyCode:   'O',
	}
	// AutocopyHotkey is the hotkey to copy the selected chip into the clipboard
	AutocopyHotkey = &win32.Hotkey{
		ID:        7,
		Modifiers: win32.ModCtrl,
		KeyCode:   'J',
	}
	// SSCAutocopyHotkey is the hotkey to copy the selected chip into the clipboard (in SSC-Mode)
	SSCAutocopyHotkey = &win32.Hotkey{
		ID:        8,
		Modifiers: win32.ModCtrl | win32.ModAlt,
		KeyCode:   'J',
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
					err := win32.ListenForHotkeys(ctx, ls.hotkeyHandler, AutotypeHotkey, AutodeleteHotkey, AutooverwriteHotkey,
						SSCAutotypeHotkey, SSCAutodeleteHotkey, SSCAutooverwriteHotkey,
						AutocopyHotkey, SSCAutocopyHotkey)
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
	case AutodeleteHotkey.ID:
		deleteAllLines()
	case AutooverwriteHotkey.ID:
		if code := ls.getLastOpenedCode(); code == code {
			overwriteYololCode(code)
		}
	case AutocopyHotkey.ID:
		copyChipToClipboard()

	// same as above, but now for the SSC
	case SSCAutotypeHotkey.ID:
		win32.SendInput(win32.KeyUpInput(win32.KeycodeAlt))
		if code := ls.getLastOpenedCode(); code == code {
			typeYololCodeSSC(code)
		}
	case SSCAutodeleteHotkey.ID:
		win32.SendInput(win32.KeyUpInput(win32.KeycodeAlt))
		deleteAllLinesSSC()
	case SSCAutooverwriteHotkey.ID:
		win32.SendInput(win32.KeyUpInput(win32.KeycodeAlt))
		if code := ls.getLastOpenedCode(); code == code {
			overwriteYololCodeSSC(code)
		}
	case SSCAutocopyHotkey.ID:
		copyChipToClipboardSSC()
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

func typeYololCodeSSC(code string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		win32.SendString(line)
		nextLineSSC()
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

func overwriteYololCodeSSC(code string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		deleteLine()
		win32.SendString(line)
		nextLineSSC()
	}
}

func deleteAllLines() {
	for i := 0; i < 20; i++ {
		deleteLine()
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
}

func deleteAllLinesSSC() {
	for i := 0; i < 20; i++ {
		deleteLine()
		nextLineSSC()
	}
}

func copyChipToClipboard() {
	text := ""
	for i := 0; i < 20; i++ {
		text += copyLine() + "\n"
		win32.SendInput(win32.KeyDownInput(win32.KeycodeDown), win32.KeyUpInput(win32.KeycodeDown))
	}
	text = strings.TrimSuffix(text, "\n")
	clipboard.WriteAll(text)
}

func copyChipToClipboardSSC() {
	text := ""
	for i := 0; i < 20; i++ {
		text += copyLine() + "\n"
		nextLineSSC()
	}
	text = strings.TrimSuffix(text, "\n")
	clipboard.WriteAll(text)
}

func nextLineSSC() {
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyDownInputArrowDownSSC(), win32.KeyUpInputArrowDownSSC())
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyDownInputArrowDownSSC(), win32.KeyUpInputArrowDownSSC())
	time.Sleep(typeDelay)
}

func deleteLine() {
	win32.SendInput(win32.KeyDownInput(win32.KeycodeCtrl), win32.KeyDownInput('A'))
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyUpInput('A'), win32.KeyUpInput(win32.KeycodeCtrl))
	win32.SendInput(win32.KeyDownInput(win32.KeycodeBackspace), win32.KeyUpInput(win32.KeycodeBackspace))
	time.Sleep(typeDelay)
}

func copyLine() string {
	win32.SendInput(win32.KeyDownInput(win32.KeycodeCtrl), win32.KeyDownInput('A'))
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyUpInput('A'), win32.KeyUpInput(win32.KeycodeCtrl))
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyDownInput(win32.KeycodeCtrl), win32.KeyDownInput('C'))
	time.Sleep(typeDelay)
	win32.SendInput(win32.KeyUpInput('C'), win32.KeyUpInput(win32.KeycodeCtrl))
	time.Sleep(typeDelay)
	line, _ := clipboard.ReadAll()
	return line
}
