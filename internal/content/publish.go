package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type publishHTTPRequest struct {
	Collections []publishHTTPCollection `json:"collections"`
}

type publishHTTPCollection struct {
	Name    string           `json:"name"`
	Records []map[string]any `json:"records"`
}

// PublishToProduction POSTs editorial payload to production import API (specs/staging.md §5.4).
func PublishToProduction(productionURL, token string, payload []CollectionPayload) error {
	if token == "" {
		return fmt.Errorf("content publish: missing publish token (set MONMS_PUBLISH_TOKEN)")
	}

	body, err := marshalPublishBody(payload)
	if err != nil {
		return err
	}

	base := strings.TrimRight(productionURL, "/")
	url := base + "/api/monms/content/import"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("content publish: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("content publish: POST %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("content publish: POST %s returned %d: %s", url, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func marshalPublishBody(payload []CollectionPayload) ([]byte, error) {
	req := publishHTTPRequest{Collections: make([]publishHTTPCollection, len(payload))}
	for i, p := range payload {
		req.Collections[i] = publishHTTPCollection{
			Name:    p.Collection,
			Records: p.Records,
		}
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("content publish: marshal payload: %w", err)
	}
	return data, nil
}
