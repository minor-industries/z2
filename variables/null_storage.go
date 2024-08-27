package variables

import (
	"github.com/minor-industries/z2/data"
)

type NullStorage struct {
}

func (n NullStorage) Load() ([]data.Variable, error) {
	return nil, nil
}

func (n NullStorage) Save(value *data.Variable) error {
	return nil
}
