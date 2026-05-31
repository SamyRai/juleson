package workspace

import (
	"fmt"
	"io"
	"os/exec"
)

type OfficialCLIStreams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func RunOfficialJulesCLI(streams OfficialCLIStreams, args ...string) error {
	julesPath, err := exec.LookPath("jules")
	if err != nil {
		return fmt.Errorf("official Jules CLI is not installed or not on PATH; install it to use 'juleson official'")
	}
	cmd := exec.Command(julesPath, args...)
	cmd.Stdin = streams.Stdin
	cmd.Stdout = streams.Stdout
	cmd.Stderr = streams.Stderr
	return cmd.Run()
}
