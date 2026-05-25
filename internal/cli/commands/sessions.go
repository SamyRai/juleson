package commands

import (
	"fmt"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/spf13/cobra"
)

// NewSessionsCommand creates the sessions command
func NewSessionsCommand(cfg *config.Config) *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Jules sessions",
		Long:  "List, monitor, and manage Jules AI coding sessions",
	}

	// List sessions
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all Jules sessions with their current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessions(cfg)
		},
	})

	// Create session
	var createNoSource bool
	createOptions := CreateSessionOptions{}
	createCmd := &cobra.Command{
		Use:   "create [source-id] [prompt]",
		Short: "Create a new session",
		Long:  "Create a new Jules session with a repository source, or pass --no-source for a repoless session",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := createOptions
			options.NoSource = createNoSource
			if createNoSource {
				if options.PromptFile != "" {
					if len(args) != 0 {
						return fmt.Errorf("--no-source with --prompt-file does not accept positional arguments")
					}
					return createSession(cfg, "", "", options)
				}
				if len(args) != 1 {
					return fmt.Errorf("--no-source accepts exactly one prompt argument, or use --prompt-file")
				}
				return createSession(cfg, "", args[0], options)
			}
			if options.PromptFile != "" {
				if len(args) != 1 {
					return fmt.Errorf("--prompt-file requires exactly one source ID argument")
				}
				return createSession(cfg, args[0], "", options)
			}
			if len(args) != 2 {
				return fmt.Errorf("provide source ID and prompt, use --prompt-file, or pass --no-source with a prompt")
			}
			return createSession(cfg, args[0], args[1], options)
		},
	}
	createCmd.Flags().BoolVar(&createNoSource, "no-source", false, "Create a repoless session without sourceContext")
	createCmd.Flags().StringVar(&createOptions.PromptFile, "prompt-file", "", "Read the session prompt from a file")
	createCmd.Flags().StringVar(&createOptions.Title, "title", "", "Optional session title")
	createCmd.Flags().StringVar(&createOptions.StartingBranch, "starting-branch", "", "Starting branch for source-backed sessions")
	createCmd.Flags().BoolVar(&createOptions.RequirePlanApproval, "require-plan-approval", false, "Require explicit plan approval before Jules starts work")
	createCmd.Flags().StringVar(&createOptions.AutomationMode, "automation-mode", "", "Automation mode such as AUTO_CREATE_PR")
	sessionsCmd.AddCommand(createCmd)

	var (
		watchInterval         string
		watchTimeout          string
		watchFollowActivities bool
		watchSince            string
		watchCursorOutput     string
	)
	watchCmd := &cobra.Command{
		Use:   "watch [session-id]",
		Short: "Watch a session until completion or user action",
		Long:  "Poll a Jules session until it completes, fails, or needs user action such as plan approval or feedback",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return watchSession(cfg, args[0], watchInterval, watchTimeout, watchFollowActivities, watchSince, watchCursorOutput)
		},
	}
	watchCmd.Flags().StringVar(&watchInterval, "interval", "30s", "Polling interval")
	watchCmd.Flags().StringVar(&watchTimeout, "timeout", "30m", "Maximum watch duration")
	watchCmd.Flags().BoolVar(&watchFollowActivities, "follow-activities", false, "Print recent activity updates while watching")
	watchCmd.Flags().StringVar(&watchSince, "since", "", "Only print activities at or after this RFC3339 createTime cursor")
	watchCmd.Flags().StringVar(&watchCursorOutput, "cursor-output", "", "Write the latest activity createTime cursor to this file")
	sessionsCmd.AddCommand(watchCmd)

	// Approve plan
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "approve [session-id]",
		Short: "Approve a plan in a session",
		Long:  "Approve a plan that is waiting for approval in the specified session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return approveSessionPlan(cfg, args[0])
		},
	})

	// Show session status
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Long:  "Show a summary of current session statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionStatus(cfg)
		},
	})

	// Get session details with activities
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "get [session-id]",
		Short: "Get session details and activities",
		Long:  "Get detailed information about a specific session including all activities",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSessionDetails(cfg, args[0])
		},
	})

	// Send message to session
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "message [session-id] [message]",
		Short: "Send a message to a session",
		Long:  "Send a message to Jules within a session to request changes or provide feedback",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendSessionMessage(cfg, args[0], args[1])
		},
	})

	var deleteForce bool
	deleteCmd := &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Delete a session",
		Long:  "Delete a Jules session. Without --force, type the session ID to confirm.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteSession(cfg, args[0], deleteForce)
		},
	}
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Delete without interactive confirmation")
	sessionsCmd.AddCommand(deleteCmd)

	var (
		applyConfirm           bool
		applyAllowDirty        bool
		applyActivityID        string
		applyArtifactIndex     int
		applyAllowBaseMismatch bool
	)
	applyCmd := &cobra.Command{
		Use:   "apply [session-id] [project-path]",
		Short: "Preview or apply session patches",
		Long:  "Preview session patches by default. Pass --confirm to apply after a clean-worktree check.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return applySessionChanges(cfg, args[0], args[1], ApplySessionOptions{
				Confirm:           applyConfirm,
				AllowDirty:        applyAllowDirty,
				ActivityID:        applyActivityID,
				ArtifactIndex:     applyArtifactIndex,
				HasArtifactIndex:  cmd.Flags().Changed("artifact-index"),
				AllowBaseMismatch: applyAllowBaseMismatch,
			})
		},
	}
	applyCmd.Flags().BoolVar(&applyConfirm, "confirm", false, "Actually apply patches instead of dry-running")
	applyCmd.Flags().BoolVar(&applyAllowDirty, "allow-dirty", false, "Allow applying patches when the target worktree has local changes")
	applyCmd.Flags().StringVar(&applyActivityID, "activity-id", "", "Apply only changes from this activity ID or resource name")
	applyCmd.Flags().IntVar(&applyArtifactIndex, "artifact-index", 0, "Apply only this artifact index within the selected scope")
	applyCmd.Flags().BoolVar(&applyAllowBaseMismatch, "allow-base-mismatch", false, "Allow applying when a patch baseCommitId differs from target HEAD")
	sessionsCmd.AddCommand(applyCmd)

	var (
		batchParallel       int
		batchTitle          string
		batchID             string
		batchGroupTitle     string
		batchStartingBranch string
		batchAutomationMode string
	)
	batchCmd := &cobra.Command{
		Use:   "batch [source-id] [task-file-or-prompt]",
		Short: "Create parallel sessions for one task",
		Long:  "Create 1-5 parallel Jules sessions for the same source and task. Batch sessions require plan approval by default.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return batchCreateSessions(cfg, args[0], args[1], BatchSessionOptions{
				Parallel:       batchParallel,
				Title:          batchTitle,
				BatchID:        batchID,
				GroupTitle:     batchGroupTitle,
				StartingBranch: batchStartingBranch,
				AutomationMode: batchAutomationMode,
			})
		},
	}
	batchCmd.Flags().IntVar(&batchParallel, "parallel", 2, "Number of parallel sessions to create (1-5)")
	batchCmd.Flags().StringVar(&batchTitle, "title", "", "Optional title prefix for created sessions")
	batchCmd.Flags().StringVar(&batchID, "batch-id", "", "Optional batch identifier to include in prompts and output")
	batchCmd.Flags().StringVar(&batchGroupTitle, "group-title", "", "Optional group title to include in prompts and output")
	batchCmd.Flags().StringVar(&batchStartingBranch, "starting-branch", "", "Starting branch for the source-backed sessions")
	batchCmd.Flags().StringVar(&batchAutomationMode, "automation-mode", "", "Automation mode such as AUTO_CREATE_PR")
	sessionsCmd.AddCommand(batchCmd)

	artifactsCmd := &cobra.Command{
		Use:   "artifacts",
		Short: "Inspect session artifact manifests",
	}
	artifactsCmd.AddCommand(&cobra.Command{
		Use:   "list [session-id]",
		Short: "List session artifact manifests",
		Long:  "List documented artifacts with activity IDs, indexes, patch metadata, media MIME types, and bash exit codes.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessionArtifacts(cfg, args[0])
		},
	})
	sessionsCmd.AddCommand(artifactsCmd)

	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "outputs [session-id]",
		Short: "Show session outputs",
		Long:  "Show Jules session outputs such as created pull requests.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionOutputs(cfg, args[0])
		},
	})

	// Download all session artifacts
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "download [session-id] [output-dir]",
		Short: "Download all artifacts from a session",
		Long:  "Download all artifacts (patches, outputs, media) from all activities in a session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 1 {
				outputDir = args[1]
			}
			return downloadSessionArtifacts(cfg, args[0], outputDir)
		},
	})

	// Download artifacts from specific activity
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "download-activity [session-id] [activity-id] [output-dir]",
		Short: "Download artifacts from a specific activity",
		Long:  "Download all artifacts from a specific activity within a session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 2 {
				outputDir = args[2]
			}
			return downloadActivityArtifacts(cfg, args[0], args[1], outputDir)
		},
	})

	// Preview all session artifacts
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "preview [session-id]",
		Short: "Preview all artifacts from a session",
		Long:  "Display artifacts (diffs, outputs, media info) from all activities in a session without downloading",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewSessionArtifacts(cfg, args[0])
		},
	})

	// Preview artifacts from specific activity
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "preview-activity [session-id] [activity-id]",
		Short: "Preview artifacts from a specific activity",
		Long:  "Display artifacts from a specific activity within a session without downloading",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewActivityArtifacts(cfg, args[0], args[1])
		},
	})

	return sessionsCmd
}
