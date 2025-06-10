package generic

import (
	"time"
)

// Pointer will return a *T, referencing the value T.
func Pointer[T any](value T) *T {
	return &value
}

var _ *time.Time = Pointer(time.Now())
