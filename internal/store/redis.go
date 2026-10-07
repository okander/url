package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"shortener/internal/base62"
)

const (
	counterKey = "shortener:next_id"
	linkPrefix = "shortener:link:"
	maxCodeLen = 32
)

// Redis is a Store backed by Upstash Redis through its REST API. It uses only
// the standard library: each command is a JSON array POSTed to the REST URL.
type Redis struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewRedis returns a Redis store for the given REST URL and token.
func NewRedis(restURL, token string) *Redis {
	return &Redis{
		baseURL: strings.TrimRight(restURL, "/"),
		token:   token,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Create atomically reserves the next ID (INCR), derives the code, and saves it.
func (r *Redis) Create(ctx context.Context, url string) (Link, error) {
	raw, err := r.command(ctx, "INCR", counterKey)
	if err != nil {
		return Link{}, err
	}
	var n uint64
	if err := json.Unmarshal(raw, &n); err != nil {
		return Link{}, fmt.Errorf("redis: unexpected INCR result %s: %w", raw, err)
	}

	id := startID + n
	code := base62.Encode(id)

	if _, err := r.command(ctx, "SET", linkPrefix+code, url); err != nil {
		return Link{}, err
	}
	return Link{ID: id, Code: code, URL: url, CreatedAt: time.Now().UTC()}, nil
}

// Get looks a link up by code.
func (r *Redis) Get(ctx context.Context, code string) (Link, error) {
	if code == "" || len(code) > maxCodeLen {
		return Link{}, ErrNotFound
	}
	raw, err := r.command(ctx, "GET", linkPrefix+code)
	if err != nil {
		return Link{}, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return Link{}, ErrNotFound
	}
	var target string
	if err := json.Unmarshal(raw, &target); err != nil {
		return Link{}, fmt.Errorf("redis: unexpected GET result %s: %w", raw, err)
	}
	return Link{Code: code, URL: target}, nil
}

// command sends one Redis command and returns the raw "result" JSON value.
func (r *Redis) command(ctx context.Context, args ...string) (json.RawMessage, error) {
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("redis: request failed: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("redis: status %d, bad response: %w", resp.StatusCode, err)
	}
	if out.Error != "" {
		return nil, fmt.Errorf("redis: %s", out.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("redis: unexpected status %d", resp.StatusCode)
	}
	return out.Result, nil
}
