package model

import (
	"testing"
)

func TestStringSet(t *testing.T) {
	s := NewStringSet()
	s.Add("test.go", `"net/http"`, `"github.com/huandu/go-clone"`, `modelName "github.com/package/model"`)
	t.Log(s.Map())
}
