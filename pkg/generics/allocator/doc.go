// The allocator package serves as an optimization utility.
//
// 1. It uses sync.Pool to manage an in-memory cache of reusable types.
// 2. Provides a generic interface to take advantage of type safety.
//
// Extensions may focus on measuring allocation pressure.
//
// To use the allocator, typed code must provide a constructor.
// With strongly typed code, a similar function is expected:
//
// ```go
// func NewDocument() (*Document, error) {
// }
// ```
//
// To take advantage of the sync.Pool back allocator, you can
// use it like so:
//
// ```go
// repo := allocator.New[*Document](NewDocument)
// value := repo.Get()
// // doing things with value...
// repo.Put(value)
// ```
//
// The type must implement a Reset() function. The reliance
// on repo.Put could be dropped with a specialized API that
// uses runtime.SetFinalizer on the T.
package allocator
