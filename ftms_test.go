package main

import (
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseIndoorBikeData(t *testing.T) {
	msg, err := hex.DecodeString("7403e80d6c008c00003600c800030000000000")
	require.NoError(t, err)
	ibd := parseIndoorBikeData(msg)
	fmt.Println(spew.Sdump(ibd))
}
