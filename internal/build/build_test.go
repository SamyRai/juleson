package build

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuildWithResultUsesOptionalBuildSettings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses a POSIX shell script")
	}

	tempDir := t.TempDir()
	fakeBin := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(fakeBin, 0755); err != nil {
		t.Fatal(err)
	}

	argsFile := filepath.Join(tempDir, "args.txt")
	envFile := filepath.Join(tempDir, "env.txt")
	fakeGo := filepath.Join(fakeBin, "go")
	script := `#!/bin/sh
printf '%s\n' "$@" > "$JULESON_TEST_ARGS"
printf 'GOOS=%s\nGOARCH=%s\nCGO_ENABLED=%s\n' "$GOOS" "$GOARCH" "$CGO_ENABLED" > "$JULESON_TEST_ENV"
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    mkdir -p "$(dirname "$1")"
    printf 'binary' > "$1"
    exit 0
  fi
  shift
done
exit 1
`
	if err := os.WriteFile(fakeGo, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("JULESON_TEST_ARGS", argsFile)
	t.Setenv("JULESON_TEST_ENV", envFile)

	config := DefaultConfig("juleson-test", "./cmd/juleson")
	config.OutputDir = filepath.Join(tempDir, "out")
	config.GOOS = "linux"
	config.GOARCH = "amd64"
	config.BuildFlags = []string{"-mod=readonly"}
	config.Tags = []string{"netgo", "release"}
	config.TrimPath = true
	config.CGOConfigured = true
	config.CGOEnabled = false
	config.LDFlags = []string{"-s", "-w"}

	result := NewBuilder(config).BuildWithResult(context.Background())
	if !result.Success {
		t.Fatalf("expected build success, got %v", result.Error)
	}

	argsBytes, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Split(strings.TrimSpace(string(argsBytes)), "\n")
	wantArgs := []string{
		"build",
		"-mod=readonly",
		"-tags",
		"netgo,release",
		"-trimpath",
		"-ldflags",
		"-s -w",
		"-o",
		result.OutputPath,
		"./cmd/juleson",
	}
	if strings.Join(args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("unexpected args:\nwant %#v\ngot  %#v", wantArgs, args)
	}

	envBytes, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatal(err)
	}
	env := string(envBytes)
	for _, want := range []string{"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0"} {
		if !strings.Contains(env, want) {
			t.Fatalf("expected env to contain %q, got %q", want, env)
		}
	}
}

func TestInstallerUninstallFromAndIsInPath(t *testing.T) {
	tempDir := t.TempDir()
	binary := "juleson-test"
	binaryPath := filepath.Join(tempDir, binary)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	if err := os.WriteFile(binaryPath, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	installer := NewInstaller("bin", []string{binary})
	t.Setenv("PATH", tempDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if !installer.IsInPath(tempDir) {
		t.Fatalf("expected %s to be in PATH", tempDir)
	}
	if err := installer.UninstallFrom(context.Background(), tempDir); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		t.Fatalf("expected binary to be removed, stat err: %v", err)
	}
}
