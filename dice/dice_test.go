package dice

import (
	"math"
	"reflect"
	"testing"
)

func binomial(n, k int) int {
	return factorial(n) / (factorial(k) * factorial(n-k))
}

func nRolls(n, k int) int {
	return binomial(n+k-1, n)
}

func TestAllPossibleRolls(t *testing.T) {
	for n := 1; n <= 5; n++ {
		rolls := AllPossibleRolls(n)
		expected := nRolls(n, 6)
		if len(rolls) != expected {
			t.Errorf("Expected %d rolls, got %d", expected, len(rolls))
		}

		for i := 1; i < len(rolls); i++ {
			for j := i + 1; j < len(rolls); j++ {
				if reflect.DeepEqual(rolls[i], rolls[j]) {
					t.Error("Found duplicate roll!")
				}
			}
		}
	}
}

func TestProbability(t *testing.T) {
	cases := []struct {
		roll     []int
		expected float64
	}{
		{
			roll:     []int{1, 1, 1, 1, 1},
			expected: 1.0 / math.Pow(6, 5),
		},
		{
			roll:     []int{1, 1, 1, 1, 2},
			expected: 5.0 / math.Pow(6, 5),
		},
		{
			roll:     []int{1, 2, 3, 4, 6},
			expected: float64(factorial(5)) / math.Pow(6, 5),
		},
	}

	for _, tc := range cases {
		result := Probability(tc.roll)
		if result != tc.expected {
			t.Errorf("Roll: %v, got: %v wanted %v",
				tc.roll, result, tc.expected)
		}
	}
}
