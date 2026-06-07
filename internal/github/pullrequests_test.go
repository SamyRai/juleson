package github

import (
	"testing"
)

func TestParsePRURL(t *testing.T) {
	service := &PullRequestService{}

	tests := []struct {
		name       string
		url        string
		wantOwner  string
		wantRepo   string
		wantNumber int
		wantErr    bool
	}{
		{
			name:       "valid url",
			url:        "https://github.com/SamyRai/juleson/pull/123",
			wantOwner:  "SamyRai",
			wantRepo:   "juleson",
			wantNumber: 123,
			wantErr:    false,
		},
		{
			name:       "valid url with trailing slash",
			url:        "https://github.com/SamyRai/juleson/pull/123/",
			wantOwner:  "SamyRai",
			wantRepo:   "juleson",
			wantNumber: 123,
			wantErr:    false,
		},
		{
			name:    "invalid format missing pull",
			url:     "https://github.com/SamyRai/juleson/issues/123",
			wantErr: true,
		},
		{
			name:    "invalid format too short",
			url:     "https://github.com/SamyRai/juleson",
			wantErr: true,
		},
		{
			name:    "invalid PR number",
			url:     "https://github.com/SamyRai/juleson/pull/abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, prNumber, err := service.parsePRURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parsePRURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("parsePRURL() repo = %v, want %v", repo, tt.wantRepo)
				}
				if prNumber != tt.wantNumber {
					t.Errorf("parsePRURL() prNumber = %v, want %v", prNumber, tt.wantNumber)
				}
			}
		})
	}
}
