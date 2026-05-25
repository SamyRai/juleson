package orchestrator

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDockerOperationsBuildConstructsCommandAndParsesImageID(t *testing.T) {
	runner := &recordingRunner{output: "Step 1/1\nSuccessfully built abcdef123456\n"}
	ops := &DockerOperations{runner: runner}

	result, err := ops.Build(context.Background(), DockerBuildOptions{
		Path:       ".",
		Tag:        "juleson:test",
		Dockerfile: "Dockerfile.dev",
		BuildArgs:  map[string]string{"VERSION": "dev"},
		NoCache:    true,
	})

	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if runner.name != "docker" {
		t.Fatalf("command name = %q, want docker", runner.name)
	}
	wantArgs := []string{"build", "-t", "juleson:test", "-f", "Dockerfile.dev", "--no-cache", "--build-arg", "VERSION=dev", "."}
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.args, wantArgs)
	}
	if result.ImageID != "abcdef123456" {
		t.Fatalf("ImageID = %q, want abcdef123456", result.ImageID)
	}
}

func TestDockerOperationsRunContainerRequiresImage(t *testing.T) {
	ops := &DockerOperations{runner: &recordingRunner{}}

	result, err := ops.RunContainer(context.Background(), DockerRunOptions{})

	if err == nil {
		t.Fatal("RunContainer returned nil error, want missing image error")
	}
	if result.Success {
		t.Fatal("Success = true, want false")
	}
}

func TestDockerOperationsImagesParsesQuietAndTableOutput(t *testing.T) {
	tests := []struct {
		name   string
		quiet  bool
		output string
		want   []string
	}{
		{
			name:   "table skips header",
			output: "REPOSITORY TAG IMAGE ID\nrepo latest abc123\n",
			want:   []string{"repo latest abc123"},
		},
		{
			name:   "quiet keeps ids",
			quiet:  true,
			output: "abc123\ndef456\n",
			want:   []string{"abc123", "def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &DockerOperations{runner: &recordingRunner{output: tt.output}}
			result, err := ops.Images(context.Background(), DockerListOptions{Quiet: tt.quiet})
			if err != nil {
				t.Fatalf("Images returned error: %v", err)
			}
			if !reflect.DeepEqual(result.Items, tt.want) {
				t.Fatalf("Items = %#v, want %#v", result.Items, tt.want)
			}
		})
	}
}

func TestDockerOperationsExecPropagatesOutputOnFailure(t *testing.T) {
	ops := &DockerOperations{runner: &recordingRunner{output: "boom", err: errors.New("exit 1")}}

	output, err := ops.Exec(context.Background(), DockerExecOptions{
		Container: "container",
		Command:   []string{"echo", "hi"},
	})

	if err == nil {
		t.Fatal("Exec returned nil error, want failure")
	}
	if output != "boom" {
		t.Fatalf("output = %q, want boom", output)
	}
}

type recordingRunner struct {
	name   string
	args   []string
	output string
	err    error
}

func (r *recordingRunner) Run(ctx context.Context, name string, args ...string) error {
	_, err := r.CombinedOutput(ctx, name, args...)
	return err
}

func (r *recordingRunner) CombinedOutput(_ context.Context, name string, args ...string) (string, error) {
	r.name = name
	r.args = append([]string(nil), args...)
	return r.output, r.err
}
