// Log the panic under unix to the log file

//+build unix

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
