# Package generic

```go
import (
	"github.com/titpetric/exp/pkg/generic"
}
```

## Types

```go
type List[T any] []T
```

```go
// Mutex protects any value T.
type Mutex[T any] struct {
	mu	sync.Mutex
	value	T
}
```

```go
type Value any
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

### NewMutex

NewMutex will create a new mutex protected value.

```go
func NewMutex (value T) *Mutex[T]
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

### ListMap

```go
func ListMap (l List[K], mapfn func(K) V) List[V]
```

### NewList

```go
func NewList () List[T]
```

### Pointer

```go
func Pointer (value T) *T
```

### Filter

```go
func (List[T]) Filter (match func(T) bool) List[T]
```

### Find

```go
func (List[T]) Find (match func(T) bool) T
```

### Get

```go
func (List[T]) Get (index int) T
```

### Value

```go
func (List[T]) Value () []T
```


