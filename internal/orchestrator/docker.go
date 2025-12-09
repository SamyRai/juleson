package orchestrator

import (
	"context"
	"fmt"
	"os"
)

// DockerBuild builds the Docker image
func (s *Service) DockerBuild(ctx context.Context) error {
	if err := s.runCommand(ctx, "docker", "build", "-t", s.config.DockerImage, "."); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	return nil
}

// DockerRun runs a Docker container with the given arguments
func (s *Service) DockerRun(ctx context.Context, args []string) error {
	// Build first
	if err := s.DockerBuild(ctx); err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	dockerArgs := []string{
		"run", "--rm", "-it",
		"-e", "JULES_API_KEY=" + os.Getenv("JULES_API_KEY"),
		"-v", fmt.Sprintf("%s:/workspace", workDir),
		"-w", "/workspace",
		s.config.DockerImage,
	}
	dockerArgs = append(dockerArgs, args...)

	if err := s.runCommand(ctx, "docker", dockerArgs...); err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	return nil
}

// DockerRunCLI runs the CLI in a Docker container
func (s *Service) DockerRunCLI(ctx context.Context, args []string) error {
	// Build first
	if err := s.DockerBuild(ctx); err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	dockerArgs := []string{
		"run", "--rm", "-it",
		"-e", "JULES_API_KEY=" + os.Getenv("JULES_API_KEY"),
		"-v", fmt.Sprintf("%s:/workspace", workDir),
		"-w", "/workspace",
		s.config.DockerImage,
		"./" + s.config.BinaryCLI,
	}
	dockerArgs = append(dockerArgs, args...)

	if err := s.runCommand(ctx, "docker", dockerArgs...); err != nil {
		return fmt.Errorf("docker run CLI failed: %w", err)
	}

	return nil
}

// DockerRunMCP runs the MCP server in a Docker container
func (s *Service) DockerRunMCP(ctx context.Context) error {
	// Build first
	if err := s.DockerBuild(ctx); err != nil {
		return err
	}

	dockerArgs := []string{
		"run", "--rm", "-it",
		"-e", "JULES_API_KEY=" + os.Getenv("JULES_API_KEY"),
		"-p", "8080:8080",
		s.config.DockerImage,
		"./" + s.config.BinaryMCP,
	}

	if err := s.runCommand(ctx, "docker", dockerArgs...); err != nil {
		return fmt.Errorf("docker run MCP failed: %w", err)
	}

	return nil
}

// DockerPush pushes the Docker image to the registry
func (s *Service) DockerPush(ctx context.Context) error {
	// Build first
	if err := s.DockerBuild(ctx); err != nil {
		return err
	}

	if err := s.runCommand(ctx, "docker", "push", s.config.DockerImage); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	return nil
}

// DockerComposeUp starts services with docker-compose
func (s *Service) DockerComposeUp(ctx context.Context) error {
	if err := s.runCommand(ctx, "docker-compose", "up", "--build"); err != nil {
		return fmt.Errorf("docker-compose up failed: %w", err)
	}

	return nil
}

// DockerComposeDown stops services with docker-compose
func (s *Service) DockerComposeDown(ctx context.Context) error {
	if err := s.runCommand(ctx, "docker-compose", "down"); err != nil {
		return fmt.Errorf("docker-compose down failed: %w", err)
	}

	return nil
}

// DockerClean cleans Docker artifacts
func (s *Service) DockerClean(ctx context.Context) error {
	if err := s.runCommand(ctx, "docker", "system", "prune", "-f"); err != nil {
		return fmt.Errorf("docker system prune failed: %w", err)
	}

	// Try to remove the image (ignore errors if it doesn't exist)
	_ = s.runCommand(ctx, "docker", "rmi", s.config.DockerImage)

	return nil
}
