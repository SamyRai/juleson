package sessions

import (
	"github.com/spf13/cobra"
)

// ListCmd returns the command for listing sessions.
func (h *CommandHandler) ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all Jules sessions with their current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessions(h.cfg)
		},
	}
}

// ApproveCmd returns the command for approving a session plan.
func (h *CommandHandler) ApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve [session-id]",
		Short: "Approve a plan in a session",
		Long:  "Approve a plan that is waiting for approval in the specified session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return approveSessionPlan(h.cfg, args[0])
		},
	}
}

// StatusCmd returns the command for showing session status.
func (h *CommandHandler) StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Long:  "Show a summary of current session statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionStatus(h.cfg)
		},
	}
}

// GetCmd returns the command for getting session details.
func (h *CommandHandler) GetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [session-id]",
		Short: "Get session details and activities",
		Long:  "Get detailed information about a specific session including all activities",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getSessionDetails(h.cfg, args[0])
		},
	}
}

// PlansCmd returns the command for showing session plans.
func (h *CommandHandler) PlansCmd() *cobra.Command {
	var (
		plansLatest bool
		plansJSON   bool
	)

	plansCmd := &cobra.Command{
		Use:   "plans [session-id]",
		Short: "Show generated session plans",
		Long:  "Show generated Jules plans with activity IDs, plan IDs, approval state, and full step details.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionPlans(h.cfg, args[0], plansLatest, plansJSON)
		},
	}
	plansCmd.Flags().BoolVar(&plansLatest, "latest", false, "Show only the newest generated plan")
	plansCmd.Flags().BoolVar(&plansJSON, "json", false, "Print machine-readable JSON")

	return plansCmd
}

// MessageCmd returns the command for sending a message to a session.
func (h *CommandHandler) MessageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "message [session-id] [message]",
		Short: "Send a message to a session",
		Long:  "Send a message to Jules within a session to request changes or provide feedback",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendSessionMessage(h.cfg, args[0], args[1])
		},
	}
}

// DeleteCmd returns the command for deleting a session.
func (h *CommandHandler) DeleteCmd() *cobra.Command {
	var deleteForce bool

	deleteCmd := &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Delete a session",
		Long:  "Delete a Jules session. Without --force, type the session ID to confirm.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteSession(h.cfg, args[0], deleteForce)
		},
	}
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Delete without interactive confirmation")

	return deleteCmd
}

// OutputsCmd returns the command for showing session outputs.
func (h *CommandHandler) OutputsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outputs [session-id]",
		Short: "Show session outputs",
		Long:  "Show Jules session outputs such as created pull requests.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionOutputs(h.cfg, args[0])
		},
	}
}

// DownloadCmd returns the command for downloading session artifacts.
func (h *CommandHandler) DownloadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "download [session-id] [output-dir]",
		Short: "Download all artifacts from a session",
		Long:  "Download all artifacts (patches, outputs, media) from all activities in a session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 1 {
				outputDir = args[1]
			}
			return downloadSessionArtifacts(h.cfg, args[0], outputDir)
		},
	}
}

// DownloadActivityCmd returns the command for downloading activity artifacts.
func (h *CommandHandler) DownloadActivityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "download-activity [session-id] [activity-id] [output-dir]",
		Short: "Download artifacts from a specific activity",
		Long:  "Download all artifacts from a specific activity within a session",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := "."
			if len(args) > 2 {
				outputDir = args[2]
			}
			return downloadActivityArtifacts(h.cfg, args[0], args[1], outputDir)
		},
	}
}

// PreviewCmd returns the command for previewing session artifacts.
func (h *CommandHandler) PreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview [session-id]",
		Short: "Preview all artifacts from a session",
		Long:  "Display artifacts (diffs, outputs, media info) from all activities in a session without downloading",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewSessionArtifacts(h.cfg, args[0])
		},
	}
}

// PreviewActivityCmd returns the command for previewing activity artifacts.
func (h *CommandHandler) PreviewActivityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview-activity [session-id] [activity-id]",
		Short: "Preview artifacts from a specific activity",
		Long:  "Display artifacts from a specific activity within a session without downloading",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewActivityArtifacts(h.cfg, args[0], args[1])
		},
	}
}

// AutocleanCmd returns the command for autocleaning sessions.
func (h *CommandHandler) AutocleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "autoclean",
		Short: "Automatically clean up merged sessions globally",
		Long:  "List all COMPLETED sessions, clone their repo to a tmpfs to verify if the patch is merged, and delete the remote session if it is.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return autocleanSessions(h.cfg)
		},
	}
}
