package workspace

import (
	"strconv"
	"strings"
)

// FileChange represents changes to a single file.
type FileChange struct {
	Path         string `json:"path"`
	LinesAdded   int    `json:"linesAdded"`
	LinesRemoved int    `json:"linesRemoved"`
}

// SessionChanges represents a summary of changes in a session.
type SessionChanges struct {
	SessionID               string       `json:"sessionId"`
	Files                   []FileChange `json:"files"`
	SuggestedCommitMessages []string     `json:"suggestedCommitMessages,omitempty"`
	Warnings                []string     `json:"warnings,omitempty"`
	BaseCommitMismatches    []string     `json:"baseCommitMismatches,omitempty"`
	TotalPatches            int          `json:"totalPatches"`
}

// parsePatchFiles extracts file changes from a git patch.
func parsePatchFiles(patch string) []FileChange {
	var changes []FileChange
	lines := strings.Split(patch, "\n")

	var currentFile *FileChange

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				changes = append(changes, *currentFile)
			}
			currentFile = &FileChange{Path: extractDiffGitPath(line)}
		} else if currentFile != nil {
			if renamedPath, ok := strings.CutPrefix(line, "rename to "); ok {
				currentFile.Path = strings.TrimSpace(renamedPath)
				continue
			}
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentFile.LinesAdded++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentFile.LinesRemoved++
			}
		}
	}

	if currentFile != nil {
		changes = append(changes, *currentFile)
	}

	return changes
}

func extractDiffGitPath(line string) string {
	remainder := strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	first, rest := nextPatchToken(remainder)
	second, _ := nextPatchToken(rest)
	if second != "" && second != "/dev/null" {
		return stripPatchPrefix(second)
	}
	return stripPatchPrefix(first)
}

func nextPatchToken(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	if value[0] == '"' {
		for i := 1; i < len(value); i++ {
			if value[i] == '"' && value[i-1] != '\\' {
				token := value[:i+1]
				unquoted, err := strconv.Unquote(token)
				if err != nil {
					unquoted = strings.Trim(token, `"`)
				}
				return unquoted, value[i+1:]
			}
		}
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func stripPatchPrefix(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	return path
}

func parseGitApplyOutput(output string) []string {
	var files []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Checking patch ") {
			file := strings.TrimPrefix(line, "Checking patch ")
			file = strings.TrimSuffix(file, "...")
			files = append(files, file)
		} else if strings.HasPrefix(line, "Applying patch to ") {
			file := strings.TrimPrefix(line, "Applying patch to ")
			file = strings.TrimSuffix(file, "...")
			files = append(files, file)
		}
	}

	seen := make(map[string]bool)
	var unique []string
	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
		}
	}

	return unique
}

func appendUniqueStrings(values []string, candidates ...string) []string {
	seen := make(map[string]bool, len(values)+len(candidates))
	for _, value := range values {
		seen[value] = true
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		values = append(values, candidate)
		seen[candidate] = true
	}
	return values
}
