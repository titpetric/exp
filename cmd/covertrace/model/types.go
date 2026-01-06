package model

type CoverageInfo struct {
	File        string  `json:"file"`
	RawFile     string  `json:"rawFile,omitempty"`
	PackageName string  `json:"packageName"`
	StartLine   int     `json:"start_line"`
	EndLine     int     `json:"end_line"`
	NumStmts    int     `json:"num_stmts"`
	NumCov      int     `json:"num_cov"`
	Symbol      string  `json:"symbol"`
	Receiver    string  `json:"receiver,omitempty"`
	Coverage    float64 `json:"coverage"`
}
