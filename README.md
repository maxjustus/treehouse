# ch-ast-pal

A utility for parsing, introspecting and manipulating ClickHouse ASTs from the output of `explain ast`

The main idea here is that everyone tries to reinvent the wheel and make their own ClickHouse query parsers.
ClickHouse has functionality built in for outputting queries as ASTs via `explain ast #{query/create statement/etc}`.
This output is much simpler to parse than ClickHouse's SQL syntax and is likely to remain stable as ClickHouse's
functionality evolves.

Goals of this project:
- Provide an abstraction for parsing the output of `explain ast` with a simple entrypoint for obtaining that result from a ClickHouse client, clickHouse-local or chdb
- Provide a CLI for AST printing/manipulation that uses either clickhouse-local or chdb - chdb could be neat because it would be embedded.
  - could print AST as JSON, etc.
- Provide functionality for generating a directed acyclic graph of multiple queries such that any queries which create tables/views/materialized views/functions are
  parents of queries which depend on them.
  - Provide functionality for outputting the DAG as a sorted SQL file for execution against a database.
- Provide functionality for reformatting queries from their ASTs
- Provide functionality for diffing ASTs
  - Provide functionality for generating alter table statements based on those diffs.
- More???
