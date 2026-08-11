package builder

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/pkg/build"
)

func (s *Service) CleanArtifacts(ctx context.Context, options CleanOptions) ([]string, error) {
	cleaner := build.NewCleaner(s.config.BinDir, []string{s.config.CoverageFile, s.config.CoverageHTML})
	cleaned := make([]string, 0, 4)

	if options.All {
		if err := cleaner.CleanAll(ctx); err != nil {
			return cleaned, err
		}
		return []string{"artifacts", "build cache", "module cache", "test cache"}, nil
	}

	if err := cleaner.Clean(ctx); err != nil {
		return cleaned, err
	}
	cleaned = append(cleaned, "artifacts")

	if options.Cache {
		if err := cleaner.CleanCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "build cache")
	}
	if options.ModCache {
		if err := cleaner.CleanModCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "module cache")
	}
	if options.TestCache {
		if err := cleaner.CleanTestCache(ctx); err != nil {
			return cleaned, err
		}
		cleaned = append(cleaned, "test cache")
	}

	return cleaned, nil
}

// Clean removes generated build artifacts and coverage files.
func (s *Service) Clean(ctx context.Context) error {
	_, err := s.CleanArtifacts(ctx, CleanOptions{})
	return err
}

func (s *Service) RunModuleMaintenance(ctx context.Context, operation string, packages ...string) error {
	manager := build.NewModuleManager()
	switch operation {
	case "tidy":
		return manager.Tidy(ctx)
	case "download":
		return manager.Download(ctx)
	case "verify":
		return manager.Verify(ctx)
	case "vendor":
		return manager.Vendor(ctx)
	case "graph":
		return manager.Graph(ctx)
	case "why":
		return manager.Why(ctx, packages...)
	default:
		return fmt.Errorf("unknown module operation: %s", operation)
	}
}

func (s *Service) InstallWithResult(ctx context.Context, options InstallOptions) (*build.InstallResult, error) {
	if !options.SkipBuild {
		if _, err := s.BuildWithResults(ctx, BuildOptions{Target: "all", Version: "dev"}); err != nil {
			return nil, err
		}
	}

	installer := build.NewInstaller(s.config.BinDir, []string{s.config.BinaryCLI, s.config.BinaryAlias})
	if options.Path != "" {
		return installer.InstallTo(ctx, options.Path)
	}
	return installer.Install(ctx)
}
