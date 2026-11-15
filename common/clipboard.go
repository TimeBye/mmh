package common

import "github.com/atotto/clipboard"

// ClipboardWriteAll writes text to the system clipboard
// macOS uses pbcopy, Linux uses xsel/xclip/wl-copy, Windows uses the system API
func ClipboardWriteAll(text string) error {
	return clipboard.WriteAll(text)
}
