package generic

// List[T] is analogous with []T with utility functions.
//
// It's generally asumed you hold an exclusive reference to a list,
// and thus the following is not concurrency safe.
type List[T any] []T

// NewList creates a new List[T].
func NewList[T any]() List[T] {
	return List[T]{}
}

// Filter traverses the list, and returns a new list with matching items.
func (l List[T]) Filter(match func(T) bool) List[T] {
	var result List[T]
	for _, v := range l {
		if match(v) {
			result = append(result, v)
		}
	}
	return result
}

// Find will return the first matching T element from the list, or the zero value of T.
func (l List[T]) Find(match func(T) bool) T {
	var result T
	for _, v := range l {
		if match(v) {
			return v
		}
	}
	return result
}

// Get will return the T from index in the list, or the zero value of T.
func (l List[T]) Get(index int) T {
	var result T
	if len(l) > index {
		return l[index]
	}
	return result
}

// Value function transforms the list to a native []T slice.
func (l List[T]) Value() []T {
	return []T(l)
}

// ListMap converts a List[K] to a List[V] given a mapping function.
func ListMap[K any, V any](l List[K], mapfn func(K) V) List[V] {
	var result List[V]
	for _, v := range l {
		result = append(result, mapfn(v))
	}
	return result
}
