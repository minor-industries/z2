package source

import (
	"fmt"
	"testing"
	"tinygo.org/x/bluetooth"
)

func TestUUIDs(t *testing.T) {
	fmt.Println(bluetooth.ServiceUUIDHeartRate.String())
	fmt.Println(bluetooth.CharacteristicUUIDHeartRateMeasurement.String())
	fmt.Println(bluetooth.ServiceUUIDFitnessMachine.String())
	fmt.Println(bluetooth.CharacteristicUUIDIndoorBikeData.String())

}
