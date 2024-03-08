package main

import (
	"github.com/minor-industries/codelab/cmd/bike/database"
	"github.com/minor-industries/codelab/cmd/bike/handler"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_serve(t *testing.T) {
	db, err := database.Get(os.ExpandEnv("$HOME/bike.db"))
	require.NoError(t, err)

	handler, err := handler.NewBleHandler(db)
	require.NoError(t, err)
	_ = handler

	serve(handler)
}
