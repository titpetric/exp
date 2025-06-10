package generic

import (
	"sync"

	clone "github.com/huandu/go-clone/generic"
)

// Mutex protects any value T.
type Mutex[T any] struct {
	mu    sync.Mutex
	value T
}

// NewMutex will create a new mutex protected value.
func NewMutex[T any](value T) *Mutex[T] {
	return &Mutex[T]{
		value: value,
	}
}

// Set will replace the protected value.
func (m *Mutex[T]) Set(value T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = value
}

// Use will run a callback over the stored value.
func (m *Mutex[T]) Use(callback func(T)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	callback(m.value)
}

// Use copy will run a callback over a copy of the
// stored value, returning the copy for exclusive use.
func (m *Mutex[T]) UseCopy(callback func(T)) T {
	m.mu.Lock()
	c := clone.Clone(m.value)
	m.mu.Unlock()

	callback(c)
	return c
}

// Copy will copy the value protected by mutex for exclusive use.
func (m *Mutex[T]) Copy() T {
	m.mu.Lock()
	defer m.mu.Unlock()

	return clone.Clone(m.value)
}

// UseMutex will take a Mutex[T] and safely invoke a tranform function.
func UseMutex[T any, R any](m *Mutex[T], transform func(T) R) R {
	m.mu.Lock()
	defer m.mu.Unlock()

	return transform(m.value)
}

// UseMutexCopy will copy a mutex-protected value, and invoke tranform.
func UseMutexCopy[T any, R any](m *Mutex[T], transform func(T) R) R {
	m.mu.Lock()
	c := clone.Clone(m.value)
	m.mu.Unlock()

	return transform(c)
}
