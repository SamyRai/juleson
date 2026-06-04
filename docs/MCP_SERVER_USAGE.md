# MCP Server Usage

Juleson serves MCP over stdio with the official Go MCP SDK.

## Start The Server

```bash
juleson mcp serve
```

The short alias works the same way:

```bash
jsn mcp serve
```

Use `--version` for a smoke check that does not start the stdio loop:

```bash
juleson mcp serve --version
```

## Client Configuration

Use an absolute path to the installed `juleson` binary:

```json
{
  "mcpServers": {
    "juleson": {
      "command": "/usr/local/bin/juleson",
      "args": ["mcp", "serve"],
      "env": {
        "JULES_API_KEY": "..."
      }
    }
  }
}
```

## Tool Scope

The MCP server exposes Jules-focused tools:

- version and config status
- source list/get
- session list/get/create/delete
- plan approval and session messaging
- activity list/get
- session plans, review, artifacts, and outputs
- developer build/test/check helpers

Mutating tools require explicit confirmation fields such as `confirm=true`.

Juleson does not expose general GitHub or Actions MCP tools. Use the official
GitHub MCP server or `gh` for that surface.
