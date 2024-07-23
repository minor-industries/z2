package data

type Marker struct {
	ID string `gorm:"primary_key"`

	Type      string `gorm:"index;not null"`
	Ref       string `gorm:"index;not null"`
	Timestamp int64  `gorm:"index;not null"`
}
