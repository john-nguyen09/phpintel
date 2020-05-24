// Log the panic under linux to the log file

//+build linux

package util

import (
	"log"
	"os"
	"syscall"
)

// RedirectStderr to the file passed in
func RedirectStderr(f *os.File) {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}
