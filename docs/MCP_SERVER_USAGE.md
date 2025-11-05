# Juleson MCP Server Usage

The Juleson MCP Server uses the official Model Context Protocol Go SDK and runs over
stdin/stdout transport, making it compatible with AI assistants and development tools.

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

### **7. docker_build**

Build a Docker image from a Dockerfile.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to the directory containing the Dockerfile (default: .)"
    },
    "tag": {
      "type": "string",
      "description": "Image tag (default: latest)"
    },
    "dockerfile": {
      "type": "string",
      "description": "Path to Dockerfile (default: Dockerfile)"
    },
    "build_args": {
      "type": "object",
      "description": "Build arguments as key-value pairs"
    },
    "no_cache": {
      "type": "boolean",
      "description": "Do not use cache when building the image (default: false)"
    }
  }
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_build",
    "arguments": {
      "path": "./my-app",
      "tag": "my-app:v1.0",
      "build_args": {
        "VERSION": "1.0.0"
      }
    }
  }
}
```

### **8. docker_run**

Run a Docker container.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "image": {
      "type": "string",
      "description": "Docker image to run"
    },
    "name": {
      "type": "string",
      "description": "Container name (optional)"
    },
    "command": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Command to run in the container (optional)"
    },
    "environment": {
      "type": "object",
      "description": "Environment variables as key-value pairs"
    },
    "ports": {
      "type": "object",
      "description": "Port mappings as host:container"
    },
    "volumes": {
      "type": "object",
      "description": "Volume mappings as host:container"
    },
    "detach": {
      "type": "boolean",
      "description": "Run container in background (default: false)"
    },
    "remove": {
      "type": "boolean",
      "description": "Automatically remove the container when it exits (default: false)"
    },
    "interactive": {
      "type": "boolean",
      "description": "Keep STDIN open even if not attached (default: false)"
    },
    "tty": {
      "type": "boolean",
      "description": "Allocate a pseudo-TTY (default: false)"
    }
  },
  "required": ["image"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_run",
    "arguments": {
      "image": "nginx:latest",
      "name": "my-nginx",
      "ports": {
        "8080": "80"
      },
      "detach": true
    }
  }
}
```

### **9. docker_images**

List Docker images.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "all": {
      "type": "boolean",
      "description": "Show all images (default: false)"
    },
    "filter": {
      "type": "string",
      "description": "Filter output based on conditions provided"
    },
    "format": {
      "type": "string",
      "description": "Pretty-print images using a Go template"
    },
    "quiet": {
      "type": "boolean",
      "description": "Only show image IDs (default: false)"
    }
  }
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_images",
    "arguments": {
      "all": true
    }
  }
}
```

### **10. docker_containers**

List Docker containers.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "all": {
      "type": "boolean",
      "description": "Show all containers (default: false)"
    },
    "filter": {
      "type": "string",
      "description": "Filter output based on conditions provided"
    },
    "format": {
      "type": "string",
      "description": "Pretty-print containers using a Go template"
    },
    "quiet": {
      "type": "boolean",
      "description": "Only show container IDs (default: false)"
    },
    "latest": {
      "type": "boolean",
      "description": "Show the latest created container (includes all states) (default: false)"
    }
  }
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_containers",
    "arguments": {
      "all": true
    }
  }
}
```

### **11. docker_stop**

Stop a running Docker container.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "container": {
      "type": "string",
      "description": "Container ID or name to stop"
    },
    "time": {
      "type": "integer",
      "description": "Seconds to wait before killing the container (default: 10)"
    }
  },
  "required": ["container"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_stop",
    "arguments": {
      "container": "my-container"
    }
  }
}
```

### **12. docker_remove**

Remove a Docker container.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "container": {
      "type": "string",
      "description": "Container ID or name to remove"
    },
    "force": {
      "type": "boolean",
      "description": "Force the removal of a running container (uses SIGKILL) (default: false)"
    },
    "volumes": {
      "type": "boolean",
      "description": "Remove anonymous volumes associated with the container (default: false)"
    }
  },
  "required": ["container"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_remove",
    "arguments": {
      "container": "my-container",
      "force": true
    }
  }
}
```

### **13. docker_rmi**

Remove a Docker image.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "image": {
      "type": "string",
      "description": "Image ID or name to remove"
    },
    "force": {
      "type": "boolean",
      "description": "Force removal of the image (default: false)"
    }
  },
  "required": ["image"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_rmi",
    "arguments": {
      "image": "my-image:latest"
    }
  }
}
```

### **14. docker_prune**

Clean up Docker system (remove unused containers, networks, images).

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "all": {
      "type": "boolean",
      "description": "Remove all unused images not just dangling ones (default: false)"
    },
    "volumes": {
      "type": "boolean",
      "description": "Prune volumes (default: false)"
    }
  }
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_prune",
    "arguments": {
      "all": true,
      "volumes": true
    }
  }
}
```

### **15. docker_exec**

Execute a command in a running Docker container.

**Input Schema:**

```json
{
  "type": "object",
  "properties": {
    "container": {
      "type": "string",
      "description": "Container ID or name"
    },
    "command": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Command to execute"
    },
    "user": {
      "type": "string",
      "description": "Username or UID (format: <name|uid>[:<group|gid>])"
    },
    "workdir": {
      "type": "string",
      "description": "Working directory inside the container"
    },
    "detach": {
      "type": "boolean",
      "description": "Detached mode: run command in the background (default: false)"
    },
    "tty": {
      "type": "boolean",
      "description": "Allocate a pseudo-TTY (default: false)"
    },
    "interactive": {
      "type": "boolean",
      "description": "Pass stdin to the container (default: false)"
    }
  },
  "required": ["container", "command"]
}
```

**Example Usage:**

```json
{
  "method": "tools/call",
  "params": {
    "name": "docker_exec",
    "arguments": {
      "container": "my-container",
      "command": ["ls", "-la"],
      "workdir": "/app"
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

Juleson MCP Server - Native integration with AI assistants using official MCP Go SDK
