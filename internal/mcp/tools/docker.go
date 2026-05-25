package tools

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/orchestrator"
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

func dockerOps() *orchestrator.DockerOperations {
	return orchestrator.NewDockerOperations()
}

func dockerBuildHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerBuildInput) (
	*mcp.CallToolResult,
	DockerBuildOutput,
	error,
) {
	result, err := dockerOps().Build(ctx, orchestrator.DockerBuildOptions{
		Path:       input.Path,
		Tag:        input.Tag,
		Dockerfile: input.Dockerfile,
		BuildArgs:  input.BuildArgs,
		NoCache:    input.NoCache,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker build failed: %v\nOutput: %s", err, result.Output)},
			},
		}, DockerBuildOutput{Success: false, Tag: result.Tag, Output: result.Output}, nil
	}

	return nil, DockerBuildOutput{
		Success: true,
		ImageID: result.ImageID,
		Tag:     result.Tag,
		Output:  result.Output,
	}, nil
}

func dockerRunHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRunInput) (
	*mcp.CallToolResult,
	DockerRunOutput,
	error,
) {
	result, err := dockerOps().RunContainer(ctx, orchestrator.DockerRunOptions{
		Image:       input.Image,
		Name:        input.Name,
		Command:     input.Command,
		Environment: input.Environment,
		Ports:       input.Ports,
		Volumes:     input.Volumes,
		Detach:      input.Detach,
		Remove:      input.Remove,
		Interactive: input.Interactive,
		TTY:         input.TTY,
	})
	if err != nil {
		message := fmt.Sprintf("Docker run failed: %v\nOutput: %s", err, result.Output)
		if input.Image == "" {
			message = "Image is required"
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, DockerRunOutput{Success: false, Output: result.Output}, nil
	}

	return nil, DockerRunOutput{
		Success:       true,
		ContainerID:   result.ContainerID,
		ContainerName: input.Name,
		Output:        result.Output,
	}, nil
}

func dockerImagesHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerImagesInput) (
	*mcp.CallToolResult,
	DockerImagesOutput,
	error,
) {
	result, err := dockerOps().Images(ctx, orchestrator.DockerListOptions{
		All:    input.All,
		Filter: input.Filter,
		Format: input.Format,
		Quiet:  input.Quiet,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker images failed: %v\nOutput: %s", err, result.Output)},
			},
		}, DockerImagesOutput{Success: false, Output: result.Output}, nil
	}

	return nil, DockerImagesOutput{
		Success: true,
		Images:  result.Items,
		Count:   len(result.Items),
		Output:  result.Output,
	}, nil
}

func dockerContainersHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerContainersInput) (
	*mcp.CallToolResult,
	DockerContainersOutput,
	error,
) {
	result, err := dockerOps().Containers(ctx, orchestrator.DockerListOptions{
		All:    input.All,
		Filter: input.Filter,
		Format: input.Format,
		Quiet:  input.Quiet,
		Latest: input.Latest,
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker ps failed: %v\nOutput: %s", err, result.Output)},
			},
		}, DockerContainersOutput{Success: false, Output: result.Output}, nil
	}

	return nil, DockerContainersOutput{
		Success:    true,
		Containers: result.Items,
		Count:      len(result.Items),
		Output:     result.Output,
	}, nil
}

func dockerStopHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerStopInput) (
	*mcp.CallToolResult,
	DockerStopOutput,
	error,
) {
	output, err := dockerOps().Stop(ctx, input.Container, input.Time)
	if err != nil {
		message := fmt.Sprintf("Docker stop failed: %v\nOutput: %s", err, output)
		if input.Container == "" {
			message = "Container ID or name is required"
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, DockerStopOutput{Success: false, Output: output}, nil
	}

	return nil, DockerStopOutput{
		Success: true,
		Output:  output,
	}, nil
}

func dockerRemoveHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRemoveInput) (
	*mcp.CallToolResult,
	DockerRemoveOutput,
	error,
) {
	output, err := dockerOps().RemoveContainer(ctx, input.Container, input.Force, input.Volumes)
	if err != nil {
		message := fmt.Sprintf("Docker rm failed: %v\nOutput: %s", err, output)
		if input.Container == "" {
			message = "Container ID or name is required"
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, DockerRemoveOutput{Success: false, Output: output}, nil
	}

	return nil, DockerRemoveOutput{
		Success: true,
		Output:  output,
	}, nil
}

func dockerRmiHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerRmiInput) (
	*mcp.CallToolResult,
	DockerRmiOutput,
	error,
) {
	output, err := dockerOps().RemoveImage(ctx, input.Image, input.Force)
	if err != nil {
		message := fmt.Sprintf("Docker rmi failed: %v\nOutput: %s", err, output)
		if input.Image == "" {
			message = "Image ID or name is required"
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, DockerRmiOutput{Success: false, Output: output}, nil
	}

	return nil, DockerRmiOutput{
		Success: true,
		Output:  output,
	}, nil
}

func dockerPruneHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerPruneInput) (
	*mcp.CallToolResult,
	DockerPruneOutput,
	error,
) {
	output, err := dockerOps().Prune(ctx, input.All, input.Volumes)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Docker prune failed: %v\nOutput: %s", err, output)},
			},
		}, DockerPruneOutput{Success: false, Output: output}, nil
	}

	return nil, DockerPruneOutput{
		Success: true,
		Output:  output,
	}, nil
}

func dockerExecHandler(ctx context.Context, req *mcp.CallToolRequest, input DockerExecInput) (
	*mcp.CallToolResult,
	DockerExecOutput,
	error,
) {
	output, err := dockerOps().Exec(ctx, orchestrator.DockerExecOptions{
		Container:   input.Container,
		Command:     input.Command,
		User:        input.User,
		WorkDir:     input.WorkDir,
		Detach:      input.Detach,
		TTY:         input.TTY,
		Interactive: input.Interactive,
	})
	if err != nil {
		message := fmt.Sprintf("Docker exec failed: %v\nOutput: %s", err, output)
		if input.Container == "" {
			message = "Container ID or name is required"
		} else if len(input.Command) == 0 {
			message = "Command is required"
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: message},
			},
		}, DockerExecOutput{Success: false, Output: output}, nil
	}

	return nil, DockerExecOutput{
		Success: true,
		Output:  output,
	}, nil
}
