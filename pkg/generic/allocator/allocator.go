package allocator

import (
	"sync"
)

// Reseter is the interface that types must implement to be managed by Allocator.
type Reseter interface {
	Reset()
}

// Allocator holds a sync.Pool of objects of type T.
type Allocator[T Reseter] struct {
	pool sync.Pool
}

// New creates an Allocator for type T using the provided constructor.
func New[T Reseter](newFunc func() T) *Allocator[T] {
	return &Allocator[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get retrieves an object from the internal pool.
func (a *Allocator[T]) Get() T {
	return a.pool.Get().(T) //nolint
}

// Put returns an object to the pool after resetting it.
func (a *Allocator[T]) Put(t T) {
	t.Reset()
	a.pool.Put(t)
}
