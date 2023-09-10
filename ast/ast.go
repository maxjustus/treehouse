package ast

type ExplainAstRow struct {
	Explain string
}

func NewFromQuery(query string, execQueryFunc func(query string) *[]ExplainAstRow) (*Ast, error) {
	lines := execQueryFunc("explain ast " + query)

	explainStrings := make([]string, len(*lines))

	for i, row := range *lines {
		explainStrings[i] = row.Explain
	}

	return NewFromExplainQuery(query, explainStrings)
}

func NewFromExplainQuery(query string, lines []string) (*Ast, error) {
	rootNode, err := Parse(lines)
	if err != nil {
		return &Ast{}, err
	}

	return &Ast{Root: rootNode, Query: query}, nil
}

type Ast struct {
	Root          *AstNode
	Query         string
	DependentAsts []*Ast
}

func (a *Ast) ValuesForMatch(matcher func(node *AstNode) bool) []string {
	var values []string

	a.Root.Walk(func(node *AstNode) {
		if matcher(node) {
			values = append(values, node.Value)
		}
	})

	return values
}

func (a *Ast) TableIdentifiers() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return node.Type == "TableIdentifier"
	})
}

func (a *Ast) CalledFunctions() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return node.Type == "Function"
	})
}
