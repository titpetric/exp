package generic

import (
	"time"
)

type Value any

func Pointer[T Value](val T) *T {
	return &val
}

var _ *time.Time = Pointer(time.Now())
