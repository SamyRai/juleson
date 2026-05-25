package orchestrator

import (
	"context"
	"fmt"
	"strings"
)

type DockerOperations struct {
	runner commandRunner
}

func NewDockerOperations() *DockerOperations {
	return &DockerOperations{runner: directCommandRunner{}}
}

type directCommandRunner struct{}

func (directCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	_, err := directCommandRunner{}.CombinedOutput(ctx, name, args...)
	return err
}

func (directCommandRunner) CombinedOutput(ctx context.Context, name string, args ...string) (string, error) {
	return shellCommandRunner{}.CombinedOutput(ctx, name, args...)
}

type DockerBuildOptions struct {
	Path       string
	Tag        string
	Dockerfile string
	BuildArgs  map[string]string
	NoCache    bool
}

type DockerBuildResult struct {
	Success bool
	ImageID string
	Tag     string
	Output  string
}

func (d *DockerOperations) Build(ctx context.Context, options DockerBuildOptions) (DockerBuildResult, error) {
	if options.Path == "" {
		options.Path = "."
	}
	if options.Tag == "" {
		options.Tag = "latest"
	}
	if options.Dockerfile == "" {
		options.Dockerfile = "Dockerfile"
	}

	args := []string{"build", "-t", options.Tag}
	if options.Dockerfile != "Dockerfile" {
		args = append(args, "-f", options.Dockerfile)
	}
	if options.NoCache {
		args = append(args, "--no-cache")
	}
	for key, value := range options.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}
	args = append(args, options.Path)

	output, err := d.runner.CombinedOutput(ctx, "docker", args...)
	result := DockerBuildResult{Success: err == nil, Tag: options.Tag, Output: output}
	if err != nil {
		return result, err
	}
	result.ImageID = parseDockerBuildImageID(output)
	return result, nil
}

type DockerRunOptions struct {
	Image       string
	Name        string
	Command     []string
	Environment map[string]string
	Ports       map[string]string
	Volumes     map[string]string
	Detach      bool
	Remove      bool
	Interactive bool
	TTY         bool
}

type DockerRunResult struct {
	Success       bool
	ContainerID   string
	ContainerName string
	Output        string
}

func (d *DockerOperations) RunContainer(ctx context.Context, options DockerRunOptions) (DockerRunResult, error) {
	if options.Image == "" {
		return DockerRunResult{}, fmt.Errorf("Image is required")
	}

	args := []string{"run"}
	if options.Detach {
		args = append(args, "-d")
	}
	if options.Remove {
		args = append(args, "--rm")
	}
	if options.Interactive {
		args = append(args, "-i")
	}
	if options.TTY {
		args = append(args, "-t")
	}
	for key, value := range options.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	for host, container := range options.Ports {
		args = append(args, "-p", fmt.Sprintf("%s:%s", host, container))
	}
	for host, container := range options.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, container))
	}
	if options.Name != "" {
		args = append(args, "--name", options.Name)
	}
	args = append(args, options.Image)
	args = append(args, options.Command...)

	output, err := d.runner.CombinedOutput(ctx, "docker", args...)
	result := DockerRunResult{Success: err == nil, ContainerName: options.Name, Output: output}
	if err != nil {
		return result, err
	}
	result.ContainerID = trimContainerID(output)
	return result, nil
}

type DockerListOptions struct {
	All    bool
	Filter string
	Format string
	Quiet  bool
	Latest bool
}

type DockerListResult struct {
	Success bool
	Items   []string
	Output  string
}

func (d *DockerOperations) Images(ctx context.Context, options DockerListOptions) (DockerListResult, error) {
	args := []string{"images"}
	if options.All {
		args = append(args, "-a")
	}
	if options.Filter != "" {
		args = append(args, "--filter", options.Filter)
	}
	if options.Format != "" {
		args = append(args, "--format", options.Format)
	}
	if options.Quiet {
		args = append(args, "-q")
	}

	output, err := d.runner.CombinedOutput(ctx, "docker", args...)
	return DockerListResult{Success: err == nil, Items: parseDockerListOutput(output, options.Quiet), Output: output}, err
}

func (d *DockerOperations) Containers(ctx context.Context, options DockerListOptions) (DockerListResult, error) {
	args := []string{"ps"}
	if options.All {
		args = append(args, "-a")
	}
	if options.Filter != "" {
		args = append(args, "--filter", options.Filter)
	}
	if options.Format != "" {
		args = append(args, "--format", options.Format)
	}
	if options.Quiet {
		args = append(args, "-q")
	}
	if options.Latest {
		args = append(args, "-l")
	}

	output, err := d.runner.CombinedOutput(ctx, "docker", args...)
	return DockerListResult{Success: err == nil, Items: parseDockerListOutput(output, options.Quiet), Output: output}, err
}

func (d *DockerOperations) Stop(ctx context.Context, container string, timeout int) (string, error) {
	if container == "" {
		return "", fmt.Errorf("Container ID or name is required")
	}
	args := []string{"stop"}
	if timeout > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", timeout))
	}
	args = append(args, container)
	return d.runner.CombinedOutput(ctx, "docker", args...)
}

func (d *DockerOperations) RemoveContainer(ctx context.Context, container string, force, volumes bool) (string, error) {
	if container == "" {
		return "", fmt.Errorf("Container ID or name is required")
	}
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	if volumes {
		args = append(args, "-v")
	}
	args = append(args, container)
	return d.runner.CombinedOutput(ctx, "docker", args...)
}

func (d *DockerOperations) RemoveImage(ctx context.Context, image string, force bool) (string, error) {
	if image == "" {
		return "", fmt.Errorf("Image ID or name is required")
	}
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, image)
	return d.runner.CombinedOutput(ctx, "docker", args...)
}

func (d *DockerOperations) Prune(ctx context.Context, all, volumes bool) (string, error) {
	args := []string{"system", "prune", "-f"}
	if all {
		args = append(args, "-a")
	}
	if volumes {
		args = append(args, "--volumes")
	}
	return d.runner.CombinedOutput(ctx, "docker", args...)
}

type DockerExecOptions struct {
	Container   string
	Command     []string
	User        string
	WorkDir     string
	Detach      bool
	TTY         bool
	Interactive bool
}

func (d *DockerOperations) Exec(ctx context.Context, options DockerExecOptions) (string, error) {
	if options.Container == "" {
		return "", fmt.Errorf("Container ID or name is required")
	}
	if len(options.Command) == 0 {
		return "", fmt.Errorf("Command is required")
	}

	args := []string{"exec"}
	if options.User != "" {
		args = append(args, "-u", options.User)
	}
	if options.WorkDir != "" {
		args = append(args, "-w", options.WorkDir)
	}
	if options.Detach {
		args = append(args, "-d")
	}
	if options.TTY {
		args = append(args, "-t")
	}
	if options.Interactive {
		args = append(args, "-i")
	}
	args = append(args, options.Container)
	args = append(args, options.Command...)

	return d.runner.CombinedOutput(ctx, "docker", args...)
}

func parseDockerBuildImageID(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return ""
	}
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "Successfully built ") {
		return ""
	}
	parts := strings.Fields(lastLine)
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

func parseDockerListOutput(output string, quiet bool) []string {
	if output == "" {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if quiet {
		return lines
	}
	if len(lines) > 1 {
		return lines[1:]
	}
	return nil
}

func trimContainerID(output string) string {
	containerID := strings.TrimSpace(output)
	if len(containerID) > 12 {
		return containerID[:12]
	}
	return containerID
}
