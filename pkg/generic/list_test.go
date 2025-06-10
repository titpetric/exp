package generic

import (
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestList(t *testing.T) {
	l := NewList[*time.Time]()
	l = append(l, Pointer(time.Now()))

	assert.Len(t, l, 1)
}

func TestList_Find(t *testing.T) {
	type KV struct {
		Key, Value string
	}

	l := List[*KV]([]*KV{
		&KV{"foo", "bar"},
		&KV{"xxx", "yyy"},
	})

	assert.Len(t, l, 2)
	assert.Equal(t, &KV{"foo", "bar"}, l.Find(func(v *KV) bool { return v.Key == "foo" }))
	assert.Nil(t, l.Find(func(v *KV) bool { return v.Key == "bar" }))
}

func TestList_Get(t *testing.T) {
	type KV struct {
		Key, Value string
	}

	l := List[*KV]([]*KV{
		&KV{"foo", "bar"},
		&KV{"xxx", "yyy"},
	})

	assert.Len(t, l, 2)
	assert.Equal(t, &KV{"foo", "bar"}, l.Get(0))
	assert.Equal(t, &KV{"xxx", "yyy"}, l.Get(1))
	assert.Nil(t, l.Get(2))
}

func TestList_Filter(t *testing.T) {
	type User struct {
		Name  string
		Admin bool
	}

	l := List[*User]([]*User{
		&User{"black", true},
		&User{"crazy", true},
		&User{"zobo", false},
	})

	isAdmin := func(v *User) bool {
		return v.Admin
	}
	isNotAdmin := func(v *User) bool {
		return v.Admin == false
	}

	assert.Len(t, l, 3)
	assert.Len(t, l.Filter(isAdmin), 2)
	assert.Len(t, l.Filter(isNotAdmin), 1)
}

func TestListMap(t *testing.T) {
	type User struct {
		Name  string
		Admin bool
	}

	l := List[*User]([]*User{
		&User{"black", true},
		&User{"crazy", true},
		&User{"zobo", false},
	})

	getNames := func(v *User) string {
		return v.Name
	}

	want := []string{"black", "crazy", "zobo"}

	got := ListMap[*User, string](l, getNames).Value()

	assert.Equal(t, want, got)
}
