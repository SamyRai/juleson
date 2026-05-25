package jules

import (
	"context"
	"fmt"
	"net/url"
)

// ListSourcesOptions controls source pagination and filtering.
type ListSourcesOptions struct {
	PageSize  int
	PageToken string
	Filter    string
}

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
	return c.ListSourcesWithOptions(ctx, &ListSourcesOptions{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
	})
}

// ListSourcesWithOptions lists available code sources with pagination and filtering.
func (c *Client) ListSourcesWithOptions(ctx context.Context, options *ListSourcesOptions) (*SourcesResponse, error) {
	pageSize := 30
	pageToken := ""
	filter := ""
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		filter = options.Filter
	}
	if pageSize <= 0 {
		pageSize = 30 // default page size per API docs
	}
	if pageSize > 100 {
		pageSize = 100 // max page size per API docs
	}

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	if filter != "" {
		query.Set("filter", filter)
	}
	requestURL := fmt.Sprintf("%s/sources?%s", c.BaseURL, query.Encode())

	var response SourcesResponse
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return &response, nil
}

// ListAllSources retrieves every source by following nextPageToken.
func (c *Client) ListAllSources(ctx context.Context, pageSize int, filter string) ([]Source, error) {
	var sources []Source
	pageToken := ""
	for {
		response, err := c.ListSourcesWithPagination(ctx, pageSize, pageToken, filter)
		if err != nil {
			return nil, err
		}
		sources = append(sources, response.Sources...)
		if response.NextPageToken == "" {
			return sources, nil
		}
		pageToken = response.NextPageToken
	}
}

// GetSource retrieves a specific source by ID
func (c *Client) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("source ID is required")
	}

	resourcePath, err := sourcePath(sourceID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", c.BaseURL, resourcePath)

	var source Source
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &source); err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return &source, nil
}
