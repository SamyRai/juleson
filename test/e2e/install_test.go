package e2e

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestInstallShellOnUbuntu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E tests in short mode")
	}

	ctx := context.Background()

	// 1. Build local binary and package it
	releaseDir := t.TempDir()
	buildAndPackage(t, releaseDir)

	installScriptPath, err := filepath.Abs("../../scripts/install.sh")
	require.NoError(t, err)

	// 2. Start Ubuntu Container
	req := testcontainers.ContainerRequest{
		Image: "ubuntu:22.04",
		Cmd:   []string{"tail", "-f", "/dev/null"}, // keep alive
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      installScriptPath,
				ContainerFilePath: "/mnt/install.sh",
				FileMode:          0o755,
			},
			{
				HostFilePath:      filepath.Join(releaseDir, "juleson-linux-amd64.tar.gz"),
				ContainerFilePath: "/mnt/juleson-linux-amd64.tar.gz",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(releaseDir, "jsn-linux-amd64.tar.gz"),
				ContainerFilePath: "/mnt/jsn-linux-amd64.tar.gz",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForExec([]string{"ls", "/mnt/install.sh"}).WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Ensure cleanup
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// 3. Install necessary prerequisites (curl, tar, gzip) since Ubuntu base might lack them
	exitCode, _, err := container.Exec(ctx, []string{"apt-get", "update"})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	exitCode, _, err = container.Exec(ctx, []string{"apt-get", "install", "-y", "curl", "tar", "gzip", "ca-certificates"})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	// 4. Run the installer script
	// Simulate the web server providing the release asset by copying the mounted files
	// curl supports file:// URLs, so we pass JULESON_INSTALL_BASE_URL
	exitCode, reader, err := container.Exec(ctx, []string{
		"bash", "-c", "JULESON_INSTALL_BASE_URL=file:///mnt bash /mnt/install.sh --version v0.0.0-test --install-dir /usr/local/bin",
	})
	require.NoError(t, err)

	outBytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("Installer output: %s", string(outBytes))
	require.Equal(t, 0, exitCode, "install script failed")

	// 5. Verify juleson exists and is executable
	exitCode, reader, err = container.Exec(ctx, []string{"/usr/local/bin/juleson", "version"})
	require.NoError(t, err)
	outBytes, err = io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, 0, exitCode, "juleson version failed")
	require.Contains(t, string(outBytes), "Juleson CLI")
}

func TestInstallShellOnAlpine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E tests in short mode")
	}

	ctx := context.Background()

	releaseDir := t.TempDir()
	buildAndPackage(t, releaseDir)

	installScriptPath, err := filepath.Abs("../../scripts/install.sh")
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		Image: "alpine:3.19",
		Cmd:   []string{"tail", "-f", "/dev/null"}, // keep alive
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      installScriptPath,
				ContainerFilePath: "/mnt/install.sh",
				FileMode:          0o755,
			},
			{
				HostFilePath:      filepath.Join(releaseDir, "juleson-linux-amd64.tar.gz"),
				ContainerFilePath: "/mnt/juleson-linux-amd64.tar.gz",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(releaseDir, "jsn-linux-amd64.tar.gz"),
				ContainerFilePath: "/mnt/jsn-linux-amd64.tar.gz",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForExec([]string{"ls", "/mnt/install.sh"}).WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// Alpine base needs bash, curl, tar
	exitCode, _, err := container.Exec(ctx, []string{"apk", "add", "--no-cache", "bash", "curl", "tar"})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	exitCode, reader, err := container.Exec(ctx, []string{
		"bash", "-c", "JULESON_INSTALL_BASE_URL=file:///mnt bash /mnt/install.sh --version v0.0.0-test --install-dir /usr/local/bin",
	})
	require.NoError(t, err)

	outBytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	t.Logf("Installer output: %s", string(outBytes))
	require.Equal(t, 0, exitCode, "install script failed")

	exitCode, reader, err = container.Exec(ctx, []string{"/usr/local/bin/juleson", "version"})
	require.NoError(t, err)
	outBytes, err = io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, 0, exitCode, "juleson version failed")
	require.Contains(t, string(outBytes), "Juleson CLI")
}
