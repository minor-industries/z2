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
	host   string
	dbName string
}

func NewClient(host string, dbName string) *Client {
	return &Client{host: host, dbName: dbName}
}

func (c *Client) SendSeries(series *NamedSeries) (*SyncResponse, error) {
	return c.send(series, "series")
}

func (c *Client) SendMarkers(markers Markers) (*SyncResponse, error) {
	return c.send(markers, "markers")
}

func (c *Client) send(data msgp.Encodable, function string) (*SyncResponse, error) {
	var buf bytes.Buffer
	if err := msgp.Encode(&buf, data); err != nil {
		return nil, fmt.Errorf("error encoding: %w", err)
	}

	url := fmt.Sprintf("http://%s/sync/%s/%s", c.host, c.dbName, function)

	resp, err := http.Post(url, "application/x-msgpack", &buf)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp map[string]interface{}
		if json.Unmarshal(body, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return nil, fmt.Errorf("server error: %s", msg)
			}
		}
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

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
