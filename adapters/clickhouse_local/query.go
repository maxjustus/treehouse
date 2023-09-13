package clickhouse_local

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// TODO: change this to start clickhouse local and then pipe into it.
// Most of the runtime comes from starting clickhouse-local.
func ExecQuery(query string) ([]map[string]interface{}, error) {
	// invoke clickhouse-local command below using shell
	out, err := exec.Command("sh", "-c", "clickhouse-local --query \""+query+"\" --output-format=\"JSONEachRow\"").Output()

	outLines := strings.Split(string(out), "\n")
	var parsedOutLines []map[string]interface{}

	for _, line := range outLines {
		if line != "" {
			parsedOutLines = append(parsedOutLines, ParseJSON(line))
		}
	}

	return parsedOutLines, err
}

func ParseJSON(s string) map[string]interface{} {
	var parsed map[string]interface{}
	json.Unmarshal([]byte(s), &parsed)
	return parsed
}
