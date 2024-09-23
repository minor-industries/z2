//go:build !wasm

package heartrate

import (
	"context"
	"github.com/minor-industries/z2/source"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	address = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
)

func TestHeartrate(t *testing.T) {
	errCh := make(chan error)
	src := &Source{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	go func() {
		errCh <- source.Run(ctx, errCh, address, nil)
	}()

	require.NoError(t, <-errCh)
}
