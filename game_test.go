package yahtzee

import (
	"reflect"
	"testing"
)

func TestGameOver(t *testing.T) {
	for game := NewGame(); game < boxesMask; game++ {
		if game.GameOver() {
			t.Errorf("Game %v should not be over", game)
		}
	}

	cases := []struct {
		game     GameState
		expected bool
	}{
		{GameState(boxesMask), true},
		{GameState(0xff0c), false},
		{GameState(6000000), false},
		{GameState(6000000 | boxesMask), true},
		{GameState(6000000 | boxesMask&^(1<<Twos)), false},
	}

	for _, tc := range cases {
		result := tc.game.GameOver()
		if result != tc.expected {
			t.Errorf("%v: expected %v got %v", tc.game, tc.expected, result)
		}
	}
}

func TestFillBox(t *testing.T) {
	game := NewGame()

	for box := Ones; box <= Yahtzee; box++ {
		if game.BoxFilled(box) {
			t.Errorf("Box %v should not be filled at start of game", box)
		}
	}

	game, _ = game.FillBox(Twos, NewRollFromBase10Counts(221))
	for box := Ones; box <= Yahtzee; box++ {
		if box == Twos && !game.BoxFilled(box) {
			t.Errorf("Box %v should be filled", box)
		} else if box != Twos && game.BoxFilled(box) {
			t.Errorf("Box %v should not be filled", box)
		}
	}

	game, _ = game.FillBox(Yahtzee, NewRollFromBase10Counts(50))
	for box := Ones; box <= Yahtzee; box++ {
		if (box == Twos || box == Yahtzee) && !game.BoxFilled(box) {
			t.Errorf("Box %v should be filled", box)
		} else if (box != Twos && box != Yahtzee) && game.BoxFilled(box) {
			t.Errorf("Box %v should not be filled", box)
		}
	}

	// Exceed the UHS bonus threshold.
	game, _ = game.FillBox(Sixes, NewRollFromBase10Counts(500000))
	game, _ = game.FillBox(Fives, NewRollFromBase10Counts(50000))
	game, _ = game.FillBox(Fours, NewRollFromBase10Counts(5000))
	game, _ = game.FillBox(Threes, NewRollFromBase10Counts(131))
	game, _ = game.FillBox(Ones, NewRollFromBase10Counts(122))
	for box := Ones; box <= Sixes; box++ {
		if !game.BoxFilled(box) {
			t.Errorf("Box %v should be filled", box)
		}
	}
}

func TestBonusEligible(t *testing.T) {
	game := NewGame()

	if game.BonusEligible() {
		t.Error("New game should not be bonus eligible")
	}

	game, _ = game.FillBox(Sixes, NewRollFromBase10Counts(500000))
	if game.BonusEligible() {
		t.Error("Game should not be bonus eligible until Yahtzee is filled")
	}

	game, _ = game.FillBox(Yahtzee, NewRollFromBase10Counts(500000))
	if !game.BonusEligible() {
		t.Error("Game should be bonus eligible once Yahtzee is filled")
	}

	game = NewGame()
	game, _ = game.FillBox(Yahtzee, NewRollFromBase10Counts(122))
	if game.BonusEligible() {
		t.Error("Game should not be bonus eligible if a zero is taken")
	}

	game, _ = game.FillBox(Ones, NewRollFromBase10Counts(5))
	if game.BonusEligible() {
		t.Error("Game should not be bonus eligible if a zero is taken")
	}
}

func TestUpperHalfScore(t *testing.T) {
	game := NewGame()
	if game.UpperHalfScore() != 0 {
		t.Error("UHS should be 0 at game start")
	}

	game, _ = game.FillBox(Sixes, NewRollFromBase10Counts(500000))
	if game.UpperHalfScore() != 30 {
		t.Errorf("UHS should be 30, got %v", game.UpperHalfScore())
	}

	game, _ = game.FillBox(Fives, NewRollFromBase10Counts(50000))
	if game.UpperHalfScore() != 55 {
		t.Errorf("UHS should be 55, got %v", game.UpperHalfScore())
	}

	game, _ = game.FillBox(Fours, NewRollFromBase10Counts(5000))
	if game.UpperHalfScore() != 63 {
		t.Errorf("UHS should be capped at 63, got %v", game.UpperHalfScore())
	}

	game, _ = game.FillBox(Ones, NewRollFromBase10Counts(122))
	if game.UpperHalfScore() != 63 {
		t.Errorf("UHS should be capped at 63, got %v", game.UpperHalfScore())
	}
}

func TestAvailableBoxes(t *testing.T) {
	game := NewGame()

	if len(game.AvailableBoxes()) != 13 {
		t.Error("New game should have 13 available boxes")
	}

	game, _ = game.FillBox(Sixes, NewRollFromBase10Counts(212))
	if len(game.AvailableBoxes()) != 12 {
		t.Error("Game should have 12 available boxes")
	}

	for _, box := range []Box{Twos, Yahtzee, ThreeOfAKind, FourOfAKind} {
		game, _ = game.FillBox(box, NewRollFromBase10Counts(212))
	}

	result := game.AvailableBoxes()
	expected := []Box{Ones, Threes, Fours, Fives, FullHouse, SmallStraight, LargeStraight, Chance}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	for _, box := range result {
		game, _ = game.FillBox(box, NewRollFromBase10Counts(212))
	}

	if len(game.AvailableBoxes()) > 0 {
		t.Error("All boxes should be filled")
	}

	if !game.GameOver() {
		t.Error("Game should be over")
	}
}

func TestTotalScore(t *testing.T) {
	game := NewGame()

	if game.TotalScore() != 0 {
		t.Error("Total score of new game should be 0")
	}

	game = game.AddScore(12)
	if game.TotalScore() != 12 {
		t.Error("Total score of game should be 12")
	}

	game = game.AddScore(47)
	if game.TotalScore() != 59 {
		t.Error("Total score of game should be 59")
	}

	game, _ = game.FillBox(Twos, NewRollFromBase10Counts(212))
	if game.TotalScore() != 61 {
		t.Error("Total score of game should be 61")
	}

	game, _ = game.FillBox(Yahtzee, NewRollFromBase10Counts(5))
	if game.TotalScore() != 111 {
		t.Error("Total score of game should be 111")
	}

	if game.Unscored().TotalScore() != 0 {
		t.Error("Unscored game should always have score 0")
	}

	game = game.Unscored().AddScore(MaxScore)
	if game.TotalScore() != MaxScore {
		t.Error("Game should be able to track maximum possible score")
	}
}
