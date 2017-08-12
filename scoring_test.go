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
		{Roll(11111), Yahtzee, 0},
		{Roll(11111), Ones, 1},
		{Roll(11111), SmallStraight, 30},
		{Roll(11111), LargeStraight, 40},
		{Roll(11120), LargeStraight, 0},
		{Roll(11120), SmallStraight, 30},
		{Roll(50000), Yahtzee, 50},
		{Roll(50000), FourOfAKind, 25},
		{Roll(132), Twos, 6},
		{Roll(132), Threes, 3},
		{Roll(3002), FullHouse, 25},
		{Roll(4001), FullHouse, 0},
		{Roll(4001), Chance, 17},
		{Roll(4001), ThreeOfAKind, 17},
		{Roll(11201), ThreeOfAKind, 0},
		{Roll(11201), FourOfAKind, 0},
		{Roll(401), FourOfAKind, 13},
	}

	for _, tc := range cases {
		result := tc.box.Score(tc.roll)
		if result != tc.expected {
			t.Errorf("%v: Score = %v, expected %v", tc, result, tc.expected)
		}
	}
}
