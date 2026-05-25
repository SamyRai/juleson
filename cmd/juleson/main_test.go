package main

import "testing"

func TestIsConfigValidateCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "config validate",
			args: []string{"config", "validate"},
			want: true,
		},
		{
			name: "config validate with flags",
			args: []string{"config", "validate", "--help"},
			want: true,
		},
		{
			name: "config parent",
			args: []string{"config"},
			want: false,
		},
		{
			name: "other command",
			args: []string{"sessions", "list"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConfigValidateCommand(tt.args); got != tt.want {
				t.Fatalf("isConfigValidateCommand(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
