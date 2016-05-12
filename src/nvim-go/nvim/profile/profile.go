package profile

import (
	"log"
	"time"
)

// Profile measurement of the time it took to any func and output log file.
// Usage: defer nvim.Profile(time.Now(), "func name")
func Start(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s: %s\n", name, elapsed)
}
