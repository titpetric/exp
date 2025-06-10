# Package ./allocator

```go
import (
	"github.com/titpetric/exp/pkg/generic/allocator"
}
```

The allocator package serves as an optimization utility.

1. It uses sync.Pool to manage an in-memory cache of reusable types.
2. Provides a generic interface to take advantage of type safety.

Extensions may focus on measuring allocation pressure.

To use the allocator, typed code must provide a constructor.
With strongly typed code, a similar function is expected:

```go
func NewDocument() (*Document, error) {
}
```

To take advantage of the sync.Pool back allocator, you can
use it like so:

```go
repo := allocator.New[*Document](NewDocument)
value := repo.Get()
// doing things with value...
repo.Put(value)
```

The type must implement a Reset() function. The reliance
on repo.Put could be dropped with a specialized API that
uses runtime.SetFinalizer on the T.

## Types

```go
// Allocator holds a sync.Pool of objects of type T.
type Allocator[T Reseter] struct {
	pool sync.Pool
}
```

```go
// Reseter is the interface that types must implement to be managed by Allocator.
type Reseter interface {
	Reset()
}
```

## Function symbols

- `func New (newFunc func() T) *Allocator[T]`
- `func (*Allocator[T]) Get () T`
- `func (*Allocator[T]) Put (t T)`

### New

New creates an Allocator for type T using the provided constructor.

```go
func New (newFunc func() T) *Allocator[T]
```

### Get

Get retrieves an object from the internal pool.

```go
func (*Allocator[T]) Get () T
```

### Put

Put returns an object to the pool after resetting it.

```go
func (*Allocator[T]) Put (t T)
```


