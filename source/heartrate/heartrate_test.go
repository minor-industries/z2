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

const (
	address = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
)

type Source struct{}

func (s *Source) Convert(msg source.Message) []source.Value {
	//TODO implement me
	panic("implement me")
}

func (s *Source) Services() []bluetooth.UUID {
	return []bluetooth.UUID{
		bluetooth.ServiceUUIDHeartRate,
	}
}

func (s *Source) Characteristics() []bluetooth.UUID {
	return []bluetooth.UUID{
		bluetooth.CharacteristicUUIDHeartRateMeasurement,
	}
}

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
