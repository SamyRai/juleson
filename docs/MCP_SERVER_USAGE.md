# MCP Server Usage

Juleson implements the Model Context Protocol (MCP) using the official Go MCP SDK. The server communicates exclusively over standard input/output (`stdio`).

Because it operates via `stdio`, Juleson does not bind to a network socket or listen on a TCP port. It is designed to run as a subprocess spawned by an MCP client (such as an IDE or an AI agent).

## Start The Server

Launch the server using the main binary:

```bash
juleson mcp serve
```

Or the short alias:

```bash
jsn mcp serve
```

For a fast smoke check that confirms execution without entering the `stdio` blocking loop:

```bash
juleson mcp serve --version
```

## Client Configuration

Clients must execute Juleson as a subprocess. Configure your client using the absolute path to the installed binary. Pass required credentials through the environment.

Example configuration for Claude Desktop or similar MCP clients:

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

Juleson exposes tools focused strictly on Jules operations and local automation:

- **Diagnostics**: Version and configuration status.
- **Sources**: List and retrieve configured sources.
- **Sessions**: Lifecycle management (list, get, create, delete).
- **Execution**: Plan approval and session messaging.
- **Inspection**: Activity lists, plan details, reviews, artifacts, and outputs.
- **Development**: Local build, test, and check orchestration.

Mutating tools require explicit confirmation arguments (e.g., `confirm=true`) to prevent accidental execution.

Juleson does not duplicate generic source control tooling. For general GitHub or Actions workflows, use the official GitHub MCP server.
