package replay

import (
	"bytes"
	"context"
	"embed"
	"encoding/gob"
	"fmt"
	"github.com/minor-industries/z2/lib/data"
	"github.com/minor-industries/z2/lib/source"
)

//go:embed *.gob
var FS embed.FS

func FromFile(
	ctx context.Context,
	filename string,
	callback source.MessageCallback,
) error {
	fileData, err := FS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file from embed.FS: %w", err)
	}
	dataBuffer := bytes.NewBuffer(fileData)
	decoder := gob.NewDecoder(dataBuffer)
	var rawValues []data.RawValue

	if err := decoder.Decode(&rawValues); err != nil {
		return fmt.Errorf("failed to decode gob file: %w", err)
	}

	return replay(
		ctx,
		rawValues,
		callback,
	)
}
