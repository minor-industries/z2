package parser

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

type OptFloat64 struct {
	float64
	present bool
}

func (f *OptFloat64) Set(v float64) {
	f.present = true
	f.float64 = v
}

type IndoorBikeData struct {
	InstantSpeed        OptFloat64 `series:"bike_instant_speed"`
	AverageSpeed        OptFloat64
	InstantCadence      OptFloat64 `series:"bike_instant_cadence"`
	AverageCadence      OptFloat64
	TotalDistance       OptFloat64 `series:"bike_total_distance"`
	ResistanceLevel     OptFloat64 `series:"bike_resistance_level"`
	InstantPower        OptFloat64 `series:"bike_instant_power"`
	AveragePower        OptFloat64
	TotalEnergy         OptFloat64 `series:"bike_total_energy"`
	EnergyPerHour       OptFloat64 `series:"bike_energy_per_hour"`
	EnergyPerMinute     OptFloat64 `series:"bike_energy_per_minute"`
	HeartRate           OptFloat64 `series:"bike_heartrate"`
	MetabolicEquivalent OptFloat64
	ElapsedTime         OptFloat64
	RemainingTime       OptFloat64
}

func (ibd *IndoorBikeData) AllPresentFields(callback func(series string, value float64)) {
	val := reflect.ValueOf(ibd).Elem()
	//structName := val.Type().String()

	for i := 0; i < val.NumField(); i++ {
		name := val.Type().Field(i).Tag.Get("series")
		if !val.Field(i).Field(1).Bool() {
			continue // field value not present
		}

		if name == "" {
			panic(fmt.Errorf("value present but field not tagged for %s", val.Type().Field(i).Name)) // TODO: replace with continue
		}

		v := val.Field(i).Field(0).Float()
		callback(name, v)
	}
}

func bit(flags uint16, index int) bool {
	return flags&(1<<index) != 0
}

func ParseIndoorBikeData(message []byte) *IndoorBikeData {
	flags := binary.LittleEndian.Uint16(message[0:2])
	c1 := !bit(flags, 0)
	c2 := bit(flags, 1)
	c3 := bit(flags, 2)
	c4 := bit(flags, 3)
	c5 := bit(flags, 4)
	c6 := bit(flags, 5)
	c7 := bit(flags, 6)
	c8 := bit(flags, 7)
	c9 := bit(flags, 8)
	c10 := bit(flags, 9)
	c11 := bit(flags, 10)
	c12 := bit(flags, 11)
	c13 := bit(flags, 12)

	v := &IndoorBikeData{}

	i := 2
	if c1 {
		v.InstantSpeed.Set(float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.01)
		i += 2
	}
	if c2 {
		v.AverageSpeed.Set(float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.01)
		i += 2
	}
	if c3 {
		v.InstantCadence.Set(float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.5)
		i += 2
	}
	if c4 {
		v.AverageCadence.Set(float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.5)
		i += 2
	}
	if c5 {
		v.TotalDistance.Set(float64(int(binary.LittleEndian.Uint32(append(message[i:i+3], 0)))))
		i += 3
	}
	if c6 {
		v.ResistanceLevel.Set(float64(int(binary.LittleEndian.Uint16(message[i : i+2]))))
		i += 2
	}
	if c7 {
		v.InstantPower.Set(float64(int(binary.LittleEndian.Uint16(message[i : i+2]))))
		i += 2
	}
	if c8 {
		v.AveragePower.Set(float64(int(binary.LittleEndian.Uint16(message[i : i+2]))))
		i += 2
	}
	if c9 {
		v.TotalEnergy.Set(float64(int(binary.LittleEndian.Uint16(message[i : i+2]))))
		v.EnergyPerHour.Set(float64(int(binary.LittleEndian.Uint16(message[i+2 : i+4]))))
		v.EnergyPerMinute.Set(float64(int(message[i+4])))
		i += 5
	}
	if c10 {
		v.HeartRate.Set(float64(message[i]))
		i += 1
	}
	if c11 {
		v.MetabolicEquivalent.Set(float64(message[i]) * 0.1)
		i += 1
	}
	if c12 {
		v.ElapsedTime.Set(float64(binary.LittleEndian.Uint16(message[i : i+2])))
		i += 2
	}

	if c13 {
		v.RemainingTime.Set(float64(binary.LittleEndian.Uint16(message[i : i+2])))
		i += 2
	}

	return v
}
