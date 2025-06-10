package generic

import (
	"time"
)

type Value any

func Pointer[T Value](value T) *T {
	return &value
}

var _ *time.Time = Pointer(time.Now())
