package rule

type RuleSyntaxVisitor struct {
	Raw    []byte
	Errors []error
	Rules  []*Rule

	// Set environment variable rules
	Environment map[string]string
}
