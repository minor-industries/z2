//go:build !wasm

package source

import (
	"encoding/hex"
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/pkg/errors"
	"strings"
	"tinygo.org/x/bluetooth"
)

func Scan() error {
	var adapter = bluetooth.DefaultAdapter

	var err error
	enableOnce.Do(func() {
		fmt.Println("enabling")
		err = adapter.Enable()
	})
	if err != nil {
		return errors.Wrap(err, "enable adapter")
	}

	fmt.Println("scanning...")

	found := set.Set[string]{}

	err = adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

		key := strings.Join([]string{
			result.Address.String(),
			result.LocalName(),
		}, "::")

		if found.Has(key) {
			return
		}

		found.Add(key)

		mfd := showManufacturerData(result.ManufacturerData())

		fmt.Printf(
			"%s: %*d [%s] %s\n",
			result.Address.String(),
			4, result.RSSI,
			result.LocalName(),
			mfd,
		)
	})

	return nil
}

func showManufacturerData(items []bluetooth.ManufacturerDataElement) string {
	results := make([]string, len(items))

	for i, item := range items {
		results[i] = fmt.Sprintf("0x%04x %s", item.CompanyID, hex.EncodeToString(item.Data))
	}

	return strings.Join(results, ", ")
}
