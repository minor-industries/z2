//go:build wasm

package sync

import (
	"github.com/minor-industries/rtgraph/storage"
)

func sendMarkers(src storage.StorageBackend, client *Client, info func(string)) error {
	info("markers sync unsupported")
	return nil
}
