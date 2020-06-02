package util

import (
	"log"
	"time"
)

var ShowTimeTrack = false

func TimeTrack(start time.Time, name string) {
	if !ShowTimeTrack {
		return
	}
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
