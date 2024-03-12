package database

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Get(filename string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}

	for _, table := range []any{
		&Value{},
		&Series{},
		&RawValue{},
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

	"rower_stroke_count",
	"rower_power",
	"rower_speed",
	"rower_spm",
}

func RandomID() []byte {
	var result [16]byte
	_, err := rand.Read(result[:])
	if err != nil {
		panic(err)
	}
	return result[:]
}

func HashedID(s string) []byte {
	var result [16]byte
	h := sha256.New()
	h.Write([]byte(s))
	sum := h.Sum(nil)
	copy(result[:], sum[:16])
	return result[:]
}

func LoadAllSeries(db *gorm.DB) (map[string]*Series, error) {
	seriesMap, err := loadSeries(db)
	if err != nil {
		return nil, errors.Wrap(err, "initial load")
	}

	for _, name := range seriesNames {
		if _, found := seriesMap[name]; found {
			continue
		}
		db.Create(&Series{
			ID:   HashedID(name),
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

func LoadData(db *gorm.DB, series []byte) ([]Value, error) {
	var result []Value

	tx := db.Where("series_id = ?", series).Order("timestamp asc").Find(&result)
	if tx.Error != nil {
		return nil, errors.Wrap(tx.Error, "find")
	}

	return result, nil
}
