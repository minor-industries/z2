package workouts

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestData(t *testing.T) {
	errCh := make(chan error)

	db, err := database.Get(os.ExpandEnv("$HOME/.z2/z2.db"), errCh)
	require.NoError(t, err)

	data, err := GenerateData(db.GetORM(), "", "")
	require.NoError(t, err)

	for _, d := range data {
		fmt.Println(d.Info)
		for _, row := range d.Data {
			fmt.Println("\t", strings.Join(row[:], "  "))
		}
	}
}
