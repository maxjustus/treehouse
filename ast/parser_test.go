package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func astLines() []string {
	return []string{
		"SelectWithUnionQuery (children 1)",
		" ExpressionList (children 1)",
		"  SelectQuery (children 2)",
		"   ExpressionList (children 1)",
		"    Asterisk",
		"    Identifier z (alias n)",
		"    Function z (alias r) (children 1)",
		"     ExpressionList",
		"   TablesInSelectQuery (children 1)",
		"    TablesInSelectQueryElement (children 1)",
		"     TableExpression (children 1)",
		"      TableIdentifier x",
	}
}

func TestParse(t *testing.T) {
	root, err := Parse(astLines())

	assert.Nil(t, err)

	// TODO: add something like a test helper function where I can assert on paths of types and optional fields.
	// like assertTypeBranch("SelectWithUnionQuery", "ExpressionList", "SelectQuery", "ExpressionList", "Asterisk")
	assert.Equal(t, "SelectWithUnionQuery", root.Type)
	assert.Equal(t, "ExpressionList", root.Children[0].Type)
	assert.Equal(t, "SelectQuery", root.Children[0].Children[0].Type)
	assert.Equal(t, "ExpressionList", root.Children[0].Children[0].Children[0].Type)
	assert.Equal(t, "Asterisk", root.Children[0].Children[0].Children[0].Children[0].Type)
	assert.Equal(t, "Identifier", root.Children[0].Children[0].Children[0].Children[1].Type)
	assert.Equal(t, "Function", root.Children[0].Children[0].Children[0].Children[2].Type)
	assert.Equal(t, "TablesInSelectQuery", root.Children[0].Children[0].Children[1].Type)
	assert.Equal(t, "TablesInSelectQueryElement", root.Children[0].Children[0].Children[1].Children[0].Type)
	assert.Equal(t, "TableExpression", root.Children[0].Children[0].Children[1].Children[0].Children[0].Type)
	assert.Equal(t, "TableIdentifier", root.Children[0].Children[0].Children[1].Children[0].Children[0].Children[0].Type)
}

func TestTableIdentifiers(t *testing.T) {
	ast, _ := NewFromExplainQuery("raw query", astLines())

	tableIdentifiers := ast.TableIdentifiers()

	assert.Equal(t, []string{"x"}, tableIdentifiers)
}

func TestCalledFunctions(t *testing.T) {
	ast, _ := NewFromExplainQuery("raw query", astLines())

	calledFunctions := ast.CalledFunctions()

	assert.Equal(t, []string{"z"}, calledFunctions)
}
