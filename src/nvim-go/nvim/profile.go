package nvim

import (
	"log"
	"time"
)

func Profile(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s: %s", name, elapsed)
}
