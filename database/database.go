package database

import (
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Value struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	Timestamp time.Time
	Value     float64
	SeriesID  uint16
	Series    *Series `gorm:"foreignKey:SeriesID"`
}

type Series struct {
	ID   uint16 `gorm:"primaryKey"`
	Name string `gorm:"unique"`
	Unit string
}

func Get(filename string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	for _, table := range []any{
		&Value{},
		&Series{},
	} {
		err = db.AutoMigrate(table)
		if err != nil {
			return nil, errors.Wrap(err, "migrate measurement")
		}
	}
	return db, nil
}

var seriesNames = []string{
	"bike_instant_speed",
	"bike_instant_cadence",
	"bike_total_distance",
	"bike_resistance_level",
	"bike_instant_power",
	"bike_total_energy",
	"bike_energy_per_hour",
	"bike_energy_per_minute",
	"bike_heartrate",
}

func LoadSeries(db *gorm.DB) (map[string]*Series, error) {
	seriesMap, err := loadSeries(db)
	if err != nil {
		return nil, errors.Wrap(err, "initial load")
	}

	for _, name := range seriesNames {
		if _, found := seriesMap[name]; found {
			continue
		}
		db.Create(&Series{
			Name: name,
			Unit: "",
		})
	}

	return loadSeries(db)
}

func loadSeries(db *gorm.DB) (map[string]*Series, error) {
	typeMap := map[string]*Series{}
	{
		var types []*Series
		tx := db.Find(&types)
		if tx.Error != nil {
			return nil, errors.Wrap(tx.Error, "find")
		}

		for _, mt := range types {
			typeMap[mt.Name] = mt
		}
	}

	return typeMap, nil
}
