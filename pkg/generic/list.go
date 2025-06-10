package generic

type List[T any] []T

func NewList[T any]() List[T] {
	return List[T]{}
}

func (l List[T]) Filter(match func(T) bool) List[T] {
	var result List[T]
	for _, v := range l {
		if match(v) {
			result = append(result, v)
		}
	}
	return result
}

func (l List[T]) Find(match func(T) bool) T {
	var result T
	for _, v := range l {
		if match(v) {
			return v
		}
	}
	return result
}

func (l List[T]) Get(index int) T {
	var result T
	if len(l) > index {
		return l[index]
	}
	return result
}

func (l List[T]) Value() []T {
	return []T(l)
}

func ListMap[K any, V any](l List[K], mapfn func(K) V) List[V] {
	var result List[V]
	for _, v := range l {
		result = append(result, mapfn(v))
	}
	return result
}
