package db

type Variable struct {
	Name  string `gorm:"primary_key"`
	Value float64
}
