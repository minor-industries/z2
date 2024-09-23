//go:build !wasm

package handler

import (
	"context"
	"github.com/minor-industries/z2/gen/go/api"
)

func (a *ApiServer) DisconnectBluetoothDevices(ctx context.Context, empty *api.Empty) (*api.Empty, error) {
	a.disconnectOnce.Do(func() {
		close(a.disconnect)
	})
	return &api.Empty{}, nil
}
