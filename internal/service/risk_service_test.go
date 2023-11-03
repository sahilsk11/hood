package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_createGroupsWithTreshold(t *testing.T) {
	graph := graph{
		"A": {"B": 0.8, "C": 0.1, "D": 0.1},
		"B": {"A": 0.8, "C": 0.8, "D": 0.8},
		"C": {"A": 0.1, "B": 0.8, "D": 0.8},
		"D": {"A": 0.1, "B": 0.8, "C": 0.8},
	}
	groups := createGroupsWithThreshold(graph, 0.7)

	require.Equal(
		t,
		[][]string{
			{"A", "B"},
			{"B", "C", "D"},
		},
		groups,
	)
}

func Test_keepLargestGroups(t *testing.T) {
	out := keepLargestGroups([][]string{
		{"A", "B"},
		{"B", "C", "D"},
	}, map[string]float64{
		"A": 100,
		"B": 10,
		"C": 1,
		"D": 1,
	})

	require.Equal(
		t,
		[][]string{
			{"A", "B"},
		},
		out,
	)
}
