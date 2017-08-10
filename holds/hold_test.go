package holds

import (
	"reflect"
	"testing"
)

func TestAllPossibleHolds(t *testing.T) {
	for i := 1; i <= 5; i++ {
		result := AllPossibleHolds(i)
		if len(result) != pow(2, i) {
			t.Errorf("Expected %v results, got %v", pow(2, i), len(result))
		}

		for _, hold := range result {
			if len(hold) != i {
				t.Errorf("Hold should have len %v, got %v", i, len(hold))
			}
		}
	}
}

func TestKeep(t *testing.T) {
	cases := []struct {
		roll     []int
		hold     []bool
		expected []int
	}{
		{
			roll:     []int{1, 2, 3, 4, 5},
			hold:     []bool{true, true, false, true, false},
			expected: []int{1, 2, 4},
		},
		{
			roll:     []int{1, 2, 3, 1, 1},
			hold:     []bool{true, true, false, true, false},
			expected: []int{1, 2, 1},
		},
		{
			roll:     []int{1, 2, 3, 1, 1},
			hold:     []bool{false, false, false, false, false},
			expected: []int{},
		},
		{
			roll:     []int{5, 4, 3, 2, 1},
			hold:     []bool{true, true, true, true, true},
			expected: []int{5, 4, 3, 2, 1},
		},
	}

	for _, tc := range cases {
		result := Keep(tc.roll, tc.hold)
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("Expected %v, got %v", tc.expected, result)
		}
	}
}

func pow(n, k int) int {
	result := 1
	for i := 0; i < k; i++ {
		result *= n
	}

	return result
}
