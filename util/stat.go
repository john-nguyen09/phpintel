package util

import (
	"log"
	"runtime"
	"time"
)

// ShowTimeTrack controls if TimeTrack should be logged
var ShowTimeTrack = false

// TimeTrack tracks the time
// Example: `defer util.TimeTrack(time.Now(), "An example")`
func TimeTrack(start time.Time, name string) {
	if !ShowTimeTrack {
		return
	}
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	log.Printf("TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("Sys = %v MiB", bToMb(m.Sys))
	log.Printf("NumGC = %v", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
