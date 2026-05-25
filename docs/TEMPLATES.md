# Templates

Templates describe repeatable automation tasks that can be listed, inspected,
searched, created, and executed through the CLI or MCP server.

## Built-In Templates

The built-in templates live under `templates/builtin`.

| Category | Templates |
| --- | --- |
| `reorganization` | `modular-restructure`, `layered-architecture`, `microservices-split` |
| `testing` | `test-generation`, `test-coverage-improvement`, `integration-tests` |
| `refactoring` | `code-cleanup`, `dependency-update`, `api-modernization`, `code-quality-improvement` |
| `documentation` | `api-documentation`, `readme-generation` |

Template metadata is indexed in `templates/registry/registry.yaml`.

## CLI Usage

```bash
juleson template list
juleson template list testing
juleson template show test-generation
juleson template search coverage
juleson execute template test-generation ./path/to/project
juleson execute template-with-params api-documentation ./path/to/project format=markdown
```

## Custom Templates

Create a custom template:

```bash
juleson template create api-versioning refactoring "Add API versioning"
```

Custom template support is controlled by:

```yaml
templates:
  custom_path: "./templates/custom"
  enable_custom: true
```

## Template Shape

Template YAML contains metadata, task definitions, validation rules, and output
settings. Keep templates specific to one type of change, include clear
prerequisites, and avoid hidden side effects.
