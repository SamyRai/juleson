package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		FormatJSON: true,
		Output:     &buf,
	})

	l.Info("testing token", "github_token", "ghp_123456789012345678901234567890123456")

	output := buf.String()
	if strings.Contains(output, "ghp_123456789012345678901234567890123456") {
		t.Errorf("Secret was not redacted: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("Expected [REDACTED] in output, got: %s", output)
	}

	buf.Reset()
	l.Info("testing jules key", "jules_api_key", "jules_ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	output = buf.String()
	if strings.Contains(output, "jules_ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		t.Errorf("Jules secret was not redacted: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Errorf("Expected [REDACTED] in output, got: %s", output)
	}

	buf.Reset()
	// Test pattern inside a normal string
	l.Info("request failed", "error", "server returned 401 for token ghp_123456789012345678901234567890123456")
	output = buf.String()
	if strings.Contains(output, "ghp_123456789012345678901234567890123456") {
		t.Errorf("Secret in error string was not redacted: %s", output)
	}
}
