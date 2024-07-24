package data

type Variable struct {
	Name  string `gorm:"primary_key"`
	Value float64
}
