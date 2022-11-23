package data_ingestion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTradesFromOutfile(t *testing.T) {
	trades, _, err := ParseEntriesFromOutfile()
	require.NoError(t, err)
	require.Equal(t, 562, len(trades))
}
