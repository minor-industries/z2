package app

import (
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/rtgraph/storage"
)

type Backends struct {
	Samples   storage.StorageBackend
	RawValues *sqlite.Backend
}
