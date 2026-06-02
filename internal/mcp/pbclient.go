package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PBClient calls PocketBase REST as an authenticated user.
type PBClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func (c *PBClient) do(method, path string, body []byte) ([]byte, int, error) {
	if c.Client == nil {
		c.Client = http.DefaultClient
	}
	url := strings.TrimRight(c.BaseURL, "/") + path
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 400 {
		return data, resp.StatusCode, fmt.Errorf("pocketbase %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return data, resp.StatusCode, nil
}

func (c *PBClient) ListRecords(collection string, perPage int) ([]byte, error) {
	path := fmt.Sprintf("/api/collections/%s/records?perPage=%d", collection, perPage)
	data, _, err := c.do(http.MethodGet, path, nil)
	return data, err
}

func (c *PBClient) GetRecord(collection, id string) ([]byte, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, id)
	data, _, err := c.do(http.MethodGet, path, nil)
	return data, err
}

func (c *PBClient) PatchRecord(collection, id string, fields map[string]any) ([]byte, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, id)
	body, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	data, _, err := c.do(http.MethodPatch, path, body)
	return data, err
}
