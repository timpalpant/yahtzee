package holds

import (
	"testing"
)

func TestAllDistinctHolds(t *testing.T) {
	cases := []struct {
		roll               []int
		expectedNumResults int
	}{
		{
			roll:               []int{1, 1, 1, 1, 1},
			expectedNumResults: 6,
		},
		{
			roll:               []int{1, 1, 1, 1, 2},
			expectedNumResults: 10,
		},
		{
			roll:               []int{1, 1, 1, 2, 3},
			expectedNumResults: 16,
		},
		{
			roll:               []int{1, 1, 1, 2, 2},
			expectedNumResults: 12,
		},
	}

	for _, tc := range cases {
		result := AllDistinctHolds(tc.roll)
		if len(result) != tc.expectedNumResults {
			t.Errorf("Expected %v results, got %v", tc.expectedNumResults, len(result))
		}

		for _, kept := range result {
			input := make(map[int]int, len(tc.roll))
			for _, die := range tc.roll {
				input[die]++
			}

			for _, keep := range kept {
				if n := input[keep]; n <= 0 {
					t.Errorf("Kept %v not available in input %v", kept, tc.roll)
				}

				input[keep]--
			}
		}
	}
}
