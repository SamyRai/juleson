package e2e

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation/cli"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"juleson": main1,
	}))
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func main1() int {
	cfg, _ := config.LoadOptional()
	app := cli.NewApp(cfg)

	if err := app.Execute(); err != nil {
		return 1
	}
	return 0
}
