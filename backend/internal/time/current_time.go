package time

import (
	"time"
)

func CurrentTime() string {
	// Store timestamps in UTC to avoid client-side timezone skew.
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}
