package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallShellHelpAndSyntax(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell installer is for Linux/macOS")
	}

	syntax := exec.Command("bash", "-n", "install.sh")
	if output, err := syntax.CombinedOutput(); err != nil {
		t.Fatalf("install.sh syntax check failed: %v\n%s", err, output)
	}

	help := exec.Command("bash", "install.sh", "--help")
	output, err := help.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh --help failed: %v\n%s", err, output)
	}
	for _, want := range []string{"--version", "--install-dir", "INSTALL_DIR"} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("install.sh help missing %q:\n%s", want, output)
		}
	}
}

func TestInstallShellInstallsBothBinariesFromLocalRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping installer E2E test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("shell installer is for Linux/macOS")
	}

	goos, goarch, ok := installerPlatform()
	if !ok {
		t.Skipf("installer does not support %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	releaseDir := t.TempDir()
	for _, binary := range []string{"juleson", "jules-mcp"} {
		asset := fmt.Sprintf("%s-%s-%s.tar.gz", binary, goos, goarch)
		if err := writeTarGz(filepath.Join(releaseDir, asset), binary, []byte("#!/bin/sh\necho "+binary+"\n")); err != nil {
			t.Fatal(err)
		}
	}

	server := httptest.NewServer(http.FileServer(http.Dir(releaseDir)))
	t.Cleanup(server.Close)

	installDir := filepath.Join(t.TempDir(), "bin")
	cmd := exec.Command("bash", "install.sh", "--version", "v0.0.0-test", "--install-dir", installDir)
	cmd.Env = append(os.Environ(), "JULESON_INSTALL_BASE_URL="+server.URL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, output)
	}

	for _, binary := range []string{"juleson", "jules-mcp"} {
		path := filepath.Join(installDir, binary)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected installed binary %s: %v\ninstaller output:\n%s", path, err, output)
		}
		if info.Mode()&0o111 == 0 {
			t.Fatalf("expected %s to be executable, mode %s", path, info.Mode())
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(contents, []byte("echo "+binary)) {
			t.Fatalf("installed %s has unexpected contents: %q", binary, contents)
		}
	}
}

func TestPowerShellInstallerHelpWhenAvailable(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh not available")
	}

	cmd := exec.Command(pwsh, "-NoProfile", "-File", "install.ps1", "-Help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.ps1 -Help failed: %v\n%s", err, output)
	}
	for _, want := range []string{"-Version", "-InstallDir", "-BaseUrl"} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("install.ps1 help missing %q:\n%s", want, output)
		}
	}
}

func TestReleaseWorkflowPublishesInstallAssets(t *testing.T) {
	workflow, err := os.ReadFile(filepath.Join("..", ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)

	for _, want := range []string{
		"cp scripts/install.sh dist/install.sh",
		"cp scripts/install.ps1 dist/install.ps1",
		"juleson-${OS}-${ARCH}.tar.gz",
		"jules-mcp-${OS}-${ARCH}.tar.gz",
		"juleson-${OS}-${ARCH}.zip",
		"jules-mcp-${OS}-${ARCH}.zip",
		"- goos: windows\n            goarch: arm64",
		"github.com/SamyRai/juleson@${{ needs.validate.outputs.version }}",
		"if: github.event_name == 'workflow_dispatch'",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("release workflow missing %q", want)
		}
	}
}

func TestInstallationDocsReferenceInstallableAssets(t *testing.T) {
	files := []string{
		filepath.Join("..", "README.md"),
		filepath.Join("..", "docs", "INSTALLATION_GUIDE.md"),
		"README.md",
	}
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(contents)
		if strings.Contains(text, "cmd/juleson-mcp") {
			t.Fatalf("%s still references removed cmd/juleson-mcp path", file)
		}
		if strings.Contains(text, "Go 1.23+") || strings.Contains(text, "Go 1.24+") || strings.Contains(text, "Go 1.23 or higher") || strings.Contains(text, "Go 1.24 or higher") {
			t.Fatalf("%s still references an outdated Go prerequisite", file)
		}
	}
}

func installerPlatform() (goos, goarch string, ok bool) {
	switch runtime.GOOS {
	case "linux", "darwin":
		goos = runtime.GOOS
	default:
		return "", "", false
	}

	switch runtime.GOARCH {
	case "amd64", "arm64":
		goarch = runtime.GOARCH
	default:
		return "", "", false
	}
	return goos, goarch, true
}

func writeTarGz(path, name string, contents []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	header := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(contents)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tarWriter, bytes.NewReader(contents))
	return err
}
