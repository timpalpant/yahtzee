package yahtzee

import (
	"testing"
)

func TestBoxScore(t *testing.T) {
	cases := []struct {
		roll     Roll
		box      Box
		expected int
	}{
		{NewRollFromBase10Counts(11111), Yahtzee, 0},
		{NewRollFromBase10Counts(11111), Ones, 1},
		{NewRollFromBase10Counts(11111), SmallStraight, 30},
		{NewRollFromBase10Counts(11111), LargeStraight, 40},
		{NewRollFromBase10Counts(11120), LargeStraight, 0},
		{NewRollFromBase10Counts(11120), SmallStraight, 30},
		{NewRollFromBase10Counts(50000), Yahtzee, 50},
		{NewRollFromBase10Counts(50000), FourOfAKind, 25},
		{NewRollFromBase10Counts(132), Twos, 6},
		{NewRollFromBase10Counts(132), Threes, 3},
		{NewRollFromBase10Counts(3002), FullHouse, 25},
		{NewRollFromBase10Counts(4001), FullHouse, 0},
		{NewRollFromBase10Counts(4001), Chance, 17},
		{NewRollFromBase10Counts(4001), ThreeOfAKind, 17},
		{NewRollFromBase10Counts(11201), ThreeOfAKind, 0},
		{NewRollFromBase10Counts(11201), FourOfAKind, 0},
		{NewRollFromBase10Counts(401), FourOfAKind, 13},
	}

	for _, tc := range cases {
		result := tc.box.Score(tc.roll)
		if result != tc.expected {
			t.Errorf("%v: Score = %v, expected %v", tc, result, tc.expected)
		}
	}
}
