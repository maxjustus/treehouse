package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	root, err := Parse("", astLines())

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

func TestParseCreateMaterializedView(t *testing.T) {
	query := "CREATE MATERIALIZED ViEW \n my_table_or_view to some_table AS select * from z;"
	root, err := Parse(query, createQueryAstLines())

	assert.Nil(t, err)

	assert.Equal(t, "CreateQuery", root.Type)
	assert.Equal(t, "MateralizedViewToTable", root.Children[0].Type)
	assert.Equal(t, "some_table", root.Children[0].Value)
	assert.Equal(t, "TableIdentifier", root.Children[0].Children[0].Type)
	assert.Equal(t, "some_table", root.Children[0].Children[0].Value)
}

func TestTableIdentifiers(t *testing.T) {
	ast, _ := NewFromExplainLines("raw query", astLines())

	tableIdentifiers := ast.TableAndViewIdentifiers()

	assert.Equal(t, []string{"my_table_or_view"}, tableIdentifiers)
}

func TestFunctionCalls(t *testing.T) {
	ast, _ := NewFromExplainLines("raw query", astLines())

	calledFunctions := ast.FunctionCalls()

	assert.Equal(t, []string{"z"}, calledFunctions)
}
