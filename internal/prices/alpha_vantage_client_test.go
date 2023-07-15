package prices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_cleanResponseBody(t *testing.T) {
	t.Run("one digit", func(t *testing.T) {
		responseBytes := []byte(`{"01. hi": "hello"}`)
		out := cleanResponseBody(responseBytes)

		require.Equal(t, `{"hi": "hello"}`, string(out))
	})
	t.Run("two digits", func(t *testing.T) {
		responseBytes := []byte(`{"10. hi": "hello"}`)
		out := cleanResponseBody(responseBytes)

		require.Equal(t, `{"hi": "hello"}`, string(out))
	})
}
