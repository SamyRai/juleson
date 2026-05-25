package adapters

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/pkg/jules"
)

const defaultSessionListLimit = 10

type JulesSessionGateway struct {
	client *jules.Client
}

func NewJulesSessionGateway(client *jules.Client) *JulesSessionGateway {
	return &JulesSessionGateway{client: client}
}

func (g *JulesSessionGateway) ListSources(ctx context.Context, limit int) ([]domain.Source, error) {
	if g.client == nil {
		return nil, fmt.Errorf("jules client is required")
	}
	sources, err := g.client.ListSources(ctx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Source, 0, len(sources))
	for _, source := range sources {
		result = append(result, sourceToDomain(source))
	}
	return result, nil
}

func (g *JulesSessionGateway) FindReusableSession(ctx context.Context, title string) (*domain.Session, error) {
	if g.client == nil {
		return nil, fmt.Errorf("jules client is required")
	}
	sessions, err := g.client.ListSessions(ctx, defaultSessionListLimit)
	if err != nil {
		return nil, err
	}
	for _, session := range sessions {
		if session.Title == title && session.State.IsActive() {
			converted := sessionToDomain(session)
			return &converted, nil
		}
	}
	return nil, nil
}

func (g *JulesSessionGateway) CreateSession(ctx context.Context, request domain.SessionRequest) (*domain.Session, error) {
	if g.client == nil {
		return nil, fmt.Errorf("jules client is required")
	}
	session, err := g.client.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: request.Prompt,
		Title:  request.Title,
		SourceContext: &jules.SourceContext{
			Source: request.Source.Name,
			GithubRepoContext: &jules.GithubRepoContext{
				StartingBranch: request.Branch,
			},
		},
		RequirePlanApproval: request.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(request.AutomationMode),
	})
	if err != nil {
		return nil, err
	}
	converted := sessionToDomain(*session)
	return &converted, nil
}

func (g *JulesSessionGateway) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	if g.client == nil {
		return nil, fmt.Errorf("jules client is required")
	}
	session, err := g.client.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	converted := sessionToDomain(*session)
	return &converted, nil
}

func sourceToDomain(source jules.Source) domain.Source {
	converted := domain.Source{
		ID:   source.ID,
		Name: source.Name,
	}
	if source.GithubRepo != nil {
		converted.Repository = source.GithubRepo.Owner + "/" + source.GithubRepo.Repo
	}
	return converted
}

func sessionToDomain(session jules.Session) domain.Session {
	converted := domain.Session{
		ID:    session.ID,
		Name:  session.Name,
		Title: session.Title,
		URL:   session.URL,
		State: string(session.State),
	}
	if session.SourceContext != nil {
		converted.Source = domain.Source{Name: session.SourceContext.Source}
	}
	return converted
}
