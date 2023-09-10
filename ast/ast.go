package ast

import (
	"container/list"
	"fmt"
	"regexp"
)

type Ast struct {
	Root          *AstNode
	Query         string
	ParentAsts    []*Ast
	DependentAsts []*Ast
}

// TODO: return err instead of panicking
type ExecQueryFunc func(query string) ([]string, error)

var sqlCommentRegex = regexp.MustCompile(`(?m)^\s*\-\-.*$`)

func stripSqlComments(query string) string {
	// remove lines that start with --
	return sqlCommentRegex.ReplaceAllString(query, "")
}

func NewFromQuery(query string, execQueryFunc ExecQueryFunc) (*Ast, error) {
	lines, error := execQueryFunc("explain ast " + stripSqlComments(query))
	if error != nil {
		return &Ast{}, error
	}

	return NewFromExplainLines(query, lines)
}

func NewFromExplainLines(query string, lines []string) (*Ast, error) {
	rootNode, err := Parse(query, lines)
	if err != nil {
		return &Ast{}, err
	}

	return &Ast{Root: rootNode, Query: query}, nil
}

func containsAny(a []string, b []string) bool {
	for _, valA := range a {
		for _, valB := range b {
			if valA == valB {
				return true
			}
		}
	}
	return false
}

// TODO: add tests - maybe caching or multithreading of querying for ASTs - if perf necessitates..?
func QueriesInTopologicalOrder(queries []string, execQueryFunc ExecQueryFunc) ([]string, error) {
	asts := make([]*Ast, len(queries))

	for i, query := range queries {
		ast, err := NewFromQuery(query, execQueryFunc)
		if err != nil {
			return []string{}, err
		}
		asts[i] = ast
	}

	sortedAsts, err := PopulateAndSort(asts...)
	if err != nil {
		return []string{}, err
	}

	sortedQueries := make([]string, len(sortedAsts))
	for i, ast := range sortedAsts {
		sortedQueries[i] = ast.Query
	}

	return sortedQueries, nil
}

func PopulateAndSort(asts ...*Ast) ([]*Ast, error) {
	populateDependencyGraph(asts...)

	queue := list.New()
	output := make([]*Ast, 0)
	nodeDegrees := make(map[*Ast]int)

	for _, ast := range asts {
		nodeDegrees[ast] = len(ast.ParentAsts)

		if len(ast.ParentAsts) == 0 {
			queue.PushBack(ast)
		}
	}

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		ast := element.Value.(*Ast)

		output = append(output, ast)

		for _, dependentAst := range ast.DependentAsts {
			nodeDegrees[dependentAst]--

			if nodeDegrees[dependentAst] <= 0 {
				queue.PushBack(dependentAst)
			}
		}
	}

	if len(output) != len(asts) {
		return []*Ast{}, fmt.Errorf("circular dependency detected")
	}

	return output, nil
}

func populateDependencyGraph(asts ...*Ast) {
	for _, ast := range asts {
		for _, dependentAst := range asts {
			if dependentAst != ast {
				ast.addDependentIfContainsAny(dependentAst, ast.CreateTableAndViewStatements(), dependentAst.TableAndViewIdentifiers())
				ast.addDependentIfContainsAny(dependentAst, ast.CreateFunctionStatements(), dependentAst.FunctionCalls())
			}
		}
	}
}

func (ast *Ast) addDependentIfContainsAny(dependentAst *Ast, a []string, b []string) {
	if containsAny(a, b) {
		ast.addDependent(dependentAst)
	}
}

func (ast *Ast) addDependent(dependentAst *Ast) {
	ast.DependentAsts = append(ast.DependentAsts, dependentAst)
	dependentAst.ParentAsts = append(dependentAst.ParentAsts, ast)
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

func (a *Ast) valuesForNodeType(nodeType string) []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return node.Type == nodeType
	})
}

// TODO: document node types of interest
func (a *Ast) TableAndViewIdentifiers() []string {
	return a.valuesForNodeType("TableIdentifier")
}

func (a *Ast) CreateTableAndViewStatements() []string {
	return a.valuesForNodeType("CreateQuery")
}

func (a *Ast) FunctionCalls() []string {
	return a.valuesForNodeType("Function")
}

func (a *Ast) CreateFunctionStatements() []string {
	return a.valuesForNodeType("CreateFunctionQuery")
}
