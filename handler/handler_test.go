package handler_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBleHandler_Handle(t *testing.T) {
	db, err := database.Get(os.ExpandEnv("$HOME/bike.db"))
	require.NoError(t, err)

	handler, err := handler.NewBleHandler(db)
	require.NoError(t, err)
	_ = handler

	content, err := os.ReadFile("../testdata/raw.txt")
	require.NoError(t, err)

	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		t0, err := strconv.Atoi(parts[0])
		require.NoError(t, err)
		t1 := time.UnixMilli(int64(t0))
		fmt.Println(t1, parts[1])

		raw, err := hex.DecodeString(parts[1])
		require.NoError(t, err)

		err = handler.Handle(t1, raw)
		require.NoError(t, err)
	}

}
