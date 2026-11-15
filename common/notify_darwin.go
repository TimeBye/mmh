//go:build darwin

package common

import (
	"os/exec"
	"strings"
)

// Notify shows a desktop notification through the macOS notification center
func Notify(title, message string) error {
	script := "display notification " + appleScriptString(message) + " with title " + appleScriptString(title)
	return exec.Command("osascript", "-e", script).Run()
}

// appleScriptString escapes a string for use as an AppleScript string literal
func appleScriptString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}
