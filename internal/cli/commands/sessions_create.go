package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/internal/sessionops"
)

func createSession(cfg *config.Config, sourceID string, prompt string, options CreateSessionOptions) error {
	julesClient := newJulesClient(cfg)
	ctx := context.Background()
	if options.PromptFile != "" {
		loadedPrompt, err := loadPromptFile(options.PromptFile)
		if err != nil {
			return err
		}
		prompt = loadedPrompt
	}
	sourceName := sessionops.NormalizeSourceID(sourceID)
	if !options.NoSource && sourceID == "." {
		source, err := julesops.InferSourceFromGitRemote(ctx, julesClient, ".")
		if err != nil {
			return err
		}
		sourceName = source.Name
	}

	fmt.Printf("🚀 Creating new Jules session...\n")
	if options.NoSource {
		fmt.Printf("Source: repoless\n")
	} else {
		fmt.Printf("Source: %s\n", sourceName)
	}
	fmt.Printf("Prompt: %s\n\n", prompt)

	req, err := sessionops.BuildCreateSessionRequest(sessionops.CreateSessionRequestOptions{
		Prompt:              prompt,
		Source:              sourceName,
		NoSource:            options.NoSource,
		Title:               options.Title,
		StartingBranch:      options.StartingBranch,
		RequirePlanApproval: options.RequirePlanApproval,
		AutomationMode:      options.AutomationMode,
	})
	if err == sessionops.ErrStartingBranchRequiresSource {
		return fmt.Errorf("--starting-branch requires a source-backed session")
	}
	if err != nil {
		return err
	}

	session, err := julesClient.CreateSession(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("✅ Session created successfully!\n\n")
	fmt.Printf("📊 Session Details:\n")
	fmt.Printf("ID: %s\n", session.ID)
	fmt.Printf("Title: %s\n", session.Title)
	fmt.Printf("State: %s\n", session.State)
	fmt.Printf("Created: %s\n", session.CreateTime)
	if session.URL != "" {
		fmt.Printf("URL: %s\n", session.URL)
	}

	fmt.Printf("\n💡 Jules is now working on your request. Monitor progress at: %s\n", session.URL)
	fmt.Printf("💡 Use 'juleson sessions get %s' to check status and activities\n", session.ID)

	return nil
}

func batchCreateSessions(cfg *config.Config, sourceID, taskFileOrPrompt string, options BatchSessionOptions) error {
	if options.Parallel < 1 || options.Parallel > 5 {
		return fmt.Errorf("--parallel must be between 1 and 5")
	}

	prompt, err := loadPromptArgument(taskFileOrPrompt)
	if err != nil {
		return err
	}

	julesClient := newJulesClient(cfg)
	ctx := context.Background()
	sourceName := sessionops.NormalizeSourceID(sourceID)
	if options.BatchID == "" {
		options.BatchID = "batch-" + time.Now().UTC().Format("20060102150405")
	}
	if options.GroupTitle == "" {
		options.GroupTitle = options.Title
	}

	fmt.Printf("🚀 Creating %d parallel Jules session(s)\n", options.Parallel)
	fmt.Printf("Batch ID: %s\n", options.BatchID)
	if options.GroupTitle != "" {
		fmt.Printf("Group title: %s\n", options.GroupTitle)
	}
	fmt.Printf("Source: %s\n", sourceName)
	fmt.Printf("Plan approval: required\n")
	fmt.Println(strings.Repeat("=", 60))

	for i := 1; i <= options.Parallel; i++ {
		title := options.Title
		if title != "" && options.Parallel > 1 {
			title = fmt.Sprintf("%s (%d/%d)", title, i, options.Parallel)
		}
		batchPrompt := fmt.Sprintf("Batch ID: %s\n", options.BatchID)
		if options.GroupTitle != "" {
			batchPrompt += fmt.Sprintf("Group title: %s\n", options.GroupTitle)
		}
		batchPrompt += fmt.Sprintf("Parallel run: %d/%d\n\n%s", i, options.Parallel, prompt)
		req, err := sessionops.BuildCreateSessionRequest(sessionops.CreateSessionRequestOptions{
			Prompt:              batchPrompt,
			Title:               title,
			RequirePlanApproval: true,
			AutomationMode:      options.AutomationMode,
			Source:              sourceName,
			StartingBranch:      options.StartingBranch,
		})
		if err != nil {
			return err
		}

		session, err := julesClient.CreateSession(ctx, req)
		if err != nil {
			return fmt.Errorf("created %d/%d sessions before failure: %w", i-1, options.Parallel, err)
		}
		fmt.Printf("%d. %s", i, session.ID)
		if session.URL != "" {
			fmt.Printf(" - %s", session.URL)
		}
		fmt.Println()
	}

	return nil
}
