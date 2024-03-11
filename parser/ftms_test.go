package parser_test

import (
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/minor-industries/codelab/cmd/z2/parser"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseIndoorBikeData(t *testing.T) {
	msg, err := hex.DecodeString("7403e80d6c008c00003600c800030000000000")
	require.NoError(t, err)
	datum := parser.ParseIndoorBikeData(msg)
	fmt.Println(spew.Sdump(datum))

	t.Run("test fields", func(t *testing.T) {
		datum.AllPresentFields(func(series string, v float64) {
			fmt.Println(series, v)
		})
	})
}
