package allocator_test

import (
	"runtime"
	"testing"

	"github.com/tj/assert"

	"github.com/titpetric/exp/pkg/generic/allocator"
)

func TestAllocator(t *testing.T) {
	alloc := allocator.New[*Document](NewDocument)

	// Get an object.
	obj := alloc.Get()
	assert.Len(t, obj.Tags, 0)
}

func BenchmarkConstructor(b *testing.B) {
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			obj := NewDocument()
			assert.NotNil(b, obj)
			i++

			if i&0xffff == 0 {
				runtime.GC()
			}
		}
	})
}

func BenchmarkAllocator(b *testing.B) {
	alloc := allocator.New[*Document](NewDocument)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			obj := alloc.Get()
			assert.NotNil(b, obj)
			alloc.Put(obj)
			i++

			if i&0xffff == 0 {
				runtime.GC()
			}
		}
	})
}
