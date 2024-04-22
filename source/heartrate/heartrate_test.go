package heartrate

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/minor-industries/z2/source"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"tinygo.org/x/bluetooth"
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
		errCh <- source.Run(
			ctx,
			errCh,
			address,
			src,
			func(
				t time.Time,
				service bluetooth.UUID,
				characteristic bluetooth.UUID,
				msg []byte,
			) error {
				fmt.Println(hex.EncodeToString(msg))
				fmt.Println(msg[1])
				return nil
			},
		)
	}()

	require.NoError(t, <-errCh)
}
