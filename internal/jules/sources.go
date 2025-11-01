package jules

import (
	"context"
	"fmt"
)

// ListSources lists available code sources with pagination support
// Deprecated: Use ListSourcesWithPagination for full pagination support
func (c *Client) ListSources(ctx context.Context, pageSize int) ([]Source, error) {
	response, err := c.ListSourcesWithPagination(ctx, pageSize, "", "")
	if err != nil {
		return nil, err
	}
	return response.Sources, nil
}

// ListSourcesWithPagination lists available code sources with full pagination and filtering support
// filter: Optional AIP-160 filter expression (e.g., "name=sources/source1 OR name=sources/source2")
func (c *Client) ListSourcesWithPagination(ctx context.Context, pageSize int, pageToken, filter string) (*SourcesResponse, error) {
	if pageSize <= 0 {
		pageSize = 30 // default page size per API docs
	}
	if pageSize > 100 {
		pageSize = 100 // max page size per API docs
	}

	url := fmt.Sprintf("%s/sources?pageSize=%d", c.BaseURL, pageSize)
	if pageToken != "" {
		url += fmt.Sprintf("&pageToken=%s", pageToken)
	}
	if filter != "" {
		url += fmt.Sprintf("&filter=%s", filter)
	}

	var response SourcesResponse
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return &response, nil
}

// GetSource retrieves a specific source by ID
func (c *Client) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("source ID is required")
	}

	url := fmt.Sprintf("%s/sources/%s", c.BaseURL, sourceID)

	var source Source
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &source); err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return &source, nil
}
