package parser

import (
	"encoding/binary"
)

type IndoorBikeData struct {
	InstantSpeed        float64
	AverageSpeed        float64
	InstantCadence      float64
	AverageCadence      float64
	TotalDistance       float64
	ResistanceLevel     float64
	InstantPower        float64
	AveragePower        float64
	TotalEnergy         float64
	EnergyPerHour       float64
	EnergyPerMinute     float64
	HeartRate           float64
	MetabolicEquivalent float64
	ElapsedTime         float64
	RemainingTime       float64
}

func bit(flags uint16, index int) bool {
	return flags&(1<<index) != 0
}

func parseIndoorBikeData(message []byte) IndoorBikeData {
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

	var (
		instantSpeed,
		averageSpeed,
		instantCadence,
		averageCadence,
		totalDistance,
		resistanceLevel,
		instantPower,
		averagePower,
		totalEnergy,
		energyPerHour,
		energyPerMinute,
		heartRate,
		metabolicEquivalent,
		elapsedTime,
		remainingTime float64
	)

	i := 2
	if c1 {
		instantSpeed = float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.01
		i += 2
	}
	if c2 {
		averageSpeed = float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.01
		i += 2
	}
	if c3 {
		instantCadence = float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.5
		i += 2
	}
	if c4 {
		averageCadence = float64(binary.LittleEndian.Uint16(message[i:i+2])) * 0.5
		i += 2
	}
	if c5 {
		totalDistance = float64(int(binary.LittleEndian.Uint32(append(message[i:i+3], 0))))
		i += 3
	}
	if c6 {
		resistanceLevel = float64(int(binary.LittleEndian.Uint16(message[i : i+2])))
		i += 2
	}
	if c7 {
		instantPower = float64(int(binary.LittleEndian.Uint16(message[i : i+2])))
		i += 2
	}
	if c8 {
		averagePower = float64(int(binary.LittleEndian.Uint16(message[i : i+2])))
		i += 2
	}
	if c9 {
		totalEnergy = float64(int(binary.LittleEndian.Uint16(message[i : i+2])))
		energyPerHour = float64(int(binary.LittleEndian.Uint16(message[i+2 : i+4])))
		energyPerMinute = float64(int(message[i+4]))
		i += 5
	}
	if c10 {
		heartRate = float64(message[i])
		i += 1
	}
	if c11 {
		metabolicEquivalent = float64(message[i]) * 0.1
		i += 1
	}
	if c12 {
		elapsedTime = float64(binary.LittleEndian.Uint16(message[i : i+2]))
		i += 2
	}

	if c13 {
		remainingTime = float64(binary.LittleEndian.Uint16(message[i : i+2]))
		i += 2
	}

	return IndoorBikeData{
		InstantSpeed:        instantSpeed,
		AverageSpeed:        averageSpeed,
		InstantCadence:      instantCadence,
		AverageCadence:      averageCadence,
		TotalDistance:       totalDistance,
		ResistanceLevel:     resistanceLevel,
		InstantPower:        instantPower,
		AveragePower:        averagePower,
		TotalEnergy:         totalEnergy,
		EnergyPerHour:       energyPerHour,
		EnergyPerMinute:     energyPerMinute,
		HeartRate:           heartRate,
		MetabolicEquivalent: metabolicEquivalent,
		ElapsedTime:         elapsedTime,
		RemainingTime:       remainingTime,
	}
}
