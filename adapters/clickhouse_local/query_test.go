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
	r, err := AstForQuery("create table test (a Int32) Engine=MergeTree order by a")

	assert.NoError(t, err)

	assert.Equal(t, "CreateQuery", r.Root.Type)
}
