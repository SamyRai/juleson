package commands

import (
	"context"
	"fmt"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/sessionops"
)

func autocleanSessions(cfg *config.Config) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()

	fmt.Println("🔍 Fetching sessions...")
	sessionsResponse, err := julesClient.Sessions().List(ctx, &jules.ListSessionsOptions{
		PageSize: 100,
	})
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var completedSessions []jules.Session
	for _, session := range sessionsResponse.Sessions {
		if session.State == jules.SessionStateCompleted {
			completedSessions = append(completedSessions, session)
		}
	}

	if len(completedSessions) == 0 {
		fmt.Println("No completed sessions found to clean.")
		return nil
	}

	fmt.Printf("🧹 Found %d COMPLETED session(s) to verify for cleanup...\n\n", len(completedSessions))

	for _, session := range completedSessions {
		fmt.Printf("▶️  Verifying session %s (%s)...\n", session.ID, session.Title)

		merged, err := sessionops.VerifySessionMerged(ctx, julesClient, session.ID, session.SourceContext)
		if err != nil {
			fmt.Printf("   ⚠️  Could not verify session: %v\n\n", err)
			continue
		}

		if merged {
			fmt.Printf("   ✅ Patch is verified as MERGED! Deleting remote session...\n")
			if delErr := julesClient.Sessions().Delete(ctx, session.ID); delErr != nil {
				fmt.Printf("   ❌ Failed to delete session %s: %v\n", session.ID, delErr)
			} else {
				fmt.Printf("   🗑️  Deleted session %s.\n", session.ID)
			}
		} else {
			fmt.Printf("   ⏳ Patch is NOT perfectly merged (or was modified post-merge). Leaving session untouched.\n")
		}
		fmt.Println()
	}

	fmt.Println("🎉 Autoclean complete!")
	return nil
}
