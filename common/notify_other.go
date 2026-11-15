//go:build !darwin

package common

import osexec "os/exec"

// Notify shows a desktop notification through the noti command
func Notify(title, message string) error {
	return osexec.Command("noti", "-t", title, "-m", message).Run()
}
