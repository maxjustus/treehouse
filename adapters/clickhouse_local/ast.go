package clickhouse_local

import (
	"os/exec"
	"strings"

	"github.com/maxjustus/ch-ast-pal/ast"
)

func AstForQuery(query string) (*ast.Ast, error) {
	lines, err := explainAstRowsForQuery(query)
	if err != nil {
		return &ast.Ast{}, err
	}

	return ast.NewFromExplainLines(query, lines)
}

// TODO: change this to start clickhouse local and then pipe into it.
// Most of the runtime comes from starting clickhouse-local.
func explainAstRowsForQuery(query string) ([]string, error) {
	// invoke clickhouse-local command below using shell
	out, err := exec.Command("sh", "-c", "clickhouse-local --query \"explain ast "+query+"\"").Output()
	if err != nil {
		return []string{}, err
	} else {
		return strings.Split(string(out), "\n"), err
	}
}
