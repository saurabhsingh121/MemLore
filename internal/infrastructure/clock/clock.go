package clock

import (
	"time"

	"github.com/memlore/memlore/internal/application/ports"
)

// SystemClock returns UTC wall-clock time.
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// FixedClock is a test double with a constant instant.
type FixedClock struct {
	Instant time.Time
}

func (c FixedClock) Now() time.Time {
	return c.Instant.UTC()
}

var (
	_ ports.Clock = SystemClock{}
	_ ports.Clock = FixedClock{}
)
