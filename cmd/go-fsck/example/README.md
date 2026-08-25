# Package example

```go
import (
	"github.com/titpetric/exp/cmd/go-fsck/example"
}
```
Example package doc.

## Types

<details>
<summary><code>type Allocator</code></summary>

```go
// Allocator holds a sync.Pool of objects of type T.
type Allocator[T Reseter] struct {
	pool sync.Pool
}
```

</details>

<details>
<summary><code>type Body</code></summary>

```go
// Body represends a decoded body
type Body struct {
	Name string
}
```

</details>

<details>
<summary><code>type File</code></summary>

```go
// File represents a filename
type File string
```

</details>

<details>
<summary><code>type Reseter</code></summary>

```go
// Reseter is the interface that types must implement to be managed by Allocator.
type Reseter interface {
	Reset()
}
```

</details>

## Consts

<details>
<summary><code>const E_WARNING</code></summary>

```go
// Const comment
const E_WARNING = "warning"	// const line comment
```

</details>

## Function symbols

- `func GlobalFunc () error`
- `func New (newFunc func() T) *Allocator[T]`
- `func (*Allocator[T]) Get () T`
- `func (*Allocator[T]) Put (t T)`

### GlobalFunc

Global func comment

```go
func GlobalFunc () error
```

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


