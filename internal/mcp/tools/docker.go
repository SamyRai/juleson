package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterDockerTools registers Docker-related MCP tools
func RegisterDockerTools(server *mcp.Server) {
	// Build Docker image
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_build",
		Description: "Build a Docker image from a Dockerfile",
	}, dockerBuildHandler)

	// Run Docker container
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_run",
		Description: "Run a Docker container",
	}, dockerRunHandler)

	// List Docker images
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_images",
		Description: "List Docker images",
	}, dockerImagesHandler)

	// List Docker containers
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_containers",
		Description: "List Docker containers",
	}, dockerContainersHandler)

	// Stop Docker container
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_stop",
		Description: "Stop a running Docker container",
	}, dockerStopHandler)

	// Remove Docker container
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_remove",
		Description: "Remove a Docker container",
	}, dockerRemoveHandler)

	// Remove Docker image
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_rmi",
		Description: "Remove a Docker image",
	}, dockerRmiHandler)

	// Docker system prune
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_prune",
		Description: "Clean up Docker system (remove unused containers, networks, images)",
	}, dockerPruneHandler)

	// Execute command in Docker container
	mcp.AddTool(server, &mcp.Tool{
		Name:        "docker_exec",
		Description: "Execute a command in a running Docker container",
	}, dockerExecHandler)
}

// Input/Output types

type DockerBuildInput struct {
	Path       string            `json:"path" jsonschema:"Path to the directory containing the Dockerfile (default: .)"`
	Tag        string            `json:"tag" jsonschema:"Image tag (default: latest)"`
	Dockerfile string            `json:"dockerfile" jsonschema:"Path to Dockerfile (default: Dockerfile)"`
	BuildArgs  map[string]string `json:"build_args" jsonschema:"Build arguments as key-value pairs"`
	NoCache    bool              `json:"no_cache" jsonschema:"Do not use cache when building the image (default: false)"`
}

type DockerBuildOutput struct {
	Success bool   `json:"success"`
	ImageID string `json:"image_id,omitempty"`
	Tag     string `json:"tag"`
	Output  string `json:"output"`
}

type DockerRunInput struct {
	Image       string            `json:"image" jsonschema:"Docker image to run"`
	Name        string            `json:"name" jsonschema:"Container name (optional)"`
	Command     []string          `json:"command" jsonschema:"Command to run in the container (optional)"`
	Environment map[string]string `json:"environment" jsonschema:"Environment variables as key-value pairs"`
	Ports       map[string]string `json:"ports" jsonschema:"Port mappings as host:container"`
	Volumes     map[string]string `json:"volumes" jsonschema:"Volume mappings as host:container"`
	Detach      bool              `json:"detach" jsonschema:"Run container in background (default: false)"`
	Remove      bool              `json:"remove" jsonschema:"Automatically remove the container when it exits (default: false)"`
	Interactive bool              `json:"interactive" jsonschema:"Keep STDIN open even if not attached (default: false)"`
	TTY         bool              `json:"tty" jsonschema:"Allocate a pseudo-TTY (default: false)"`
}

type DockerRunOutput struct {
	Success       bool   `json:"success"`
	ContainerID   string `json:"container_id,omitempty"`
	ContainerName string `json:"container_name,omitempty"`
	Output        string `json:"output"`
}

type DockerImagesInput struct {
	All    bool   `json:"all" jsonschema:"Show all images (default: false)"`
	Filter string `json:"filter" jsonschema:"Filter output based on conditions provided"`
	Format string `json:"format" jsonschema:"Pretty-print images using a Go template"`
	Quiet  bool   `json:"quiet" jsonschema:"Only show image IDs (default: false)"`
}

type DockerImagesOutput struct {
	Success bool     `json:"success"`
	Images  []string `json:"images"`
	Count   int      `json:"count"`
	Output  string   `json:"output"`
}

type DockerContainersInput struct {
	All    bool   `json:"all" jsonschema:"Show all containers (default: false)"`
	Filter string `json:"filter" jsonschema:"Filter output based on conditions provided"`
	Format string `json:"format" jsonschema:"Pretty-print containers using a Go template"`
	Quiet  bool   `json:"quiet" jsonschema:"Only show container IDs (default: false)"`
	Latest bool   `json:"latest" jsonschema:"Show the latest created container (includes all states) (default: false)"`
}

type DockerContainersOutput struct {
	Success    bool     `json:"success"`
	Containers []string `json:"containers"`
	Count      int      `json:"count"`
	Output     string   `json:"output"`
}

type DockerStopInput struct {
	Container string `json:"container" jsonschema:"Container ID or name to stop"`
	Time      int    `json:"time" jsonschema:"Seconds to wait before killing the container (default: 10)"`
}

type DockerStopOutput struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type DockerRemoveInput struct {
	Container string `json:"container" jsonschema:"Container ID or name to remove"`
	Force     bool   `json:"force" jsonschema:"Force the removal of a running container (uses SIGKILL) (default: false)"`
	Volumes   bool   `json:"volumes" jsonschema:"Remove anonymous volumes associated with the container (default: false)"`
}

type DockerRemoveOutput struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type DockerRmiInput struct {
	Image string `json:"image" jsonschema:"Image ID or name to remove"`
	Force bool   `json:"force" jsonschema:"Force removal of the image (default: false)"`
}

type DockerRmiOutput struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type DockerPruneInput struct {
	All     bool `json:"all" jsonschema:"Remove all unused images not just dangling ones (default: false)"`
	Volumes bool `json:"volumes" jsonschema:"Prune volumes (default: false)"`
}

type DockerPruneOutput struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type DockerExecInput struct {
	Container   string   `json:"container" jsonschema:"Container ID or name"`
	Command     []string `json:"command" jsonschema:"Command to execute"`
	User        string   `json:"user" jsonschema:"Username or UID (format: <name|uid>[:<group|gid>])"`
	WorkDir     string   `json:"workdir" jsonschema:"Working directory inside the container"`
	Detach      bool     `json:"detach" jsonschema:"Detached mode: run command in the background (default: false)"`
	TTY         bool     `json:"tty" jsonschema:"Allocate a pseudo-TTY (default: false)"`
	Interactive bool     `json:"interactive" jsonschema:"Pass stdin to the container (default: false)"`
}

type DockerExecOutput struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

// Handler functions

func dockerBuildHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerBuildInput) (
	*mcp.CallToolResult,
	DockerBuildOutput,
	error,
) {
	// Set defaults
	if input.Path == "" {
		input.Path = "."
	}
	if input.Tag == "" {
		input.Tag = "latest"
	}
	if input.Dockerfile == "" {
		input.Dockerfile = "Dockerfile"
	}

	// Build docker build command
	args := []string{"build", "-t", input.Tag}

	if input.Dockerfile != "Dockerfile" {
		args = append(args, "-f", input.Dockerfile)
	}

	if input.NoCache {
		args = append(args, "--no-cache")
	}

	// Add build args
	for key, value := range input.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, input.Path)

	// Execute command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker build failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerBuildOutput{Success: false, Tag: input.Tag, Output: outputStr}, nil
	}

	// Extract image ID from output (usually the last line)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	imageID := ""
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if strings.HasPrefix(lastLine, "Successfully built ") {
			parts := strings.Fields(lastLine)
			if len(parts) >= 3 {
				imageID = parts[2]
			}
		}
	}

	return nil, DockerBuildOutput{
		Success: true,
		ImageID: imageID,
		Tag:     input.Tag,
		Output:  outputStr,
	}, nil
}

func dockerRunHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRunInput) (
	*mcp.CallToolResult,
	DockerRunOutput,
	error,
) {
	if input.Image == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Image is required"},
			},
		}, DockerRunOutput{}, nil
	}

	args := []string{"run"}

	if input.Detach {
		args = append(args, "-d")
	}
	if input.Remove {
		args = append(args, "--rm")
	}
	if input.Interactive {
		args = append(args, "-i")
	}
	if input.TTY {
		args = append(args, "-t")
	}

	// Add environment variables
	for key, value := range input.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add port mappings
	for host, container := range input.Ports {
		args = append(args, "-p", fmt.Sprintf("%s:%s", host, container))
	}

	// Add volume mappings
	for host, container := range input.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, container))
	}

	// Add container name
	if input.Name != "" {
		args = append(args, "--name", input.Name)
	}

	// Add image
	args = append(args, input.Image)

	// Add command
	if len(input.Command) > 0 {
		args = append(args, input.Command...)
	}

	// Execute command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker run failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerRunOutput{Success: false, Output: outputStr}, nil
	}

	// Extract container ID (first 12 characters of the output)
	containerID := strings.TrimSpace(outputStr)
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	return nil, DockerRunOutput{
		Success:       true,
		ContainerID:   containerID,
		ContainerName: input.Name,
		Output:        outputStr,
	}, nil
}

func dockerImagesHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerImagesInput) (
	*mcp.CallToolResult,
	DockerImagesOutput,
	error,
) {
	args := []string{"images"}

	if input.All {
		args = append(args, "-a")
	}
	if input.Filter != "" {
		args = append(args, "--filter", input.Filter)
	}
	if input.Format != "" {
		args = append(args, "--format", input.Format)
	}
	if input.Quiet {
		args = append(args, "-q")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker images failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerImagesOutput{Success: false, Output: outputStr}, nil
	}

	// Parse images from output
	var images []string
	if !input.Quiet && outputStr != "" {
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		// Skip header line
		if len(lines) > 1 {
			images = lines[1:]
		}
	} else if input.Quiet && outputStr != "" {
		images = strings.Split(strings.TrimSpace(outputStr), "\n")
	}

	return nil, DockerImagesOutput{
		Success: true,
		Images:  images,
		Count:   len(images),
		Output:  outputStr,
	}, nil
}

func dockerContainersHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerContainersInput) (
	*mcp.CallToolResult,
	DockerContainersOutput,
	error,
) {
	args := []string{"ps"}

	if input.All {
		args = append(args, "-a")
	}
	if input.Filter != "" {
		args = append(args, "--filter", input.Filter)
	}
	if input.Format != "" {
		args = append(args, "--format", input.Format)
	}
	if input.Quiet {
		args = append(args, "-q")
	}
	if input.Latest {
		args = append(args, "-l")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker ps failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerContainersOutput{Success: false, Output: outputStr}, nil
	}

	// Parse containers from output
	var containers []string
	if !input.Quiet && outputStr != "" {
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		// Skip header line
		if len(lines) > 1 {
			containers = lines[1:]
		}
	} else if input.Quiet && outputStr != "" {
		containers = strings.Split(strings.TrimSpace(outputStr), "\n")
	}

	return nil, DockerContainersOutput{
		Success:    true,
		Containers: containers,
		Count:      len(containers),
		Output:     outputStr,
	}, nil
}

func dockerStopHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerStopInput) (
	*mcp.CallToolResult,
	DockerStopOutput,
	error,
) {
	if input.Container == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Container ID or name is required"},
			},
		}, DockerStopOutput{}, nil
	}

	args := []string{"stop"}

	if input.Time > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", input.Time))
	}

	args = append(args, input.Container)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker stop failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerStopOutput{Success: false, Output: outputStr}, nil
	}

	return nil, DockerStopOutput{
		Success: true,
		Output:  outputStr,
	}, nil
}

func dockerRemoveHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRemoveInput) (
	*mcp.CallToolResult,
	DockerRemoveOutput,
	error,
) {
	if input.Container == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Container ID or name is required"},
			},
		}, DockerRemoveOutput{}, nil
	}

	args := []string{"rm"}

	if input.Force {
		args = append(args, "-f")
	}
	if input.Volumes {
		args = append(args, "-v")
	}

	args = append(args, input.Container)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker rm failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerRemoveOutput{Success: false, Output: outputStr}, nil
	}

	return nil, DockerRemoveOutput{
		Success: true,
		Output:  outputStr,
	}, nil
}

func dockerRmiHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRmiInput) (
	*mcp.CallToolResult,
	DockerRmiOutput,
	error,
) {
	if input.Image == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Image ID or name is required"},
			},
		}, DockerRmiOutput{}, nil
	}

	args := []string{"rmi"}

	if input.Force {
		args = append(args, "-f")
	}

	args = append(args, input.Image)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker rmi failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerRmiOutput{Success: false, Output: outputStr}, nil
	}

	return nil, DockerRmiOutput{
		Success: true,
		Output:  outputStr,
	}, nil
}

func dockerPruneHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerPruneInput) (
	*mcp.CallToolResult,
	DockerPruneOutput,
	error,
) {
	args := []string{"system", "prune", "-f"}

	if input.All {
		args = append(args, "-a")
	}
	if input.Volumes {
		args = append(args, "--volumes")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker prune failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerPruneOutput{Success: false, Output: outputStr}, nil
	}

	return nil, DockerPruneOutput{
		Success: true,
		Output:  outputStr,
	}, nil
}

func dockerExecHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerExecInput) (
	*mcp.CallToolResult,
	DockerExecOutput,
	error,
) {
	if input.Container == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Container ID or name is required"},
			},
		}, DockerExecOutput{}, nil
	}

	if len(input.Command) == 0 {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Command is required"},
			},
		}, DockerExecOutput{}, nil
	}

	args := []string{"exec"}

	if input.User != "" {
		args = append(args, "-u", input.User)
	}
	if input.WorkDir != "" {
		args = append(args, "-w", input.WorkDir)
	}
	if input.Detach {
		args = append(args, "-d")
	}
	if input.TTY {
		args = append(args, "-t")
	}
	if input.Interactive {
		args = append(args, "-i")
	}

	args = append(args, input.Container)
	args = append(args, input.Command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker exec failed: %v\nOutput: %s", err, outputStr)},
			},
		}, DockerExecOutput{Success: false, Output: outputStr}, nil
	}

	return nil, DockerExecOutput{
		Success: true,
		Output:  outputStr,
	}, nil
}
