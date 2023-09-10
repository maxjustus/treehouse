package clickhouse_local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAstForQuery(t *testing.T) {
	r, err := AstForQuery("select * from x")

	assert.NoError(t, err)

	assert.Equal(t, "SelectWithUnionQuery", r.Root.Type)
}
