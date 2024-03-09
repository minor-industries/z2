package replay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"github.com/minor-industries/codelab/cmd/bike/testdata"
	"github.com/pkg/errors"
	"io/fs"
	"strconv"
	"strings"
	"time"
)

type line struct {
	timestamp time.Time
	value     []byte
}

func Run(
	ctx context.Context,
	filename string,
	callback func(time.Time, []byte) error,
) error {
	content, err := fs.ReadFile(testdata.FS, filename)
	if err != nil {
		return errors.Wrap(err, "readfile")
	}

	var lines []line

	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		t0, err := strconv.Atoi(parts[0])
		if err != nil {
			return errors.Wrap(err, "atoi")
		}

		t := time.UnixMilli(int64(t0))

		raw, err := hex.DecodeString(parts[1])
		if err != nil {
			return errors.Wrap(err, "decode string")
		}

		lines = append(lines, line{t, raw})
	}

	t0 := time.Now()
	ticker := time.NewTicker(10 * time.Millisecond)

	tFirst := lines[0].timestamp
	for now := range ticker.C {
		dt := now.Sub(t0)
		t := tFirst.Add(dt)
		line := lines[0]
		if t.Before(line.timestamp) {
			continue
		}
		callback(now, line.value)
		lines = lines[1:]
		if len(lines) == 0 {
			ticker.Stop()
			break
		}
	}

	<-ctx.Done()
	return nil
}
