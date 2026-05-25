package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type SourceMatcherAdapter struct{}

func NewSourceMatcherAdapter() *SourceMatcherAdapter {
	return &SourceMatcherAdapter{}
}

func (SourceMatcherAdapter) MatchSource(ctx context.Context, project domain.ProjectContext, sources []domain.Source) (*domain.Source, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources available")
	}
	repository := strings.TrimSpace(project.Values["repository"])
	if repository == "" {
		repository = strings.TrimSpace(project.ProjectName)
	}
	for _, source := range sources {
		if repository != "" && (strings.Contains(source.Repository, repository) || strings.Contains(source.Name, repository)) {
			matched := source
			return &matched, nil
		}
	}
	matched := sources[0]
	return &matched, nil
}
