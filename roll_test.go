package yahtzee

import (
	"math"
	"reflect"
	"testing"
)

func TestNumDice(t *testing.T) {
	cases := []struct {
		roll     Roll
		expected int
	}{
		{NewRollFromBase10Counts(10230), 6},
		{NewRollFromBase10Counts(0), 0},
		{NewRollFromBase10Counts(11111), 5},
		{NewRollFromBase10Counts(50000), 5},
		{NewRollFromBase10Counts(20010), 3},
	}

	for _, tc := range cases {
		if tc.roll.NumDice() != tc.expected {
			t.Errorf("NumDice = %v, expected %v", tc.roll.NumDice(), tc.expected)
		}
	}
}

func TestCounts(t *testing.T) {
	cases := []struct {
		roll     Roll
		expected []int
	}{
		{NewRollFromBase10Counts(10230), []int{0, 3, 2, 0, 1, 0}},
		{NewRollFromBase10Counts(0), []int{0, 0, 0, 0, 0, 0}},
		{NewRollFromBase10Counts(123456), []int{6, 5, 4, 3, 2, 1}},
		{NewRollFromBase10Counts(50000), []int{0, 0, 0, 0, 5, 0}},
		{NewRollFromBase10Counts(20010), []int{0, 1, 0, 0, 2, 0}},
	}

	for _, tc := range cases {
		if !reflect.DeepEqual(tc.roll.Counts(), tc.expected) {
			t.Errorf("Counts = %v, expected %v", tc.roll.Counts(), tc.expected)
		}

		for side, expected := range tc.expected {
			result := tc.roll.CountOf(side + 1)
			if result != expected {
				t.Errorf("Count of %v = %v, expected %v", side+1, result, expected)
			}
		}
	}
}

func TestSubsequentRolls(t *testing.T) {
	cases := []struct {
		input    Roll
		expected int
	}{
		{NewRollFromBase10Counts(0), 252},
		{NewRollFromBase10Counts(11111), 1},
		{NewRollFromBase10Counts(11011), 6},
		{NewRollFromBase10Counts(2002), 6},
		{NewRollFromBase10Counts(202), 6},
		{NewRollFromBase10Counts(3), 21},
	}

	for _, tc := range cases {
		result := tc.input.SubsequentRolls()
		if len(result) != tc.expected {
			t.Errorf("%d rolls, expected %d", len(result), tc.expected)
		}

		// Verify all unique.
		seen := make(map[Roll]struct{})
		for _, roll := range result {
			if _, ok := seen[roll]; ok {
				t.Errorf("Duplicate roll %v", roll)
			}

			seen[roll] = struct{}{}
		}
	}
}

func TestPossibleHolds(t *testing.T) {
	cases := []struct {
		input    Roll
		expected int
	}{
		{NewRollFromBase10Counts(0), 1},
		{NewRollFromBase10Counts(11111), pow(2, 5)},
		{NewRollFromBase10Counts(11011), pow(2, 4)},
		{NewRollFromBase10Counts(2002), pow(3, 2)},
		{NewRollFromBase10Counts(202), pow(3, 2)},
		{NewRollFromBase10Counts(3), pow(4, 1)},
	}

	for _, tc := range cases {
		result := tc.input.PossibleHolds()
		if len(result) != tc.expected {
			t.Errorf("%d holds, expected %d", len(result), tc.expected)
		}

		// Verify all unique.
		seen := make(map[Roll]struct{})
		for _, roll := range result {
			if _, ok := seen[roll]; ok {
				t.Errorf("Duplicate hold %v", roll)
			}

			seen[roll] = struct{}{}
		}
	}
}

func TestProbability(t *testing.T) {
	cases := []Roll{
		NewRollFromBase10Counts(0),
		NewRollFromBase10Counts(11111),
		NewRollFromBase10Counts(11011),
		NewRollFromBase10Counts(2002),
		NewRollFromBase10Counts(202),
		NewRollFromBase10Counts(3),
	}

	for _, input := range cases {
		total := 0.0
		for _, roll := range input.SubsequentRolls() {
			conditionalP := (roll - input).Probability()
			total += conditionalP
		}

		eps := 1e-6
		if math.Abs(total-1.0) > eps {
			t.Errorf("%v: Total probability = %v", input, total)
		}
	}
}

func TestSumOfDice(t *testing.T) {
	cases := []struct {
		roll     Roll
		expected int
	}{
		{NewRollFromBase10Counts(10230), 3*2 + 2*3 + 1*5},
		{NewRollFromBase10Counts(0), 0},
		{NewRollFromBase10Counts(11111), 1 + 2 + 3 + 4 + 5},
		{NewRollFromBase10Counts(50000), 5 * 5},
		{NewRollFromBase10Counts(200010), 1*2 + 2*6},
	}

	for _, tc := range cases {
		if tc.roll.SumOfDice() != tc.expected {
			t.Errorf("SumOfDice = %v, expected %v", tc.roll.SumOfDice(), tc.expected)
		}
	}
}

func TestHasNOfAKind(t *testing.T) {
	cases := []struct {
		roll     Roll
		n        int
		expected bool
	}{
		{NewRollFromBase10Counts(10230), 3, true},
		{NewRollFromBase10Counts(0), 1, false},
		{NewRollFromBase10Counts(11111), 2, false},
		{NewRollFromBase10Counts(50000), 5, true},
		{NewRollFromBase10Counts(200010), 1, true},
	}

	for _, tc := range cases {
		result := tc.roll.HasNOfAKind(tc.n)
		if result != tc.expected {
			t.Errorf("HasNOfAKind(%v) = %v, expected %v", tc.n, result, tc.expected)
		}
	}
}

func TestHasNInARow(t *testing.T) {
	cases := []struct {
		roll     Roll
		n        int
		expected bool
	}{
		{NewRollFromBase10Counts(10230), 2, true},
		{NewRollFromBase10Counts(0), 1, false},
		{NewRollFromBase10Counts(11111), 6, false},
		{NewRollFromBase10Counts(11111), 5, true},
		{NewRollFromBase10Counts(11111), 4, true},
		{NewRollFromBase10Counts(50000), 3, false},
		{NewRollFromBase10Counts(211010), 2, true},
		{NewRollFromBase10Counts(201010), 2, false},
	}

	for _, tc := range cases {
		result := tc.roll.HasNInARow(tc.n)
		if result != tc.expected {
			t.Errorf("HasNInARow(%v) = %v, expected %v", tc.n, result, tc.expected)
		}
	}
}

func TestIsFullHouse(t *testing.T) {
	cases := []struct {
		roll     Roll
		expected bool
	}{
		{NewRollFromBase10Counts(230), true},
		{NewRollFromBase10Counts(0), false},
		{NewRollFromBase10Counts(11111), false},
		{NewRollFromBase10Counts(50000), false},
		{NewRollFromBase10Counts(221000), false},
		{NewRollFromBase10Counts(2030), true},
	}

	for _, tc := range cases {
		result := tc.roll.IsFullHouse()
		if result != tc.expected {
			t.Errorf("%s: IsFullHouse = %v, expected %v", tc.roll, result, tc.expected)
		}
	}
}

func TestDice(t *testing.T) {
	cases := []struct {
		roll     Roll
		expected []int
	}{
		{NewRollFromBase10Counts(230), []int{2, 2, 2, 3, 3}},
		{NewRollFromBase10Counts(0), []int{}},
		{NewRollFromBase10Counts(11111), []int{1, 2, 3, 4, 5}},
		{NewRollFromBase10Counts(50000), []int{5, 5, 5, 5, 5}},
		{NewRollFromBase10Counts(221000), []int{4, 5, 5, 6, 6}},
		{NewRollFromBase10Counts(2030), []int{2, 2, 2, 4, 4}},
	}

	for _, tc := range cases {
		result := tc.roll.Dice()
		if !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("Dice = %v, expected %v", result, tc.expected)
		}
	}
}
