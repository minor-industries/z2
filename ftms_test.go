package main

import (
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_parseIndoorBikeData(t *testing.T) {
	msg, err := hex.DecodeString("74032a081c003b00003b001c00010000")
	require.NoError(t, err)
	ibd := parseIndoorBikeData(msg)
	fmt.Println(spew.Sdump(ibd))
}
