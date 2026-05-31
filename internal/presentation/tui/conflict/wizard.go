package conflict

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// ResolutionOptions contains the user's choices for conflict resolution
type ResolutionOptions struct {
	IncludeLocalFile    bool
	IncludePatchDiff    bool
	IncludeCompilerOut  bool
	IncludeRelatedFiles bool
	Guidance            string
}

// RunWizard launches the TUI wizard to ask the user what context to gather
// and for any explicit instructions to the agent.
func RunWizard(filename string) (*ResolutionOptions, error) {
	opts := &ResolutionOptions{}
	var selectedContexts []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(fmt.Sprintf("Select context to send for resolving %s", filename)).
				Options(
					huh.NewOption("Current state of the local file", "local").Selected(true),
					huh.NewOption("The failing patch diff", "patch").Selected(true),
					huh.NewOption("Recent compiler/linter errors", "errors"),
					huh.NewOption("Additional related files", "related"),
				).
				Value(&selectedContexts),
		),
		huh.NewGroup(
			huh.NewText().
				Title("Any specific instructions for Jules?").
				Placeholder("e.g., Keep my local changes on line 42, but apply the rest").
				Value(&opts.Guidance),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	for _, ctx := range selectedContexts {
		switch ctx {
		case "local":
			opts.IncludeLocalFile = true
		case "patch":
			opts.IncludePatchDiff = true
		case "errors":
			opts.IncludeCompilerOut = true
		case "related":
			opts.IncludeRelatedFiles = true
		}
	}

	return opts, nil
}
