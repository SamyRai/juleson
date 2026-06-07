package jmcp

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/SamyRai/juleson/internal/jules/workspace"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"github.com/SamyRai/juleson/pkg/build"
	"github.com/SamyRai/juleson/pkg/builder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const ServerName = "juleson"

type ServerOptions struct {
	Config *config.Config
}

func NewServer(options ServerOptions) (*mcp.Server, error) {
	if options.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: core.Version,
	}, nil)
	tools := &toolRegistry{cfg: options.Config}
	tools.register(server)
	return server, nil
}

func RunStdio(ctx context.Context, cfg *config.Config) error {
	server, err := NewServer(ServerOptions{Config: cfg})
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}

type toolRegistry struct {
	cfg *config.Config
}

func (r *toolRegistry) register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "version",
		Description: "Return Juleson build and runtime version information.",
	}, r.version)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "config_status",
		Description: "Report whether Jules and GitHub credentials are configured without exposing secret values.",
	}, r.configStatus)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sources",
		Description: "List connected Jules repository sources.",
	}, r.listSources)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_source",
		Description: "Get one connected Jules source by ID or resource name.",
	}, r.getSource)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List Jules sessions.",
	}, r.listSessions)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session",
		Description: "Get a Jules session by ID.",
	}, r.getSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_session",
		Description: "Create a Jules session. Source-backed sessions require a Jules source ID; repoless sessions set no_source=true.",
	}, r.createSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "approve_session_plan",
		Description: "Approve the current plan for a Jules session. Requires confirm=true.",
	}, r.approvePlan)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_session_message",
		Description: "Send feedback or a follow-up message to a Jules session.",
	}, r.sendMessage)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a Jules session. Requires confirm=true.",
	}, r.deleteSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_activities",
		Description: "List Jules activities for a session.",
	}, r.listActivities)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_activity",
		Description: "Get one Jules activity from a session.",
	}, r.getActivity)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_plans",
		Description: "Return generated plans found in a Jules session's activities.",
	}, r.getPlans)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_session",
		Description: "Build a read-only operator review with plans, outputs, artifacts, patch preview, worktree state, blockers, and next actions.",
	}, r.reviewSession)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_session_artifacts",
		Description: "Return documented artifact manifests for a Jules session.",
	}, r.listArtifacts)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_session_outputs",
		Description: "Return documented Jules session outputs such as pull request links.",
	}, r.getOutputs)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_build",
		Description: "Build Juleson binaries. Target is all, cli, or alias.",
	}, r.devBuild)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_test",
		Description: "Run Juleson Go tests through the builder service.",
	}, r.devTest)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dev_check",
		Description: "Run Juleson quality checks through the builder service. Requires confirm=true because formatting may modify files.",
	}, r.devCheck)
}

func (r *toolRegistry) client() (*jules.Client, error) {
	if r.cfg.Jules.APIKey == "" {
		return nil, fmt.Errorf("jules API key is not configured; set jules.api_key or JULES_API_KEY")
	}
	return core.NewJulesClient(r.cfg), nil
}

func requireConfirm(confirm bool, action string) error {
	if !confirm {
		return fmt.Errorf("%s requires confirm=true", action)
	}
	return nil
}

func wrapAPIError(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("Jules API error during %s: %w", action, err)
}

type emptyInput struct{}

func (r *toolRegistry) version(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, core.VersionInfo, error) {
	return nil, core.GetVersionInfo(), nil
}

type configStatusOutput struct {
	JulesBaseURL          string `json:"jules_base_url"`
	JulesAPIKeyConfigured bool   `json:"jules_api_key_configured"`
	GitHubTokenConfigured bool   `json:"github_token_configured"`
}

func (r *toolRegistry) configStatus(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, configStatusOutput, error) {
	return nil, configStatusOutput{
		JulesAPIKeyConfigured: r.cfg.Jules.APIKey != "",
		JulesBaseURL:          r.cfg.Jules.BaseURL,
		GitHubTokenConfigured: r.cfg.GitHub.Token != "",
	}, nil
}

type listSourcesInput struct {
	Filter   string `json:"filter,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func (r *toolRegistry) listSources(ctx context.Context, _ *mcp.CallToolRequest, in listSourcesInput) (*mcp.CallToolResult, *jules.SourcesResponse, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	response, err := client.Sources().List(ctx, &jules.ListSourcesOptions{PageSize: pageSize, Filter: in.Filter})
	return nil, response, wrapAPIError("list sources", err)
}

type getSourceInput struct {
	SourceID string `json:"source_id" jsonschema:"Jules source ID or resource name"`
}

func (r *toolRegistry) getSource(ctx context.Context, _ *mcp.CallToolRequest, in getSourceInput) (*mcp.CallToolResult, *jules.Source, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	source, err := client.Sources().Get(ctx, in.SourceID)
	return nil, source, wrapAPIError("get source", err)
}

type listSessionsInput struct {
	Filter    string `json:"filter,omitempty"`
	PageToken string `json:"page_token,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
}

func (r *toolRegistry) listSessions(ctx context.Context, _ *mcp.CallToolRequest, in listSessionsInput) (*mcp.CallToolResult, *jules.SessionsResponse, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}
	response, err := client.Sessions().List(ctx, &jules.ListSessionsOptions{
		PageSize:  pageSize,
		PageToken: in.PageToken,
		Filter:    in.Filter,
	})
	return nil, response, wrapAPIError("list sessions", err)
}

type sessionIDInput struct {
	SessionID string `json:"session_id" jsonschema:"Jules session ID"`
}

func (r *toolRegistry) getSession(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, *jules.Session, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	session, err := client.Sessions().Get(ctx, in.SessionID)
	return nil, session, wrapAPIError("get session", err)
}

type createSessionInput struct {
	Prompt              string  `json:"prompt"`
	SourceID            *string `json:"source_id,omitempty"`
	Title               *string `json:"title,omitempty"`
	StartingBranch      *string `json:"starting_branch,omitempty"`
	AutomationMode      *string `json:"automation_mode,omitempty"`
	NoSource            bool    `json:"no_source,omitempty"`
	RequirePlanApproval bool    `json:"require_plan_approval,omitempty"`
}

func (r *toolRegistry) createSession(ctx context.Context, _ *mcp.CallToolRequest, in createSessionInput) (*mcp.CallToolResult, *jules.Session, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	req := &jules.CreateSessionRequest{
		Prompt:              in.Prompt,
		Title:               optionalString(in.Title),
		RequirePlanApproval: in.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(optionalString(in.AutomationMode)),
	}
	if !in.NoSource {
		sourceID := optionalString(in.SourceID)
		if sourceID == "" {
			return nil, nil, fmt.Errorf("source_id is required unless no_source=true")
		}
		req.SourceContext = &jules.SourceContext{
			Source: sourceID,
		}
		if startingBranch := optionalString(in.StartingBranch); startingBranch != "" {
			req.SourceContext.GithubRepoContext = &jules.GithubRepoContext{StartingBranch: startingBranch}
		}
	}
	session, err := client.Sessions().Create(ctx, req)
	return nil, session, wrapAPIError("create session", err)
}

type confirmSessionInput struct {
	SessionID string `json:"session_id"`
	Confirm   bool   `json:"confirm"`
}

type actionOutput struct {
	Message string `json:"message"`
	OK      bool   `json:"ok"`
}

func (r *toolRegistry) approvePlan(ctx context.Context, _ *mcp.CallToolRequest, in confirmSessionInput) (*mcp.CallToolResult, actionOutput, error) {
	if err := requireConfirm(in.Confirm, "approve_session_plan"); err != nil {
		return nil, actionOutput{}, err
	}
	client, err := r.client()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().ApprovePlan(ctx, in.SessionID); err != nil {
		return nil, actionOutput{}, wrapAPIError("approve session plan", err)
	}
	return nil, actionOutput{OK: true, Message: "plan approved"}, nil
}

type sendMessageInput struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func (r *toolRegistry) sendMessage(ctx context.Context, _ *mcp.CallToolRequest, in sendMessageInput) (*mcp.CallToolResult, actionOutput, error) {
	client, err := r.client()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().SendMessage(ctx, in.SessionID, &jules.SendMessageRequest{Prompt: in.Message}); err != nil {
		return nil, actionOutput{}, wrapAPIError("send session message", err)
	}
	return nil, actionOutput{OK: true, Message: "message sent"}, nil
}

func (r *toolRegistry) deleteSession(ctx context.Context, _ *mcp.CallToolRequest, in confirmSessionInput) (*mcp.CallToolResult, actionOutput, error) {
	if err := requireConfirm(in.Confirm, "delete_session"); err != nil {
		return nil, actionOutput{}, err
	}
	client, err := r.client()
	if err != nil {
		return nil, actionOutput{}, err
	}
	if err := client.Sessions().Delete(ctx, in.SessionID); err != nil {
		return nil, actionOutput{}, wrapAPIError("delete session", err)
	}
	return nil, actionOutput{OK: true, Message: "session deleted"}, nil
}

type listActivitiesInput struct {
	SessionID string `json:"session_id"`
	PageSize  int    `json:"page_size,omitempty"`
}

func (r *toolRegistry) listActivities(ctx context.Context, _ *mcp.CallToolRequest, in listActivitiesInput) (*mcp.CallToolResult, *jules.ActivitiesResponse, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	response, err := client.Activities().List(ctx, in.SessionID, &jules.ListActivitiesOptions{PageSize: pageSize})
	return nil, response, wrapAPIError("list activities", err)
}

type getActivityInput struct {
	SessionID  string `json:"session_id"`
	ActivityID string `json:"activity_id"`
}

func (r *toolRegistry) getActivity(ctx context.Context, _ *mcp.CallToolRequest, in getActivityInput) (*mcp.CallToolResult, *jules.Activity, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	activity, err := client.Activities().Get(ctx, in.SessionID, in.ActivityID)
	return nil, activity, wrapAPIError("get activity", err)
}

type getPlansInput struct {
	SessionID  string `json:"session_id"`
	LatestOnly bool   `json:"latest_only,omitempty"`
}

type plansOutput struct {
	SessionID  string                      `json:"session_id"`
	Plans      []julessessions.PlanSummary `json:"plans"`
	TotalCount int                         `json:"total_count"`
}

func (r *toolRegistry) getPlans(ctx context.Context, _ *mcp.CallToolRequest, in getPlansInput) (*mcp.CallToolResult, plansOutput, error) {
	client, err := r.client()
	if err != nil {
		return nil, plansOutput{}, err
	}
	activities, err := client.Activities().ListAll(ctx, in.SessionID, 100)
	if err != nil {
		return nil, plansOutput{}, wrapAPIError("get session plans", err)
	}
	plans := julessessions.ExtractPlanSummaries(activities)
	if in.LatestOnly {
		if latest := julessessions.LatestPlanSummary(plans); latest != nil {
			plans = []julessessions.PlanSummary{*latest}
		} else {
			plans = []julessessions.PlanSummary{}
		}
	}
	return nil, plansOutput{SessionID: in.SessionID, Plans: plans, TotalCount: len(plans)}, nil
}

type reviewSessionInput struct {
	SessionID        string `json:"session_id"`
	ProjectPath      string `json:"project_path,omitempty"`
	ActivityID       string `json:"activity_id,omitempty"`
	ArtifactIndex    int    `json:"artifact_index,omitempty"`
	HasArtifactIndex bool   `json:"has_artifact_index,omitempty"`
}

func (r *toolRegistry) reviewSession(ctx context.Context, _ *mcp.CallToolRequest, in reviewSessionInput) (*mcp.CallToolResult, *julessessions.SessionReview, error) {
	client, err := r.client()
	if err != nil {
		return nil, nil, err
	}
	review, err := julessessions.BuildSessionReview(ctx, client, julessessions.ReviewRequest{
		SessionID:        in.SessionID,
		WorkingDir:       in.ProjectPath,
		ActivityID:       in.ActivityID,
		ArtifactIndex:    in.ArtifactIndex,
		HasArtifactIndex: in.HasArtifactIndex,
	})
	return nil, review, wrapAPIError("review session", err)
}

type artifactsOutput struct {
	SessionID string                       `json:"session_id"`
	Artifacts []workspace.ArtifactManifest `json:"artifacts"`
}

func (r *toolRegistry) listArtifacts(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, artifactsOutput, error) {
	client, err := r.client()
	if err != nil {
		return nil, artifactsOutput{}, err
	}
	manifests, err := workspace.ListSessionArtifactManifests(ctx, client, in.SessionID)
	return nil, artifactsOutput{SessionID: in.SessionID, Artifacts: manifests}, wrapAPIError("list session artifacts", err)
}

type outputsOutput struct {
	SessionID string         `json:"session_id"`
	Outputs   []jules.Output `json:"outputs"`
}

func (r *toolRegistry) getOutputs(ctx context.Context, _ *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, outputsOutput, error) {
	client, err := r.client()
	if err != nil {
		return nil, outputsOutput{}, err
	}
	session, err := client.Sessions().Get(ctx, in.SessionID)
	if err != nil {
		return nil, outputsOutput{}, wrapAPIError("get session outputs", err)
	}
	return nil, outputsOutput{SessionID: in.SessionID, Outputs: julessessions.DocumentedOutputs(session)}, nil
}

type devBuildInput struct {
	Target  string `json:"target,omitempty" jsonschema:"all, cli, or alias"`
	Version string `json:"version,omitempty"`
	GOOS    string `json:"goos,omitempty"`
	GOARCH  string `json:"goarch,omitempty"`
	Race    bool   `json:"race,omitempty"`
}

func (r *toolRegistry) devBuild(ctx context.Context, _ *mcp.CallToolRequest, in devBuildInput) (*mcp.CallToolResult, *builder.BuildSummary, error) {
	target := in.Target
	if target == "" {
		target = "all"
	}
	if target != "all" && target != "cli" && target != "alias" {
		return nil, nil, fmt.Errorf("target must be all, cli, or alias")
	}
	version := in.Version
	if version == "" {
		version = "dev"
	}
	service := builder.NewService(builder.DefaultConfig(version, "", ""))
	summary, err := service.BuildWithResults(ctx, builder.BuildOptions{
		Target:  target,
		Version: version,
		GOOS:    in.GOOS,
		GOARCH:  in.GOARCH,
		Race:    in.Race,
	})
	return nil, summary, err
}

type devTestInput struct {
	RunPattern     *string  `json:"run_pattern,omitempty"`
	SkipPattern    *string  `json:"skip_pattern,omitempty"`
	Shuffle        *string  `json:"shuffle,omitempty"`
	Packages       []string `json:"packages,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Verbose        bool     `json:"verbose,omitempty"`
	Race           bool     `json:"race,omitempty"`
	Cover          bool     `json:"cover,omitempty"`
	Short          bool     `json:"short,omitempty"`
	FailFast       bool     `json:"fail_fast,omitempty"`
}

func (r *toolRegistry) devTest(ctx context.Context, _ *mcp.CallToolRequest, in devTestInput) (*mcp.CallToolResult, *build.TestResult, error) {
	testConfig := builder.DefaultTestConfig()
	testConfig.Verbose = in.Verbose
	testConfig.Race = in.Race
	testConfig.Cover = in.Cover
	testConfig.Short = in.Short
	testConfig.RunPattern = optionalString(in.RunPattern)
	testConfig.SkipPattern = optionalString(in.SkipPattern)
	testConfig.FailFast = in.FailFast
	testConfig.Shuffle = optionalString(in.Shuffle)
	testConfig.Packages = in.Packages
	if in.TimeoutSeconds > 0 {
		testConfig.Timeout = time.Duration(in.TimeoutSeconds) * time.Second
	}
	result := builder.NewService(builder.DefaultConfig("dev", "", "")).RunTestsWithResult(ctx, testConfig)
	return nil, result, result.Error
}

type devCheckInput struct {
	Confirm bool `json:"confirm"`
}

func (r *toolRegistry) devCheck(ctx context.Context, _ *mcp.CallToolRequest, in devCheckInput) (*mcp.CallToolResult, *builder.QualitySummary, error) {
	if err := requireConfirm(in.Confirm, "dev_check"); err != nil {
		return nil, nil, err
	}
	testConfig := builder.DefaultTestConfig()
	testConfig.Cover = true
	testConfig.CoverProfile = "coverage.out"
	summary, err := builder.NewService(builder.DefaultConfig("dev", "", "")).RunQualityChecks(ctx, builder.QualityOptions{
		Format:     true,
		Lint:       true,
		Test:       true,
		TestConfig: testConfig,
		Build:      true,
	})
	return nil, summary, err
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
