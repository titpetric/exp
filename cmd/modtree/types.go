package main

// xterm256 color helpers
const (
	colorReset   = "\033[0m"
	colorDim     = "\033[2m"
	colorBold    = "\033[1m"
	colorOrange  = "\033[38;5;208m" // outdated marker
	colorGreen   = "\033[38;5;114m" // up-to-date version
	colorBlue    = "\033[38;5;111m" // module names
	colorMagenta = "\033[38;5;183m" // keys
	colorGray    = "\033[38;5;245m" // punctuation/structure
	colorWhite   = "\033[38;5;255m" // values
	colorCyan    = "\033[38;5;116m" // tags
	colorYellow  = "\033[38;5;222m" // git status warnings
	colorRed     = "\033[38;5;203m" // git status alerts
)

type gitStatus struct {
	Unpushed   int
	Modified   int
	Insertions int
	Deletions  int
}

type moduleInfo struct {
	Name       string
	Latest     string
	Ahead      int
	Git        string
	GitMsgs    []string
	Uses       []string
	UsedBy     []string
}

type requireInfo struct {
	path    string
	version string
}
