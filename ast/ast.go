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
type ExecQueryFunc func(query string) ([]map[string]interface{}, error)

var sqlCommentRegex = regexp.MustCompile(`(?m)^\s*\-\-.*$`)

func stripSqlComments(query string) string {
	// remove lines that start with --
	return sqlCommentRegex.ReplaceAllString(query, "")
}

var emptyQueryRegex = regexp.MustCompile(`^\s*$`)

func NewFromQuery(query string, execQueryFunc ExecQueryFunc) (*Ast, error) {
	if emptyQueryRegex.MatchString(query) {
		return &Ast{}, fmt.Errorf("query cannot be empty")
	}

	lines, error := execQueryFunc("explain ast " + stripSqlComments(query))
	if error != nil {
		return &Ast{}, error
	}

	linesStr := make([]string, len(lines))

	for i, line := range lines {
		linesStr[i] = line["explain"].(string)
	}

	return NewFromExplainLines(query, linesStr)
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

// TODO: explore adding caching or multithreading of querying for ASTs - if perf necessitates..?
func QueriesInTopologicalOrder(queries []string, execQueryFunc ExecQueryFunc) ([]string, error) {
	asts := make([]*Ast, 0, len(queries))

	for _, query := range queries {
		if emptyQueryRegex.MatchString(query) {
			continue
		}

		ast, err := NewFromQuery(query, execQueryFunc)
		if err != nil {
			return []string{}, err
		}
		asts = append(asts, ast)
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

func RunQueriesInTopologicalOrder(queries []string, execQueryFunc ExecQueryFunc) error {
	sortedQueries, err := QueriesInTopologicalOrder(queries, execQueryFunc)
	if err != nil {
		return err
	}

	for _, query := range sortedQueries {
		_, err := execQueryFunc(query)
		if err != nil {
			return err
		}
	}

	return nil
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
		createTableAndViewStatements := ast.CreateTableAndViewStatements()
		createFunctionStatements := ast.CreateFunctionStatements()

		selectColumnIdentifiers := ast.SelectColumnIdentifiers()

		createTableColumnDeclarations := ast.CreateTableColumnDeclarations()
		addColumnDeclarations := ast.AddColumnDeclarations()
		modifyColumnDeclarations := ast.ModifyColumnDeclarations()
		commentColumnIdentifiers := ast.CommentColumnIdentifiers()
		materializeColumnIdentifiers := ast.MaterializeColumnIdentifiers()
		renameColumnFromIdentifiers := ast.RenameColumnFromIdentifiers()
		renameColumnToIdentifiers := ast.RenameColumnToIdentifiers()
		createAndAddColumnDeclarations := append(createTableColumnDeclarations, addColumnDeclarations...)
		allOriginatingColumnIdentifiers := append(
			createAndAddColumnDeclarations,
			renameColumnToIdentifiers...)

		for _, candidate := range asts {
			if candidate != ast {
				ast.addDependentIfContainsAny(candidate, createTableAndViewStatements, candidate.TableAndViewIdentifiers())
				ast.addDependentIfContainsAny(candidate, createTableAndViewStatements, candidate.AlterQueryStatements())
				ast.addDependentIfContainsAny(candidate, createFunctionStatements, candidate.FunctionCalls())

				// wayyy slower now with all of this - TODO: optimize

				// Column relationships
				ast.addDependentIfContainsAny(candidate, allOriginatingColumnIdentifiers, candidate.SelectColumnIdentifiers())
				ast.addDependentIfContainsAny(candidate, allOriginatingColumnIdentifiers, candidate.ModifyColumnDeclarations())
				ast.addDependentIfContainsAny(candidate, allOriginatingColumnIdentifiers, candidate.CommentColumnIdentifiers())
				ast.addDependentIfContainsAny(candidate, allOriginatingColumnIdentifiers, candidate.MaterializeColumnIdentifiers())
				ast.addDependentIfContainsAny(candidate, allOriginatingColumnIdentifiers, candidate.RenameColumnFromIdentifiers())
				ast.addDependentIfContainsAny(candidate, createAndAddColumnDeclarations, candidate.RenameColumnToIdentifiers())

				// drop comes after add, modify, comment, materialize, rename
				dropColumnIdentifiers := candidate.DropOrClearColumnIdentifiers()
				ast.addDependentIfContainsAny(candidate, selectColumnIdentifiers, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, addColumnDeclarations, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, modifyColumnDeclarations, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, commentColumnIdentifiers, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, materializeColumnIdentifiers, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, renameColumnFromIdentifiers, dropColumnIdentifiers)
				ast.addDependentIfContainsAny(candidate, renameColumnToIdentifiers, dropColumnIdentifiers)
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

func descendentOf(node *AstNode, parentType string) bool {
	if node.Parent == nil {
		return false
	} else if node.Parent.Type == parentType {
		return true
	} else {
		return descendentOf(node.Parent, parentType)
	}
}

func parentIsType(node *AstNode, parentType string) bool {
	return node.Parent != nil && node.Parent.Type == parentType
}

func parentIsAlterColumn(node *AstNode, commandType string) bool {
	return parentIsType(node, "AlterCommand") && node.Parent.Value == commandType
}

func (a *Ast) alterColumnIdentifiers(modificationType string, columnType string) []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return parentIsAlterColumn(node, modificationType) && node.Type == columnType
	})
}

// TODO: document node types of interest
func (a *Ast) CreateTableColumnDeclarations() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return node.Type == "ColumnDeclaration" &&
			parentIsType(node, "ExpressionList") &&
			parentIsType(node.Parent, "Columns") &&
			parentIsType(node.Parent.Parent, "CreateQuery")
	})
}

func (a *Ast) AddColumnDeclarations() []string {
	return a.alterColumnIdentifiers("ADD_COLUMN", "ColumnDeclaration")
}

// alter table clear column also uses drop column as the command type
func (a *Ast) DropOrClearColumnIdentifiers() []string {
	return a.alterColumnIdentifiers("DROP_COLUMN", "Identifier")
}

func (a *Ast) ModifyColumnDeclarations() []string {
	return a.alterColumnIdentifiers("MODIFY_COLUMN", "ColumnDeclaration")
}

func (a *Ast) CommentColumnIdentifiers() []string {
	return a.alterColumnIdentifiers("COMMENT_COLUMN", "Identifier")
}

func (a *Ast) MaterializeColumnIdentifiers() []string {
	return a.alterColumnIdentifiers("MATERIALIZE_COLUMN", "Identifier")
}

func (a *Ast) RenameColumnFromIdentifiers() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return parentIsAlterColumn(node, "RENAME_COLUMN") &&
			node.Type == "Identifier" &&
			node.Parent.Children[0] == node
	})
}

func (a *Ast) RenameColumnToIdentifiers() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return parentIsAlterColumn(node, "RENAME_COLUMN") &&
			node.Type == "Identifier" &&
			node.Parent.Children[1] == node
	})
}

func (a *Ast) SelectColumnIdentifiers() []string {
	return a.ValuesForMatch(func(node *AstNode) bool {
		return node.Type == "Identifier" && descendentOf(node, "SelectQuery")
	})
}

func (a *Ast) TableAndViewIdentifiers() []string {
	return a.valuesForNodeType("TableIdentifier")
}

func (a *Ast) CreateTableAndViewStatements() []string {
	return a.valuesForNodeType("CreateQuery")
}

func (a *Ast) AlterQueryStatements() []string {
	return a.valuesForNodeType("AlterQuery")
}

func (a *Ast) FunctionCalls() []string {
	return a.valuesForNodeType("Function")
}

func (a *Ast) CreateFunctionStatements() []string {
	return a.valuesForNodeType("CreateFunctionQuery")
}
