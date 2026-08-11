package builder

import (
	"context"
	"strings"
	"testing"
)

func TestRunModuleMaintenanceRejectsUnknownOperation(t *testing.T) {
	service := NewService(DefaultConfig("dev", "", ""))

	err := service.RunModuleMaintenance(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected unknown operation error")
	}
	if !strings.Contains(err.Error(), "unknown module operation: unknown") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectedBinariesUsesConfiguredTargets(t *testing.T) {
	service := NewService(&Config{
		BinaryCLI:   "cli-bin",
		BinaryAlias: "short-bin",
		CmdCLIDir:   "cmd/cli",
	})

	tests := map[string][]binaryTarget{
		"cli": {
			{name: "cli-bin", path: "./cmd/cli"},
		},
		"alias": {
			{name: "short-bin", path: "./cmd/cli"},
		},
		"all": {
			{name: "cli-bin", path: "./cmd/cli"},
			{name: "short-bin", path: "./cmd/cli"},
		},
		"": {
			{name: "cli-bin", path: "./cmd/cli"},
			{name: "short-bin", path: "./cmd/cli"},
		},
	}

	for target, want := range tests {
		t.Run(target, func(t *testing.T) {
			got := service.selectedBinaries(target)
			if len(got) != len(want) {
				t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("target %q item %d = %#v, want %#v", target, i, got[i], want[i])
				}
			}
		})
	}
}
