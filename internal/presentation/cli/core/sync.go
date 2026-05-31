package core

import (
	"context"
	"fmt"
	"os"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/spf13/cobra"
)

// NewSyncCommand creates the sync command
func NewSyncCommand() *cobra.Command {
	var (
		branch string
		pull   bool
		push   bool
	)

	cmd := &cobra.Command{
		Use:   "sync [project-path] [remote]",
		Short: "Sync project with remote repository",
		Long:  "Sync project changes with remote repository using Git",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := args[0]
			remote := args[1]

			fmt.Printf("🔄 Syncing project with remote '%s'...\n", remote)

			// Pull changes if requested
			if pull {
				fmt.Printf("📥 Pulling changes from %s/%s...\n", remote, branch)
			}

			// Push changes if requested
			if push {
				fmt.Printf("📤 Pushing changes to %s/%s...\n", remote, branch)
			}

			// Fetch remote changes if neither pull nor push
			if !pull && !push {
				fmt.Printf("📡 Fetching changes from %s...\n", remote)
			}

			if err := julesops.SyncGitRepository(context.Background(), julesops.GitSyncOptions{
				ProjectPath: projectPath,
				Remote:      remote,
				Branch:      branch,
				Pull:        pull,
				Push:        push,
				Stdout:      os.Stdout,
				Stderr:      os.Stderr,
			}); err != nil {
				return err
			}

			fmt.Println("✅ Sync completed successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to sync")
	cmd.Flags().BoolVar(&pull, "pull", false, "Pull changes from remote")
	cmd.Flags().BoolVar(&push, "push", false, "Push changes to remote")

	return cmd
}
