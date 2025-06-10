# Package generic

```go
import (
	"github.com/titpetric/exp/pkg/generic"
}
```

## Types

```go
// List[T] is analogous with []T with utility functions.
//
// It's generally asumed you hold an exclusive reference to a list,
// and thus the following is not concurrency safe.
type List[T any] []T
```

```go
// Mutex protects any value T.
type Mutex[T any] struct {
	mu	sync.Mutex
	value	T
}
```

## Function symbols

- `func ListMap (l List[K], mapfn func(K) V) List[V]`
- `func NewList () List[T]`
- `func NewMutex (value T) *Mutex[T]`
- `func Pointer (value T) *T`
- `func UseMutex (m *Mutex[T], transform func(T) R) R`
- `func UseMutexCopy (m *Mutex[T], transform func(T) R) R`
- `func (*Mutex[T]) Copy () T`
- `func (*Mutex[T]) Set (value T)`
- `func (*Mutex[T]) Use (callback func(T))`
- `func (*Mutex[T]) UseCopy (callback func(T)) T`
- `func (List[T]) Filter (match func(T) bool) List[T]`
- `func (List[T]) Find (match func(T) bool) T`
- `func (List[T]) Get (index int) T`
- `func (List[T]) Value () []T`

### ListMap

ListMap converts a List[K] to a List[V] given a mapping function.

```go
func ListMap (l List[K], mapfn func(K) V) List[V]
```

### NewList

NewList creates a new List[T].

```go
func NewList () List[T]
```

### NewMutex

NewMutex will create a new mutex protected value.

```go
func NewMutex (value T) *Mutex[T]
```

### Pointer

Pointer will return a *T, referencing the value T.

```go
func Pointer (value T) *T
```

### UseMutex

UseMutex will take a Mutex[T] and safely invoke a tranform function.

```go
func UseMutex (m *Mutex[T], transform func(T) R) R
```

### UseMutexCopy

UseMutexCopy will copy a mutex-protected value, and invoke tranform.

```go
func UseMutexCopy (m *Mutex[T], transform func(T) R) R
```

### Copy

Copy will copy the value protected by mutex for exclusive use.

```go
func (*Mutex[T]) Copy () T
```

### Set

Set will replace the protected value.

```go
func (*Mutex[T]) Set (value T)
```

### Use

Use will run a callback over the stored value.

```go
func (*Mutex[T]) Use (callback func(T))
```

### UseCopy

Use copy will run a callback over a copy of the
stored value, returning the copy for exclusive use.

```go
func (*Mutex[T]) UseCopy (callback func(T)) T
```

### Filter

Filter traverses the list, and returns a new list with matching items.

```go
func (List[T]) Filter (match func(T) bool) List[T]
```

### Find

Find will return the first matching T element from the list, or the zero value of T.

```go
func (List[T]) Find (match func(T) bool) T
```

### Get

Get will return the T from index in the list, or the zero value of T.

```go
func (List[T]) Get (index int) T
```

### Value

Value function transforms the list to a native []T slice.

```go
func (List[T]) Value () []T
```


