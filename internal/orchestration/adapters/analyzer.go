package adapters

import (
	"context"

	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type AnalyzerAdapter struct {
	analyzer *analyzer.ProjectAnalyzer
}

func NewAnalyzerAdapter(projectAnalyzer *analyzer.ProjectAnalyzer) *AnalyzerAdapter {
	if projectAnalyzer == nil {
		projectAnalyzer = analyzer.NewProjectAnalyzer()
	}
	return &AnalyzerAdapter{analyzer: projectAnalyzer}
}

func (a *AnalyzerAdapter) AnalyzeProject(ctx context.Context, projectPath string) (*domain.ProjectContext, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	context, err := a.analyzer.Analyze(projectPath)
	if err != nil {
		return nil, err
	}
	return projectContextToDomain(context), nil
}

func projectContextToDomain(context *analyzer.ProjectContext) *domain.ProjectContext {
	if context == nil {
		return nil
	}
	project := &domain.ProjectContext{
		ProjectPath:  context.ProjectPath,
		ProjectName:  context.ProjectName,
		ProjectType:  context.ProjectType,
		Languages:    append([]string(nil), context.Languages...),
		Frameworks:   append([]string(nil), context.Frameworks...),
		Architecture: context.Architecture,
		Complexity:   context.Complexity,
		GitStatus:    context.GitStatus,
		Dependencies: copyStringMap(context.Dependencies),
		Values:       copyStringMap(context.CustomParams),
	}
	if context.CodeQuality != nil {
		project.Quality = &domain.QualityMetrics{
			TestCoverage:    context.CodeQuality.TestCoverage,
			CodeComplexity:  context.CodeQuality.CodeComplexity,
			Maintainability: context.CodeQuality.Maintainability,
			SecurityIssues:  len(context.CodeQuality.SecurityIssues),
			CodeSmells:      len(context.CodeQuality.CodeSmells),
		}
	}
	return project
}

func copyStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
