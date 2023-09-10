package clickhouse_local

import (
	"os/exec"
	"strings"

	"github.com/maxjustus/ch-ast-pal/ast"
)

func AstForQuery(query string) (*ast.Ast, error) {
	return ast.NewFromExplainQuery(query, explainAstRowsForQuery(query))
}

func explainAstRowsForQuery(query string) []string {
	// invoke clickhouse-local command below using shell
	out, err := exec.Command("sh", "-c", "clickhouse-local --query \"explain ast "+query+"\"").Output()
	if err != nil {
		panic(err)
	}

	return strings.Split(string(out), "\n")
}
