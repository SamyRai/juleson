package automation

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/templates"
)

func TestExecuteTasksDetectsCircularDependencyBeforeSideEffects(t *testing.T) {
	engine := NewEngine(nil, nil)

	_, err := engine.executeTasks(t.Context(), []templates.TemplateTask{
		{Name: "a", DependsOn: []string{"b"}},
		{Name: "b", DependsOn: []string{"a"}},
	})
	if err == nil || !strings.Contains(err.Error(), "circular dependency") {
		t.Fatalf("expected circular dependency error, got %v", err)
	}
}

func TestProcessPromptReplacesContextAndBuiltins(t *testing.T) {
	engine := NewEngine(nil, nil)
	engine.projectPath = "/tmp/project"
	engine.context = &analyzer.ProjectContext{
		ProjectPath:  "/tmp/project",
		ProjectName:  "project",
		ProjectType:  "go",
		Languages:    []string{"Go", "Shell"},
		Frameworks:   []string{"cobra"},
		Architecture: "cli",
		Complexity:   "medium",
		GitStatus:    "clean",
		CustomParams: map[string]string{"Owner": "platform"},
	}

	got, err := engine.processPrompt("{{.ProjectName}} {{.Languages}} {{.Frameworks}} {{.Owner}} {{.Timestamp}}", map[string]string{
		"Languages":  "",
		"Frameworks": "",
		"Owner":      "",
	})
	if err != nil {
		t.Fatalf("processPrompt returned error: %v", err)
	}

	for _, want := range []string{"project", "Go, Shell", "cobra", "platform"} {
		if !strings.Contains(got, want) {
			t.Fatalf("processed prompt %q missing %q", got, want)
		}
	}
	fields := strings.Fields(got)
	if _, err := time.Parse(time.RFC3339, fields[len(fields)-1]); err != nil {
		t.Fatalf("timestamp was not RFC3339 in %q: %v", got, err)
	}
}

func TestGenerateOutputFilesWritesReport(t *testing.T) {
	dir := t.TempDir()
	engine := NewEngine(nil, nil)
	engine.projectPath = dir
	engine.context = &analyzer.ProjectContext{
		ProjectPath:  dir,
		ProjectName:  "demo",
		ProjectType:  "go",
		CustomParams: map[string]string{},
	}
	result := &ExecutionResult{
		TemplateName: "quality",
		ProjectPath:  dir,
		Duration:     time.Second,
		Success:      true,
		TasksExecuted: []TaskExecutionResult{
			{TaskName: "lint", TaskType: "quality", Success: true},
		},
	}
	template := &templates.Template{
		Output: templates.TemplateOutput{
			Files: []templates.TemplateOutputFile{
				{Path: filepath.Join(dir, "{{.ProjectName}}-report.md"), Template: "summary"},
			},
		},
	}

	if err := engine.generateOutputFiles(template, result); err != nil {
		t.Fatalf("generateOutputFiles returned error: %v", err)
	}
	if len(result.OutputFiles) != 1 {
		t.Fatalf("expected one output file, got %#v", result.OutputFiles)
	}
	content, err := os.ReadFile(result.OutputFiles[0])
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(content), "quality") || !strings.Contains(string(content), "lint") {
		t.Fatalf("unexpected report content:\n%s", content)
	}
}

func TestMatchGitRepoToSourceMatchesOrigin(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "git@github.com:SamyRai/juleson.git")

	engine := NewEngine(nil, nil)
	engine.projectPath = dir

	source, err := engine.matchGitRepoToSource([]jules.Source{
		{Name: "sources/github/SamyRai/juleson"},
	})
	if err != nil {
		t.Fatalf("matchGitRepoToSource returned error: %v", err)
	}
	if source.Name != "sources/github/SamyRai/juleson" {
		t.Fatalf("unexpected source: %#v", source)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
