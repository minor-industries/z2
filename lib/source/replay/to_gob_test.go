package replay

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_toGob(t *testing.T) {
	err := toGob()
	require.NoError(t, err)
}
