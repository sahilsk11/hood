package prices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_determineColumnOrdering(t *testing.T) {
	requiredHeaders := []string{"symbol", "price", "date"}
	headerRow := []string{"SYMBOL", "PRICE", "DATE"}

	out, err := determineColumnOrdering(headerRow, requiredHeaders)
	require.NoError(t, err)
	require.Equal(t, map[string]int{
		"symbol": 0,
		"price":  1,
		"date":   2,
	}, out)
}
