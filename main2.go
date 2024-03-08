// This example scans and then connects to a specific Bluetooth peripheral
// and then displays all of the services and characteristics.
//
// To run this on a desktop system:
//
//	go run ./examples/discover EE:74:7D:C9:2A:68
//
// To run this on a microcontroller, change the constant value in the file
// "mcu.go" to set the MAC address of the device you want to discover.
// Then, flash to the microcontroller board like this:
//
//	tinygo flash -o circuitplay-bluefruit ./examples/discover
//
// Once the program is flashed to the board, connect to the USB port
// via serial to view the output.
package main

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"tinygo.org/x/bluetooth"
)

func connectAddress() string {
	return "FC:38:34:32:0D:69"
}

var adapter = bluetooth.DefaultAdapter

func main() {
	fmt.Println("enabling")

	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	ch := make(chan bluetooth.ScanResult, 1)

	// Start scanning.
	fmt.Println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		fmt.Println("found device:", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == connectAddress() {
			adapter.StopScan()
			ch <- result
		}
	})

	var device bluetooth.Device
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println("connected to ", result.Address.String())
	}

	// get services
	fmt.Println("discovering services/characteristics")
	srvcs, err := device.DiscoverServices(nil)
	must("discover services", err)

	// buffer to retrieve characteristic data
	buf := make([]byte, 255)

	for _, srvc := range srvcs {
		fmt.Println("- service", srvc.UUID().String())

		chars, err := srvc.DiscoverCharacteristics(nil)
		if err != nil {
			fmt.Println(err)
		}
		for _, char := range chars {
			fmt.Println("-- 16 bit", srvc.Is16Bit())
			fmt.Println("-- characteristic", char.UUID().String())
			n, err := char.Read(buf)
			if err != nil {
				fmt.Println("    ", err.Error())
			} else {
				fmt.Println("    data bytes", strconv.Itoa(n))
				fmt.Println("    value =", string(buf[:n]))
			}

			if srvc.UUID().String() == "00001826-0000-1000-8000-00805f9b34fb" &&
				char.UUID().String() == "00002ad2-0000-1000-8000-00805f9b34fb" {
				if err := char.EnableNotifications(func(buf []byte) {
					fmt.Println("notify", hex.Dump(buf))
				}); err != nil {
					fmt.Println("error enabling notifications:", err.Error())
				}
			}
		}
	}

	select {}

	err = device.Disconnect()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("done")
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
