package rules

// RuleStatistics represents unified statistics for any linter rule.
type RuleStatistics struct {
	TotalSymbols      int            `yaml:"total_symbols"`
	ConsideredFuncs   int            `yaml:"considered_funcs,omitempty"`
	PassingFuncs      int            `yaml:"passing_funcs,omitempty"`
	ReportedIssues    int            `yaml:"reported_issues"`
	ArgOrderIssues    int            `yaml:"arg_order_issues,omitempty"`
	DuplicateIssues   int            `yaml:"duplicate_issues,omitempty"`
	ImportCollisions  int            `yaml:"import_collisions,omitempty"`
	IssueBreakdown    map[string]int `yaml:"issue_breakdown,omitempty"`
	ArgumentBreakdown []ArgCount     `yaml:"argument_breakdown,omitempty"`
}

// ArgCount represents the breakdown of functions by argument count.
type ArgCount struct {
	Arguments int `yaml:"arguments"`
	Functions int `yaml:"functions"`
	Valid     int `yaml:"valid"`
}
