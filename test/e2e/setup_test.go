package e2e

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildAndPackage builds juleson and packages it into a tar.gz similar to the release process.
func buildAndPackage(t *testing.T, releaseDir string) {
	// Build the juleson binary
	binPath := filepath.Join(releaseDir, "juleson")
	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/juleson")

	// We need to build for linux/amd64 since our test containers are linux
	buildCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build juleson for linux: %v\n%s", err, out)
	}

	contents, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("failed to read built binary: %v", err)
	}

	// Create juleson tar.gz
	julesonTar := filepath.Join(releaseDir, "juleson-linux-amd64.tar.gz")
	if err := writeTarGz(julesonTar, "juleson", contents); err != nil {
		t.Fatalf("failed to package juleson: %v", err)
	}

	// Create jsn alias tar.gz (same binary)
	jsnTar := filepath.Join(releaseDir, "jsn-linux-amd64.tar.gz")
	if err := writeTarGz(jsnTar, "jsn", contents); err != nil {
		t.Fatalf("failed to package jsn: %v", err)
	}
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

// getInstallerPlatform simulates the platform resolution matching the script.
func getInstallerPlatform() (string, string) {
	// We are compiling targeting linux amd64 for the containers
	return "linux", "amd64"
}
