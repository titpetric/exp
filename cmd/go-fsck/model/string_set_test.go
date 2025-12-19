package model

import (
	"testing"
)

func TestStringSet(t *testing.T) {
	s := NewStringSet()
	s.Add("test.go", `"net/http"`, `"github.com/huandu/go-clone"`, `modelName "github.com/package/model"`)
	imports := s.All()
	m, errs := s.Map(imports)
	t.Log(m, errs)
}
