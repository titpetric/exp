package generic_test

import (
	"testing"
	"time"

	"github.com/tj/assert"

	"github.com/titpetric/exp/pkg/generic"
)

type TestMutexCounter struct {
	Name  string
	Value int
}

func TestMutex(t *testing.T) {
	value := &TestMutexCounter{}

	m := generic.Mutex[*TestMutexCounter]{}
	m.Set(value)
	m.Use(func(c *TestMutexCounter) {
		assert.Equal(t, 0, c.Value)
	})

	t.Run("use value", func(t *testing.T) {
		m.Use(func(c *TestMutexCounter) {
			c.Value = 10
		})
		m.Use(func(c *TestMutexCounter) {
			// value is modified
			assert.Equal(t, 10, c.Value)
		})
	})

	t.Run("use value copy", func(t *testing.T) {
		m.UseCopy(func(c *TestMutexCounter) {
			c.Value = 0
		})
		m.UseCopy(func(c *TestMutexCounter) {
			// value remains the same
			assert.Equal(t, 10, c.Value)
		})
	})
}

func BenchmarkMutexUse(b *testing.B) {
	m := generic.Mutex[TestMutexCounter]{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Use(func(d TestMutexCounter) {

			_ = d.Value // simulate read access
		})
	}
}

func BenchmarkMutexUseParallel(b *testing.B) {
	m := generic.Mutex[TestMutexCounter]{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Use(func(d TestMutexCounter) {
				time.Sleep(time.Millisecond)
				_ = d.Value
			})
		}
	})
}

func BenchmarkMutexUseCopy(b *testing.B) {
	m := generic.Mutex[TestMutexCounter]{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.UseCopy(func(d TestMutexCounter) {
			_ = d.Value // simulate read access
		})
	}
}

func BenchmarkMutexUseCopyParallel(b *testing.B) {
	m := generic.Mutex[TestMutexCounter]{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.UseCopy(func(d TestMutexCounter) {
				time.Sleep(time.Millisecond)
				_ = d.Value
			})
		}
	})
}
