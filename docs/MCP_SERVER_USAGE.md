# Juleson MCP Server Usage

The Juleson MCP Server uses the official Model Context Protocol Go SDK and runs over stdin/stdout transport, making it compatible with AI assistants and development tools.

## 🚀 **Starting the MCP Server**

```bash
# Start the MCP server
./bin/jules-mcp
```

The server will run over stdin/stdout and wait for MCP protocol messages.

## 🔧 **Available MCP Tools**

### **1. analyze_project**

Analyze project structure and create context for automation.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to the project directory"
    }
  },
  "required": ["project_path"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "analyze_project",
    "arguments": {
      "project_path": "./my-project"
    }
  }
}
```

### **2. execute_template**

Execute a template on a project.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "template_name": {
      "type": "string",
      "description": "Name of the template to execute"
    },
    "project_path": {
      "type": "string",
      "description": "Path to the project directory"
    },
    "custom_params": {
      "type": "object",
      "description": "Custom parameters for the template"
    }
  },
  "required": ["template_name", "project_path"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "execute_template",
    "arguments": {
      "template_name": "modular-restructure",
      "project_path": "./my-project",
      "custom_params": {
        "target_layers": "presentation,business,data",
        "preserve_tests": "true"
      }
    }
  }
}
```

### **3. list_templates**

List available automation templates.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "category": {
      "type": "string",
      "description": "Filter templates by category"
    }
  }
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "list_templates",
    "arguments": {
      "category": "reorganization"
    }
  }
}
```

### **4. search_templates**

Search templates by query.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search query"
    }
  },
  "required": ["query"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "search_templates",
    "arguments": {
      "query": "modular architecture"
    }
  }
}
```

### **5. create_template**

Create a new custom template.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Template name"
    },
    "category": {
      "type": "string",
      "description": "Template category"
    },
    "description": {
      "type": "string",
      "description": "Template description"
    }
  },
  "required": ["name", "category", "description"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "create_template",
    "arguments": {
      "name": "my-custom-template",
      "category": "custom",
      "description": "My custom automation template"
    }
  }
}
```

### **6. sync_project**

Sync project with remote repository.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "project_path": {
      "type": "string",
      "description": "Path to the project directory"
    },
    "remote": {
      "type": "string",
      "description": "Remote repository name"
    }
  },
  "required": ["project_path", "remote"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "sync_project",
    "arguments": {
      "project_path": "./my-project",
      "remote": "origin/main"
    }
  }
}
```

## 🔄 **MCP Protocol Flow**

### **1. Initialize**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "clientInfo": {
      "name": "Juleson-client",
      "version": "1.0.0"
    }
  }
}
```

### **2. List Tools**

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

### **3. Call Tool**

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "analyze_project",
    "arguments": {
      "project_path": "./my-project"
    }
  }
}
```

## 🎯 **Integration Examples**

### **With Claude Desktop**

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "Juleson": {
      "command": "/path/to/Juleson/bin/jules-mcp",
      "env": {
        "JULES_API_KEY": "your-jules-api-key"
      }
    }
  }
}
```

### **With Cursor**

Configure in Cursor settings:

```json
{
  "mcp.servers": {
    "Juleson": {
      "command": "/path/to/Juleson/bin/jules-mcp",
      "env": {
        "JULES_API_KEY": "your-jules-api-key"
      }
    }
  }
}
```

### **With Custom MCP Client**

```go
package main

import (
    "context"
    "log"
    "os/exec"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    ctx := context.Background()

    // Create MCP client
    client := mcp.NewClient(&mcp.Implementation{
        Name:    "Juleson-client",
        Version: "1.0.0",
    }, nil)

    // Connect to Jules automation server
    transport := &mcp.CommandTransport{
        Command: exec.Command("/path/to/Juleson/bin/jules-mcp"),
    }

    session, err := client.Connect(ctx, transport, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // Call analyze_project tool
    params := &mcp.CallToolParams{
        Name:      "analyze_project",
        Arguments: map[string]any{"project_path": "./my-project"},
    }

    res, err := session.CallTool(ctx, params)
    if err != nil {
        log.Fatalf("CallTool failed: %v", err)
    }

    if res.IsError {
        log.Fatal("tool failed")
    }

    for _, c := range res.Content {
        log.Print(c.(*mcp.TextContent).Text)
    }
}
```

## 🔧 **Configuration**

The MCP server uses the same configuration as the CLI tool. Set environment variables:

```bash
export JULES_API_KEY="your-jules-api-key"
export JULES_BASE_URL="https://jules.googleapis.com/v1alpha"
export JULES_TIMEOUT="30s"
export JULES_RETRY_ATTEMPTS="3"
```

Or create a `configs/Juleson.yaml` file:

```yaml
jules:
  api_key: "${JULES_API_KEY}"
  base_url: "https://jules.googleapis.com/v1alpha"
  timeout: "30s"
  retry_attempts: 3

automation:
  strategies:
    - "modular"
    - "layered"
    - "microservices"
  max_concurrent_tasks: 5
  task_timeout: "300s"

projects:
  default_path: "./projects"
  backup_enabled: true
  git_integration: true
```

## 🚀 **Example Workflow**

1. **Start MCP Server**: `./bin/jules-mcp`
2. **Connect from AI Assistant**: Configure MCP server in your AI assistant
3. **Analyze Project**: Use `analyze_project` tool to understand project structure
4. **List Templates**: Use `list_templates` to see available automation options
5. **Execute Template**: Use `execute_template` to run automation on your project
6. **Sync Changes**: Use `sync_project` to commit and push changes

---

*Juleson MCP Server - Native integration with AI assistants using official MCP Go SDK*
