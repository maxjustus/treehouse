package ast

import (
	"testing"

	"github.com/k0kubun/pp"
	"github.com/maxjustus/treehouse/adapters/clickhouse_local"
	"github.com/stretchr/testify/assert"
)

func TestPopulateDependencyGraph(t *testing.T) {
	ast1, _ := NewFromExplainLines("query func", astLines())
	ast2, _ := NewFromExplainLines("create func", createFunctionAstLines())
	ast3, _ := NewFromExplainLines("create table/view/materialized view", createQueryAstLines())
	// TODO: expand to all types of creation and references. views, materialized views, tables.

	populateDependencyGraph(ast1, ast2, ast3)

	// TODO: relationship is many to many. so multiple parents are possible.
	assert.Equal(t, ast1.ParentAsts, []*Ast{ast2, ast3})
	pp.Println(ast1)
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
	t1 := "create table t1 (z Int64) engine=MergeTree order by z"
	a1 := "alter table t1 add column b UInt8"
	f1 := "create function f1 as () -> true"
	v1 := "create view v1 as select *, f1() as y from t1"
	v2 := "create view v2 as select * from v1"
	mv1_dst_t := "create table mv1_dest_t (z Int64, b UInt8) engine=MergeTree order by z "
	mv1 := "create materialized view mv1 to mv1_dest_1 as (select *, f1() as r from t1 as some_alias join v1 as other using z join v2 as other2 using z)"

	// TODO: make sure this API for entry point is decent and consistent
	out, err := QueriesInTopologicalOrder([]string{
		a1,
		mv1,
		mv1_dst_t,
		t1,
		v1,
		v2,
		f1,
	}, clickhouse_local.ExecQuery)

	assert.NoError(t, err)

	expected := []string{
		mv1_dst_t,
		t1,
		f1,
		a1,
		v1,
		v2,
		mv1,
	}

	assert.Equal(t, expected, out)
}

func TestColumnQueryiesInTopologicalOrder(t *testing.T) {
	q1 := "create table t1 (z Int64) engine=MergeTree order by z"
	q2 := "alter table t1 add column b UInt8"
	q3 := "alter table t1 add column c UInt8"
	q4 := "alter table t1 drop column c"
	q5 := "alter table t1 modify column b UInt16"
	q6 := "alter table t1 rename column b to whatever"
	q7 := "select whatever from t1"
	q8 := "alter table t1 comment column whatever 'hello'"
	q9 := "alter table t1 modify column whatever UInt8"
	q10 := "alter table t1 materialize column whatever"
	q11 := "create view v1 as select whatever from t1"
	q12 := "alter table t1 drop column whatever"
	q13 := "alter table t1 modify column whatever UInt32"
	q14 := "create view v2 as select c from t1"

	// TODO: make sure this API for entry point is decent and consistent
	out, err := QueriesInTopologicalOrder([]string{
		q12,
		q5,
		q10,
		q6,
		q4,
		q13,
		q8,
		q9,
		q2,
		q7,
		q3,
		q14,
		q11,
		q1,
	}, clickhouse_local.ExecQuery)

	assert.NoError(t, err)

	expected := []string{
		q1,
		q2,
		q3,
		q5,
		q6,
		q14,
		q10,
		q13,
		q8,
		q9,
		q7,
		q11,
		q4,
		q12,
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
		}, clickhouse_local.ExecQuery)
	}
}
