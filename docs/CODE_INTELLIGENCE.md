# Code Intelligence

Juleson includes lightweight Go project inspection commands under `juleson dev`.

## CLI Commands

```bash
juleson dev deps [path]
juleson dev check-complexity [path]
```

Dependency analysis reports module relationships for Go projects. Complexity
analysis reports AST-based complexity metrics for source files.

## MCP Tools

The integrated MCP server exposes developer workflow tools such as `dev_build`,
`dev_test`, and `dev_check`. It does not expose a separate general-purpose code
intelligence surface.

## Limits

- The current implementation is focused on Go projects.
- Results depend on parseable source files and available module context.
- Analysis output should be treated as input for review, not as an automatic edit plan.
