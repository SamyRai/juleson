package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Download dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Downloading dependencies...")
		ctx := context.Background()

		if err := svc.DownloadDeps(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Dependencies downloaded and verified")
	},
}

var tidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Tidy dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Tidying dependencies...")
		ctx := context.Background()

		if err := svc.TidyDeps(ctx); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Dependencies tidied")
	},
}
