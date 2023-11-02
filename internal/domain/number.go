package domain

import (
	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
)

// kind of experimental. but i've been toying
// with the idea of using typed numbers that
// represent the actual unit we want. this
// reduces ambiguity when looking at a float
// and wondering what unit it represents

type Percent float64
type PercentData []Percent

func (p Percent) AsFraction() float64 {
	return float64(p)
}

func (p Percent) AsPercent() float64 {
	return p.AsFraction() * 100
}

func PercentFromFraction(f float64) Percent {
	return Percent(f)
}

func (pd PercentData) ToStatsData() stats.Float64Data {
	out := make(stats.Float64Data, len(pd))
	for i, n := range pd {
		out[i] = n.AsPercent()
	}
	return out
}

type Price float64
type USD decimal.Decimal
type AllocationFraction float64
