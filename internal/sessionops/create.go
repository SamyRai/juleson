package sessionops

import (
	"errors"

	"github.com/SamyRai/juleson/pkg/jules"
)

var ErrStartingBranchRequiresSource = errors.New("starting branch requires source")

type CreateSessionRequestOptions struct {
	Prompt              string
	Source              string
	NoSource            bool
	Title               string
	StartingBranch      string
	RequirePlanApproval bool
	AutomationMode      string
}

func NormalizeSourceID(sourceID string) string {
	return jules.NormalizeSourceName(sourceID)
}

func BuildCreateSessionRequest(options CreateSessionRequestOptions) (*jules.CreateSessionRequest, error) {
	req := &jules.CreateSessionRequest{
		Prompt:              options.Prompt,
		Title:               options.Title,
		RequirePlanApproval: options.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(options.AutomationMode),
	}

	if options.NoSource {
		if options.StartingBranch != "" {
			return nil, ErrStartingBranchRequiresSource
		}
		return req, nil
	}

	req.SourceContext = &jules.SourceContext{
		Source: NormalizeSourceID(options.Source),
	}
	if options.StartingBranch != "" {
		req.SourceContext.GithubRepoContext = &jules.GithubRepoContext{
			StartingBranch: options.StartingBranch,
		}
	}

	return req, nil
}
