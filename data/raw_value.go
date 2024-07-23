package data

import "time"

type RawValue struct {
	ID []byte `gorm:"primary_key"`

	ServiceID        string
	CharacteristicID string

	Timestamp time.Time `gorm:"index"`
	Message   []byte
}
