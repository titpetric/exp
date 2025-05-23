package lint

import "github.com/titpetric/exp/cmd/schema-gen/model"

type Linter interface {
	GetName() string
	Do(*options, *model.PackageInfo) *LintError
}

type LinterFunc func(*options, *model.PackageInfo) *LintError

func NewLinter(name string, doFunc LinterFunc) Linter {
	return &linter{
		name:   name,
		doFunc: doFunc,
	}
}

type linter struct {
	name   string
	doFunc LinterFunc
}

func (l *linter) GetName() string {
	return l.name
}

func (l *linter) Do(cfg *options, in *model.PackageInfo) *LintError {
	return l.doFunc(cfg, in)
}
