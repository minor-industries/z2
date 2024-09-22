package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tinylib/msgp/msgp"
	"io"
	"net/http"
)

type Client struct {
	Endpoint string
}

func NewClient(host string) *Client {
	return &Client{Endpoint: fmt.Sprintf("%s/sync", host)}
}

func (c *Client) SendSeries(series NamedSeries) (*SyncResponse, error) {
	var buf bytes.Buffer
	if err := msgp.Encode(&buf, &series); err != nil {
		return nil, fmt.Errorf("error encoding: %w", err)
	}

	resp, err := http.Post(c.Endpoint, "application/x-msgpack", &buf)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var serverResp SyncResponse
	if err := json.Unmarshal(body, &serverResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &serverResp, nil
}
