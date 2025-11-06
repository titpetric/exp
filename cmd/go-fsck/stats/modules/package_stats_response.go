package modules

import "fmt"

type PackageStatsResponse struct {
	Files     int            `json:"files"`
	Functions int            `json:"functions"`
	Types     int            `json:"types"`
	Vars      int            `json:"vars"`
	Consts    int            `json:"consts"`
	Imports   map[string]int `json:"imports"` // import path → usage count
}

func NewPackageStatsResponse() PackageStatsResponse {
	return PackageStatsResponse{
		Imports: make(map[string]int),
	}
}

func (p PackageStatsResponse) String() string {
	return fmt.Sprintf(
		"Files %d, functions %d, types %d, vars %d, consts %d.",
		p.Files, p.Functions, p.Types, p.Vars, p.Consts,
	)
}

func (p *PackageStatsResponse) Merge(in PackageStatsResponse) {
	p.Files += in.Files
	p.Functions += in.Functions
	p.Types += in.Types
	p.Vars += in.Vars
	p.Consts += in.Consts
	for k, v := range in.Imports {
		p.Imports[k] += v
	}
}
