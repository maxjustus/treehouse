package ast

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func queryClickHouseLocal(query string) ([]string, error) {
	// invoke clickhouse-local command below using shell
	out, err := exec.Command("sh", "-c", "clickhouse-local --query \"explain ast "+query+"\"").Output()
	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(string(out), "\n")
	return lines, nil
}

func TestPopulateDependencyGraph(t *testing.T) {
	ast1, _ := NewFromExplainLines("query func", astLines())
	ast2, _ := NewFromExplainLines("create func", createFunctionAstLines())
	ast3, _ := NewFromExplainLines("create table/view/materialized view", createQueryAstLines())
	// TODO: expand to all types of creation and references. views, materialized views, tables.

	populateDependencyGraph(ast1, ast2, ast3)

	// TODO: relationship is many to many. so multiple parents are possible.
	assert.Equal(t, ast1.ParentAsts, []*Ast{ast2, ast3})
	assert.Equal(t, ast2.DependentAsts[0], ast1)
	assert.Equal(t, ast3.DependentAsts[0], ast1)
}

func TestSortedQueriesFromDependencyGraph(t *testing.T) {
	ast1, _ := NewFromExplainLines("query func", astLines())
	ast2, _ := NewFromExplainLines("create func", createFunctionAstLines())
	ast3, _ := NewFromExplainLines("create table/view/materialized view", createQueryAstLines())

	out, err := PopulateAndSort(ast1, ast2, ast3)

	assert.NoError(t, err)

	outStr := []string{}
	for _, ast := range out {
		outStr = append(outStr, ast.Query)
	}

	expected := []string{ast2.Query, ast3.Query, ast1.Query}
	assert.Equal(t, expected, outStr)
}

func TestQueriesInTopologicalOrder(t *testing.T) {
	// TODO: put actual queries
	t1 := "create table t1 (z Int64) engine=MergeTree order by z"
	f1 := "create function f1 as () -> true"
	v1 := "create view v1 as select *, f1() as y from t1"
	v2 := "create view v2 as select * from v1"
	mv1_dst_t := "create table mv1_dest_t (z Int64, b UInt8) engine=MergeTree order by z "
	mv1 := "create materialized view mv1 to mv1_dest_1 as (select *, f1() as r from t1 as some_alias join v1 as other using z join v2 as other2 using z)"

	// TODO: make sure this API for entry point is decent and consistent
	out, err := QueriesInTopologicalOrder([]string{
		mv1,
		mv1_dst_t,
		t1,
		v1,
		v2,
		f1,
	}, queryClickHouseLocal)

	assert.NoError(t, err)

	expected := []string{
		mv1_dst_t,
		t1,
		f1,
		v1,
		v2,
		mv1,
	}

	assert.Equal(t, expected, out)
}

func BenchmarkQueriesInTopologicalOrder(b *testing.B) {
	t1 := "create table t1 (z Int64) engine=MergeTree order by z"
	f1 := "create function f1 as () -> true"
	v1 := "create view v1 as select *, f1() as y from t1"
	v2 := "create view v2 as select * from v1"
	mv1_dst_t := "create table mv1_dest_t (z Int64, b UInt8) engine=MergeTree order by z "
	mv1 := "create materialized view mv1 to mv1_dest_1 as (select *, f1() as r from t1 as some_alias join v1 as other using z join v2 as other2 using z)"

	for i := 0; i < b.N; i++ {
		QueriesInTopologicalOrder([]string{
			mv1,
			mv1_dst_t,
			t1,
			v1,
			v2,
			f1,
		}, queryClickHouseLocal)
	}
}
