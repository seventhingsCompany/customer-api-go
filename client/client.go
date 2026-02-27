// Package client provides a Go client for the seventhings API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// Response holds the raw HTTP response data.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Client is the seventhings API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	clientID   string
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client (useful for testing).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithClientID sets the OAuth client ID on the Client.
func WithClientID(id string) Option {
	return func(c *Client) {
		c.clientID = id
	}
}

// New creates a new Client. instanceURL is the base URL of the seventhings
// instance (e.g. "https://example.seventhings.com").
func New(instanceURL string, opts ...Option) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    strings.TrimRight(instanceURL, "/") + "/customer-api/v1",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetToken sets the Bearer JWT token used for authenticated requests.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Token returns the current Bearer JWT token.
func (c *Client) Token() string {
	return c.token
}

// ClientID returns the stored OAuth client ID.
func (c *Client) ClientID() string {
	return c.clientID
}

// NewWithCredentials creates a new Client and immediately authenticates using
// the provided username/password credentials.
func NewWithCredentials(ctx context.Context, instanceURL, username, password, clientID string, opts ...Option) (*Client, error) {
	c := New(instanceURL, opts...)
	if _, err := c.Login(ctx, username, password, clientID); err != nil {
		return nil, err
	}
	return c, nil
}

// NewWithToken creates a new Client with a pre-existing Bearer token.
func NewWithToken(instanceURL, token string, opts ...Option) *Client {
	c := New(instanceURL, opts...)
	c.token = token
	return c
}

// Get performs an authenticated GET request.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.doAuthenticated(ctx, http.MethodGet, path, nil)
}

// Post performs an authenticated POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*Response, error) {
	return c.doAuthenticated(ctx, http.MethodPost, path, body)
}

// Patch performs an authenticated PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body io.Reader) (*Response, error) {
	return c.doAuthenticated(ctx, http.MethodPatch, path, body)
}

// Put performs an authenticated PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*Response, error) {
	return c.doAuthenticated(ctx, http.MethodPut, path, body)
}

// Delete performs an authenticated DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.doAuthenticated(ctx, http.MethodDelete, path, nil)
}

// DoUnauthenticated performs an HTTP request without the Authorization header.
func (c *Client) DoUnauthenticated(ctx context.Context, method, path string, body io.Reader) (*Response, error) {
	req, err := c.newRequest(ctx, method, path, body, false)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) doAuthenticated(ctx context.Context, method, path string, body io.Reader) (*Response, error) {
	req, err := c.newRequest(ctx, method, path, body, true)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, authenticated bool) (*http.Request, error) {
	url := c.baseURL
	if path != "" {
		url += "/" + strings.TrimLeft(path, "/")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if authenticated && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &models.APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(bodyBytes),
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       bodyBytes,
	}, nil
}

// UUIDFromLocationHeader extracts the last path segment from the Location
// header of a response. This is used to obtain the UUID of newly created
// resources where the API returns a 201 with a Location header.
func UUIDFromLocationHeader(resp *Response) (string, error) {
	loc := resp.Header.Get("Location")
	if loc == "" {
		return "", fmt.Errorf("missing Location header")
	}
	uuid := path.Base(loc)
	if uuid == "" || uuid == "." || uuid == "/" {
		return "", fmt.Errorf("empty path in Location header: %s", loc)
	}
	return uuid, nil
}

// GetRaw performs an authenticated GET request without the Accept:
// application/json header. This is used for binary downloads.
func (c *Client) GetRaw(ctx context.Context, path string) (*Response, error) {
	return c.doRaw(ctx, http.MethodGet, path, nil, "")
}

// PostMultipart performs an authenticated POST request with the given
// Content-Type (which should include the multipart boundary). This is used for
// file uploads.
func (c *Client) PostMultipart(ctx context.Context, path string, body io.Reader, contentType string) (*Response, error) {
	return c.doRaw(ctx, http.MethodPost, path, body, contentType)
}

func (c *Client) doRaw(ctx context.Context, method, reqPath string, body io.Reader, contentType string) (*Response, error) {
	url := c.baseURL
	if reqPath != "" {
		url += "/" + strings.TrimLeft(reqPath, "/")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.do(req)
}

// IntFromLocationIDHeader extracts the integer ID from the Location-Id
// response header. This is used by CircularityHub endpoints that return a
// numeric ID instead of a UUID.
func IntFromLocationIDHeader(resp *Response) (int, error) {
	raw := resp.Header.Get("Location-Id")
	if raw == "" {
		return 0, fmt.Errorf("missing Location-Id header")
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid Location-Id header %q: %w", raw, err)
	}
	return id, nil
}

// DecodeJSON decodes the response body into dest.
func DecodeJSON(resp *Response, dest any) error {
	return json.Unmarshal(resp.Body, dest)
}
