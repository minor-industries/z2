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
	host string
}

func NewClient(host string) *Client {
	return &Client{host: host}
}

func (c *Client) SendSeries(series msgp.Encodable) (*SyncResponse, error) {
	return c.send(series, "sync/series")
}

func (c *Client) SendMarkers(markers msgp.Encodable) (*SyncResponse, error) {
	return c.send(markers, "sync/markers")
}

func (c *Client) send(data msgp.Encodable, endpoint string) (*SyncResponse, error) {
	var buf bytes.Buffer
	if err := msgp.Encode(&buf, data); err != nil {
		return nil, fmt.Errorf("error encoding: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s", c.host, endpoint)

	resp, err := http.Post(url, "application/x-msgpack", &buf)
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
