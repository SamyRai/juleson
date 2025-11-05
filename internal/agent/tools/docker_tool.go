package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/SamyRai/juleson/internal/agent"
)

// DockerTool provides Docker container and image management capabilities to the agent
type DockerTool struct{}

// NewDockerTool creates a new Docker tool
func NewDockerTool() *DockerTool {
	return &DockerTool{}
}

// Name returns the tool name
func (d *DockerTool) Name() string {
	return "docker"
}

// Description returns what this tool does
func (d *DockerTool) Description() string {
	return "Manage Docker containers, images, and Docker operations. Build images, run containers, manage lifecycle, and execute commands."
}

// Parameters returns tool parameters
func (d *DockerTool) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "action",
			Description: "Action to perform: build, run, images, containers, stop, remove, rmi, prune, exec",
			Type:        ParameterTypeString,
			Required:    true,
		},
		{
			Name:        "path",
			Description: "Path to Dockerfile or project directory (for build)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "image",
			Description: "Docker image name/tag (for run, rmi)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "container",
			Description: "Container name/ID (for stop, remove, exec)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "command",
			Description: "Command to execute in container (for exec)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "tag",
			Description: "Image tag (for build)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "options",
			Description: "Additional Docker options (for run, build)",
			Type:        ParameterTypeString,
			Required:    false,
		},
	}
}

// Execute runs the Docker tool
func (d *DockerTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required")
	}

	switch action {
	case "build":
		return d.buildImage(ctx, params)
	case "run":
		return d.runContainer(ctx, params)
	case "images":
		return d.listImages(ctx, params)
	case "containers":
		return d.listContainers(ctx, params)
	case "stop":
		return d.stopContainer(ctx, params)
	case "remove":
		return d.removeContainer(ctx, params)
	case "rmi":
		return d.removeImage(ctx, params)
	case "prune":
		return d.pruneSystem(ctx, params)
	case "exec":
		return d.execInContainer(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// RequiresApproval returns whether this tool needs approval
func (d *DockerTool) RequiresApproval() bool {
	return true // Docker operations can be destructive, require approval
}

// CanHandle returns whether this tool can handle a task
func (d *DockerTool) CanHandle(task agent.Task) bool {
	// Can handle Docker-related tasks
	return task.Tool == "docker" ||
		containsString(task.Description, "docker") ||
		containsString(task.Description, "container") ||
		containsString(task.Description, "image") ||
		containsString(task.Prompt, "docker") ||
		containsString(task.Prompt, "container") ||
		containsString(task.Prompt, "image")
}

// buildImage builds a Docker image
func (d *DockerTool) buildImage(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required for build")
	}

	tag := "latest"
	if t, ok := params["tag"].(string); ok && t != "" {
		tag = t
	}

	options := ""
	if o, ok := params["options"].(string); ok && o != "" {
		options = o
	}

	// Construct build command
	args := []string{"build"}
	if options != "" {
		args = append(args, strings.Fields(options)...)
	}
	args = append(args, "-t", tag, path)

	output, err := d.runDockerCommand(ctx, args...)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "build",
			"path":   path,
			"tag":    tag,
			"output": output,
		},
	}, nil
}

// runContainer runs a Docker container
func (d *DockerTool) runContainer(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	image, ok := params["image"].(string)
	if !ok {
		return nil, fmt.Errorf("image parameter is required for run")
	}

	options := ""
	if o, ok := params["options"].(string); ok && o != "" {
		options = o
	}

	// Construct run command
	args := []string{"run"}
	if options != "" {
		args = append(args, strings.Fields(options)...)
	}
	args = append(args, image)

	output, err := d.runDockerCommand(ctx, args...)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "run",
			"image":  image,
			"output": output,
		},
	}, nil
}

// listImages lists Docker images
func (d *DockerTool) listImages(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	output, err := d.runDockerCommand(ctx, "images")
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "images",
			"output": output,
		},
	}, nil
}

// listContainers lists Docker containers
func (d *DockerTool) listContainers(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	output, err := d.runDockerCommand(ctx, "ps", "-a")
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "containers",
			"output": output,
		},
	}, nil
}

// stopContainer stops a Docker container
func (d *DockerTool) stopContainer(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	container, ok := params["container"].(string)
	if !ok {
		return nil, fmt.Errorf("container parameter is required for stop")
	}

	output, err := d.runDockerCommand(ctx, "stop", container)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action":    "stop",
			"container": container,
			"output":    output,
		},
	}, nil
}

// removeContainer removes a Docker container
func (d *DockerTool) removeContainer(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	container, ok := params["container"].(string)
	if !ok {
		return nil, fmt.Errorf("container parameter is required for remove")
	}

	output, err := d.runDockerCommand(ctx, "rm", container)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action":    "remove",
			"container": container,
			"output":    output,
		},
	}, nil
}

// removeImage removes a Docker image
func (d *DockerTool) removeImage(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	image, ok := params["image"].(string)
	if !ok {
		return nil, fmt.Errorf("image parameter is required for rmi")
	}

	output, err := d.runDockerCommand(ctx, "rmi", image)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "rmi",
			"image":  image,
			"output": output,
		},
	}, nil
}

// pruneSystem prunes Docker system
func (d *DockerTool) pruneSystem(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	output, err := d.runDockerCommand(ctx, "system", "prune", "-f")
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action": "prune",
			"output": output,
		},
	}, nil
}

// execInContainer executes a command in a running container
func (d *DockerTool) execInContainer(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	container, ok := params["container"].(string)
	if !ok {
		return nil, fmt.Errorf("container parameter is required for exec")
	}

	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required for exec")
	}

	args := []string{"exec", container}
	args = append(args, strings.Fields(command)...)

	output, err := d.runDockerCommand(ctx, args...)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err,
		}, err
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"action":    "exec",
			"container": container,
			"command":   command,
			"output":    output,
		},
	}, nil
}

// runDockerCommand executes a Docker command and returns the output
func (d *DockerTool) runDockerCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
