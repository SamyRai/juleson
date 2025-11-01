package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

			// Validate project path exists
			absPath, err := filepath.Abs(projectPath)
			if err != nil {
				return fmt.Errorf("invalid project path: %w", err)
			}

			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				return fmt.Errorf("project path does not exist: %s", absPath)
			}

			// Check if it's a git repository
			gitDir := filepath.Join(absPath, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				return fmt.Errorf("not a git repository: %s", absPath)
			}

			fmt.Printf("🔄 Syncing project with remote '%s'...\n", remote)

			// Pull changes if requested
			if pull {
				fmt.Printf("📥 Pulling changes from %s/%s...\n", remote, branch)
				pullCmd := exec.Command("git", "pull", remote, branch)
				pullCmd.Dir = absPath
				pullCmd.Stdout = os.Stdout
				pullCmd.Stderr = os.Stderr

				if err := pullCmd.Run(); err != nil {
					return fmt.Errorf("failed to pull changes: %w", err)
				}
				fmt.Println("✅ Pull completed successfully")
			}

			// Push changes if requested
			if push {
				fmt.Printf("📤 Pushing changes to %s/%s...\n", remote, branch)
				pushCmd := exec.Command("git", "push", remote, branch)
				pushCmd.Dir = absPath
				pushCmd.Stdout = os.Stdout
				pushCmd.Stderr = os.Stderr

				if err := pushCmd.Run(); err != nil {
					return fmt.Errorf("failed to push changes: %w", err)
				}
				fmt.Println("✅ Push completed successfully")
			}

			// Fetch remote changes if neither pull nor push
			if !pull && !push {
				fmt.Printf("📡 Fetching changes from %s...\n", remote)
				fetchCmd := exec.Command("git", "fetch", remote)
				fetchCmd.Dir = absPath
				fetchCmd.Stdout = os.Stdout
				fetchCmd.Stderr = os.Stderr

				if err := fetchCmd.Run(); err != nil {
					return fmt.Errorf("failed to fetch changes: %w", err)
				}
				fmt.Println("✅ Fetch completed successfully")
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
