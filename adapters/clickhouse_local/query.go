package clickhouse_local

import (
	"os/exec"
	"strings"
)

// TODO: change this to start clickhouse local and then pipe into it.
// Most of the runtime comes from starting clickhouse-local.
func ExecQuery(query string) ([]string, error) {
	// invoke clickhouse-local command below using shell
	out, err := exec.Command("sh", "-c", "clickhouse-local --query \""+query+"\"").Output()

	if err != nil {
		return []string{}, err
	} else {
		return strings.Split(string(out), "\n"), err
	}
}
