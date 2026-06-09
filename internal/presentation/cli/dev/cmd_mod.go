package dev

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SamyRai/juleson/internal/logger"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/spf13/cobra"
)

func (h *CommandHandler) CleanCmd() *cobra.Command {
	var (
		all       bool
		cache     bool
		modCache  bool
		testCache bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean build artifacts",
		Long:  "Clean build artifacts, caches, and generated files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			slog.Info("Cleaning...")

			_, err := h.svc.CleanArtifacts(ctx, builder.CleanOptions{
				All:       all,
				Cache:     cache,
				ModCache:  modCache,
				TestCache: testCache,
			})
			if err != nil {
				fmt.Printf("❌ Clean failed: %v\n", err)
				return err
			}

			logger.Success(slog.Default(), "Cleaned successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Clean everything including caches")
	cmd.Flags().BoolVar(&cache, "cache", false, "Clean build cache")
	cmd.Flags().BoolVar(&modCache, "modcache", false, "Clean module cache")
	cmd.Flags().BoolVar(&testCache, "testcache", false, "Clean test cache")

	return cmd
}

func (h *CommandHandler) ModCmd() *cobra.Command {
	modCmd := &cobra.Command{
		Use:   "mod",
		Short: "Module maintenance",
		Long:  "Go module maintenance commands",
	}

	modCmd.AddCommand(&cobra.Command{
		Use:   "tidy",
		Short: "Tidy dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Tidying dependencies...")
			if err := h.svc.RunModuleMaintenance(context.Background(), "tidy"); err != nil {
				return err
			}
			logger.Success(slog.Default(), "Dependencies tidied")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "download",
		Short: "Download dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Downloading dependencies...")
			if err := h.svc.RunModuleMaintenance(context.Background(), "download"); err != nil {
				return err
			}
			logger.Success(slog.Default(), "Dependencies downloaded")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "verify",
		Short: "Verify dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Verifying dependencies...")
			if err := h.svc.RunModuleMaintenance(context.Background(), "verify"); err != nil {
				return err
			}
			logger.Success(slog.Default(), "Dependencies verified")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "vendor",
		Short: "Vendor dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Vendoring dependencies...")
			if err := h.svc.RunModuleMaintenance(context.Background(), "vendor"); err != nil {
				return err
			}
			logger.Success(slog.Default(), "Dependencies vendored")
			return nil
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "graph",
		Short: "Print dependency graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.svc.RunModuleMaintenance(context.Background(), "graph")
		},
	})

	modCmd.AddCommand(&cobra.Command{
		Use:   "why [packages...]",
		Short: "Explain why packages are needed",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.svc.RunModuleMaintenance(context.Background(), "why", args...)
		},
	})

	return modCmd
}
