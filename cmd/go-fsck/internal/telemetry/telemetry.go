package telemetry

import (
	"fmt"
	"runtime"
	"time"
)

func MemoryUse() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func Start(name string) Span {
	return Span{
		Name:  name,
		Start: time.Now(),
		Alloc: MemoryUse(),
	}
}

type Span struct {
	Name  string
	Start time.Time
	Alloc uint64
}

func (s Span) String() string {
	seconds := time.Since(s.Start).Seconds()
	alloc := MemoryUse()
	return fmt.Sprintf("%s: took %.2f seconds, memory before: %dMB, memory after %dMB", s.Name, seconds, s.Alloc/1024/1024, alloc/1024/1024)
}

func (s Span) End() {
	// DEBUG: print span details
	// fmt.Println(s)
}
