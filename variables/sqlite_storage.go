package variables

import (
	"github.com/minor-industries/z2/data"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SqliteStorage struct {
	db *gorm.DB
}

func NewSQLiteStorage(db *gorm.DB) *SqliteStorage {
	return &SqliteStorage{db: db}
}

func (s *SqliteStorage) Load() ([]data.Variable, error) {
	var rows []data.Variable
	tx := s.db.Find(&rows)
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "find rows")
	}
	return rows, nil
}

func (s *SqliteStorage) Save(v *data.Variable) error {
	return s.db.Save(v).Error
}
