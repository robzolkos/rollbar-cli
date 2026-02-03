package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL        = "https://api.rollbar.com/api/1"
	DefaultTimeout = 30 * time.Second
)

// Client is the Rollbar API client
type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

// NewClient creates a new Rollbar API client
func NewClient(accessToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		accessToken: accessToken,
		baseURL:     BaseURL,
	}
}

// APIError represents an error from the Rollbar API
type APIError struct {
	StatusCode int
	Message    string
	Err        int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("rollbar API error: %s (status: %d, err: %d)", e.Message, e.StatusCode, e.Err)
}

// IsAuthError returns true if this is an authentication error
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == 401 || e.StatusCode == 403
}

// IsNotFound returns true if the resource was not found
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsRateLimited returns true if rate limited
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

func (c *Client) doRequest(method, path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Message,
				Err:        errResp.Err,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return body, nil
}

func (c *Client) doRequestWithBody(method, path string, query url.Values, payload interface{}) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if payload != nil {
		jsonBody, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Rollbar-Access-Token", c.accessToken)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Message,
				Err:        errResp.Err,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return body, nil
}

// UpdateItemStatus updates the status of an item (e.g., "resolved", "active", "muted")
func (c *Client) UpdateItemStatus(id int64, status string) (*Item, error) {
	payload := map[string]string{
		"status": status,
	}

	body, err := c.doRequestWithBody("PATCH", fmt.Sprintf("/item/%d", id), nil, payload)
	if err != nil {
		return nil, err
	}

	var resp ItemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	resp.Result.ComputeFields()
	return &resp.Result, nil
}

// ItemsOptions configures the list items request
type ItemsOptions struct {
	Status      string // active, resolved, muted, any
	Level       string // debug, info, warning, error, critical (comma-separated)
	Environment string
	Query       string    // Text search
	DateFrom    time.Time // Filter by last_occurrence_timestamp >= date
	DateTo      time.Time // Filter by last_occurrence_timestamp <= date
	Page        int
	Limit       int // Items per page (max 100)
}

// ListItems returns items matching the given options
func (c *Client) ListItems(opts ItemsOptions) ([]Item, int, error) {
	// Handle comma-separated levels by making multiple requests
	if opts.Level != "" && strings.Contains(opts.Level, ",") {
		return c.listItemsMultiLevel(opts)
	}

	return c.listItemsSingleLevel(opts)
}

// listItemsMultiLevel handles comma-separated level filters by making multiple API calls
func (c *Client) listItemsMultiLevel(opts ItemsOptions) ([]Item, int, error) {
	levels := strings.Split(opts.Level, ",")
	seen := make(map[int64]bool)
	var allItems []Item

	for _, level := range levels {
		levelOpts := opts
		levelOpts.Level = strings.TrimSpace(level)

		items, _, err := c.listItemsSingleLevel(levelOpts)
		if err != nil {
			return nil, 0, err
		}

		// Deduplicate by item ID
		for _, item := range items {
			if !seen[item.ID.Int64()] {
				seen[item.ID.Int64()] = true
				allItems = append(allItems, item)
			}
		}
	}

	return allItems, 1, nil
}

// listItemsSingleLevel makes a single API call for items
func (c *Client) listItemsSingleLevel(opts ItemsOptions) ([]Item, int, error) {
	q := url.Values{}

	if opts.Status != "" && opts.Status != "any" {
		q.Set("status", opts.Status)
	}
	if opts.Level != "" {
		q.Set("level", opts.Level)
	}
	if opts.Environment != "" {
		q.Set("environment", opts.Environment)
	}
	if opts.Query != "" {
		q.Set("query", opts.Query)
	}
	if !opts.DateFrom.IsZero() {
		q.Set("date_from", opts.DateFrom.Format("2006-01-02T15:04:05"))
	}
	if !opts.DateTo.IsZero() {
		q.Set("date_to", opts.DateTo.Format("2006-01-02T15:04:05"))
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}

	body, err := c.doRequest("GET", "/items", q)
	if err != nil {
		return nil, 0, err
	}

	var resp ItemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, fmt.Errorf("parsing response: %w", err)
	}

	// Compute derived fields
	for i := range resp.Result.Items {
		resp.Result.Items[i].ComputeFields()
	}

	return resp.Result.Items, resp.Result.Page, nil
}

// GetItem returns an item by its internal ID
func (c *Client) GetItem(id int64) (*Item, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/item/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp ItemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	resp.Result.ComputeFields()
	return &resp.Result, nil
}

// GetItemByCounter returns an item by its project-local counter (e.g., #123)
func (c *Client) GetItemByCounter(counter int) (*Item, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/item_by_counter/%d", counter), nil)
	if err != nil {
		return nil, err
	}

	var resp ItemResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	resp.Result.ComputeFields()
	return &resp.Result, nil
}

// InstancesOptions configures the list instances request
type InstancesOptions struct {
	ItemID int64 // If set, list instances for this item only
	Page   int
}

// ListInstances returns instances (occurrences) for an item or all items
func (c *Client) ListInstances(opts InstancesOptions) ([]Instance, error) {
	q := url.Values{}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}

	var path string
	if opts.ItemID > 0 {
		path = fmt.Sprintf("/item/%d/instances", opts.ItemID)
	} else {
		path = "/instances"
	}

	body, err := c.doRequest("GET", path, q)
	if err != nil {
		return nil, err
	}

	var resp InstancesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	for i := range resp.Result.Instances {
		resp.Result.Instances[i].ComputeFields()
	}

	return resp.Result.Instances, nil
}

// GetInstance returns a single occurrence by ID
func (c *Client) GetInstance(id int64) (*Instance, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/instance/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp InstanceResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	resp.Result.ComputeFields()
	return &resp.Result, nil
}

// ProjectInfo represents basic project info for whoami
type ProjectInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProjectInfoResponse is the response from /project
type ProjectInfoResponse struct {
	Err    int         `json:"err"`
	Result ProjectInfo `json:"result"`
}

// GetProjectInfo returns info about the current project (based on access token)
func (c *Client) GetProjectInfo() (*ProjectInfo, error) {
	body, err := c.doRequest("GET", "/project", nil)
	if err != nil {
		return nil, err
	}

	var resp ProjectInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &resp.Result, nil
}
