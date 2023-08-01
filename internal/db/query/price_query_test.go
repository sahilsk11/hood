package db

import (
	"hood/internal/db/models/postgres/public/model"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func TestGetAdjustedPrices(t *testing.T) {
	t.Run("asset split", func(t *testing.T) {
		tx := SetupTestDb(t)
		_, err := AddAssetsSplits(tx, []*model.AssetSplit{
			{
				Symbol: "AAPL",
				Ratio:  4,
				Date:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		})
		require.NoError(t, err)

		_, err = AddPrices(tx, []model.Price{
			{
				Symbol: "AAPL",
				Price:  dec(100),
				Date:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		})
		require.NoError(t, err)

		out, err := GetAdjustedPrices(tx, []string{"AAPL"}, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)

		require.Equal(
			t,
			"",
			cmp.Diff(
				[]model.Price{
					{
						Symbol: "AAPL",
						Price:  dec(25),
						Date:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
				out,
				cmpopts.IgnoreFields(model.Price{}, "PriceID"),
				cmpopts.IgnoreFields(model.Price{}, "UpdatedAt"),
			),
		)
	})
}
