package clickhouse_local

import (
	"testing"

	"github.com/maxjustus/treehouse/ast"
	"github.com/stretchr/testify/assert"
)

func AstForQuery(query string) (*ast.Ast, error) {
	return ast.NewFromQuery(query, ExecQuery)
}

func TestAstForQuery(t *testing.T) {
	r, err := AstForQuery("select * from x")

	assert.NoError(t, err)

	assert.Equal(t, "SelectWithUnionQuery", r.Root.Type)
}
