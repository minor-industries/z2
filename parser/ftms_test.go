package parser_test

import (
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/minor-industries/codelab/cmd/bike/parser"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_parseIndoorBikeData(t *testing.T) {
	msg, err := hex.DecodeString("7403e80d6c008c00003600c800030000000000")
	require.NoError(t, err)
	datum := parser.ParseIndoorBikeData(msg)
	fmt.Println(spew.Sdump(datum))

	t.Run("test fields", func(t *testing.T) {
		val := reflect.ValueOf(datum).Elem()
		//structName := val.Type().String()

		for i := 0; i < val.NumField(); i++ {
			series := val.Type().Field(i).Tag.Get("series")
			if series == "" {
				continue // series not tagged
			}

			if !val.Field(i).Field(1).Bool() {
				continue // field value not present
			}

			v := val.Field(i).Field(0).Float()
			fmt.Println(series, v)
		}
	})
}
