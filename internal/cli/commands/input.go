package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func mustMarkFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

func scanPromptValue(target *string) error {
	_, err := fmt.Scanln(target)
	if err != nil {
		if err.Error() == "unexpected newline" {
			*target = ""
			return nil
		}
		return err
	}
	*target = strings.TrimSpace(*target)
	return nil
}
