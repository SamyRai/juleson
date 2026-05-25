# Code Intelligence

Code intelligence tools analyze Go source structure, symbols, imports, call
graphs, references, static checks, and complexity.

## MCP Tools

- `analyze_code_graph`: build a call graph and report dependencies, entry points,
  cycles, and graph metrics.
- `analyze_code_context`: inspect one file for package name, imports, symbols, and structure.
- `find_symbol_references`: search project code for definition, reference, and call sites.
- `run_static_analysis`: run static checks for Go code.
- `analyze_complexity`: report cyclomatic and cognitive complexity metrics.

## CLI Commands

The CLI exposes project-level analysis:

```bash
juleson analyze [project-path]
```

More granular code intelligence is currently exposed through MCP tools and
internal packages under `internal/codeintel`.

## Output Formats

Graph analysis can export formats such as Mermaid or DOT where supported by the
tool input. Prefer JSON output in MCP clients when a downstream tool needs to
consume the result.

## Limits

- The current implementation is focused on Go projects.
- Results depend on parseable source files and available module context.
- Static analysis output should be treated as input for review, not as an automatic edit plan.
