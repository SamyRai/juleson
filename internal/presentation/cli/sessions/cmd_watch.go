package sessions

import (
	"github.com/spf13/cobra"
)

// WatchCmd returns the command for watching a session.
func (h *CommandHandler) WatchCmd() *cobra.Command {
	var (
		watchInterval         string
		watchTimeout          string
		watchFollowActivities bool
		watchSince            string
		watchCursorOutput     string
		watchInitialState     string
		watchOnStatusChange   bool
		watchOnAgentMessage   bool
		watchWakePolicy       string
	)

	watchCmd := &cobra.Command{
		Use:   "watch [session-id]",
		Short: "Watch a session until completion or user action",
		Long:  "Poll a Jules session until it completes, fails, or needs user action such as plan approval or feedback",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return watchSession(h.cfg, args[0], watchInterval, watchTimeout, watchFollowActivities, watchSince, watchCursorOutput, watchInitialState, watchOnStatusChange, watchOnAgentMessage, watchWakePolicy)
		},
	}

	watchCmd.Flags().StringVar(&watchInterval, "interval", "30s", "Polling interval")
	watchCmd.Flags().StringVar(&watchTimeout, "timeout", "30m", "Maximum watch duration")
	watchCmd.Flags().BoolVar(&watchFollowActivities, "follow-activities", false, "Print recent activity updates while watching")
	watchCmd.Flags().StringVar(&watchSince, "since", "", "Only print activities at or after this RFC3339 createTime cursor")
	watchCmd.Flags().StringVar(&watchCursorOutput, "cursor-output", "", "Write the latest activity createTime cursor to this file")
	watchCmd.Flags().StringVar(&watchInitialState, "initial-state", "", "Known current session state for --wake-on-status-change")
	watchCmd.Flags().BoolVar(&watchOnStatusChange, "wake-on-status-change", false, "Stop when the session state changes from --initial-state or the first observed state")
	watchCmd.Flags().BoolVar(&watchOnAgentMessage, "wake-on-agent-message", false, "Stop when a new Jules-authored message activity appears")
	watchCmd.Flags().StringVar(&watchWakePolicy, "wake-policy", "actionable", "When to stop watching: actionable, any-status, terminal, or none")

	return watchCmd
}
