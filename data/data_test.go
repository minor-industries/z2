package data

import (
	"fmt"
	"github.com/minor-industries/rtgraph/database"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestData(t *testing.T) {
	errCh := make(chan error)

	db, err := database.Get(os.ExpandEnv("$HOME/.z2/z2.db"), errCh)
	require.NoError(t, err)

	orm := db.GetORM()

	err = orm.AutoMigrate(&Marker{})
	require.NoError(t, err)

	var markers []Marker
	tx := orm.Where("ref = ?", "bike").Order("timestamp asc").Find(&markers)
	require.NoError(t, tx.Error)

	fmt.Println(len(markers))

	for _, marker := range markers {
		fmt.Println(marker)
	}
}
