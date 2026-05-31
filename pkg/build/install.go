package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type InstallResult struct {
	InstallDir string
	Installed  []string
	Failed     []string
}

type Installer struct {
	binDir   string
	binaries []string
}

func NewInstaller(binDir string, binaries []string) *Installer {
	return &Installer{binDir: binDir, binaries: binaries}
}

func (i *Installer) GetInstallPath() (string, error) {
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		return gobin, nil
	}
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", err
	}
	gopath := strings.TrimSpace(string(out))
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}
	return filepath.Join(gopath, "bin"), nil
}

func (i *Installer) Install(ctx context.Context) (*InstallResult, error) {
	path, err := i.GetInstallPath()
	if err != nil {
		return nil, err
	}
	return i.InstallTo(ctx, path)
}

func (i *Installer) InstallTo(ctx context.Context, installDir string) (*InstallResult, error) {
	_ = ctx
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, err
	}
	result := &InstallResult{InstallDir: installDir}
	for _, binary := range i.binaries {
		src := filepath.Join(i.binDir, binary)
		if runtime.GOOS == "windows" && !strings.HasSuffix(src, ".exe") {
			src += ".exe"
		}
		dst := filepath.Join(installDir, filepath.Base(src))
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(dst, data, 0755); err != nil {
			return nil, err
		}
		result.Installed = append(result.Installed, dst)
	}
	return result, nil
}

func (i *Installer) Uninstall(ctx context.Context) error {
	installDir, err := i.GetInstallPath()
	if err != nil {
		return err
	}
	return i.UninstallFrom(ctx, installDir)
}

func (i *Installer) UninstallFrom(ctx context.Context, installDir string) error {
	_ = ctx
	for _, binary := range i.binaries {
		binPath := filepath.Join(installDir, binary)
		if runtime.GOOS == "windows" && !strings.HasSuffix(binPath, ".exe") {
			binPath += ".exe"
		}
		if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (i *Installer) IsInPath(dir string) bool {
	for _, pathDir := range filepath.SplitList(os.Getenv("PATH")) {
		if pathDir == dir {
			return true
		}
	}
	return false
}
