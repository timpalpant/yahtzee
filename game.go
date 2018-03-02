package yahtzee

import (
	"fmt"

	"github.com/golang/glog"
)

const (
	MaxGame       GameState = 64 << shiftUHS
	MaxScoredGame GameState = MaxGame + (MaxScore << shiftScore)
	NumTurns                = int(Yahtzee + 1)

	UpperHalfBonusThreshold = 63
	UpperHalfBonus          = 35
	YahtzeeBonus            = 100
	MaxScore                = 1575
)

const (
	bonusBit     uint = uint(Yahtzee + 1)
	shiftUHS     uint = bonusBit + 1
	boxesMask         = (1 << bonusBit) - 1
	shiftScore   uint = shiftUHS + 6 + 1
	unscoredMask      = (1 << shiftScore) - 1
)

// Each distinct game is represented by an integer as follows:
//
//   1. The lowest 13 bits represent whether a box has been filled.
//      Bits 0-5 are the Upper half (ones, twos, ... sixes).
//      Bits 6-12 are the Lower half (three of a kind ... yahtzee)
//   2. Bit 13 represents whether you are eligible for the bonus,
//      meaning that you have previously filled the Yahtzee for points.
//      Therefore bit 13 can only be set if bit 12 is also set.
//   3. Bits 14-19 represent the upper half score in
//      the range [0, 63]. Since for all upper half scores >= 63 you
//      get the upper half bonus, they are equivalent and the upper
//      half score is capped at 63.
//   4. Bits 20-32 store the total score. The total score is optional
//      and doesn't need to be set; the unscored GameState can be
//      obtained by calling Unscored().
//
// This means that all games are represented by an integer < 2^20 (MaxGame).
type GameState uint

func NewGame() GameState {
	return GameState(0)
}

// Enumerate all reachable game states.
func AllGameStates(withScore bool) []GameState {
	games := make(map[GameState]struct{})
	start := NewGame()
	games[start] = struct{}{}
	enumerateGames(games, start, withScore)

	result := make([]GameState, 0, len(games))
	for game := range games {
		result = append(result, game)
	}

	return result
}

func enumerateGames(games map[GameState]struct{}, game GameState, withScore bool) {
	if game.GameOver() {
		return
	}

	for _, roll := range AllDistinctRolls() {
		for _, box := range game.AvailableBoxes() {
			newGame, _ := game.FillBox(box, roll)
			if !withScore {
				newGame = newGame.Unscored()
			}

			if _, ok := games[newGame]; !ok {
				games[newGame] = struct{}{}
				if len(games)%100000 == 0 {
					glog.V(1).Infof("Enumerated %v games", len(games))
				}
				enumerateGames(games, newGame, withScore)
			}
		}
	}
}

// Unscored returns this GameState with the total score set to 0.
func (game GameState) Unscored() GameState {
	return game & unscoredMask
}

func (game GameState) Turn() int {
	return NumTurns - game.TurnsRemaining()
}

func (game GameState) TurnsRemaining() int {
	return len(game.AvailableBoxes())
}

func (game GameState) GameOver() bool {
	return (game & boxesMask) == boxesMask
}

func (game GameState) BoxFilled(b Box) bool {
	return (game & (1 << b)) != 0
}

func (game GameState) BonusEligible() bool {
	return (game & (1 << bonusBit)) != 0
}

func (game GameState) UpperHalfScore() int {
	return int(game.Unscored() >> shiftUHS)
}

func (game GameState) TotalScore() int {
	return int(game >> shiftScore)
}

// Statically pre-computed set of available boxes for each game.
var availableBoxes = computeAvailableBoxes()

func (game GameState) AvailableBoxes() []Box {
	return availableBoxes[game.Unscored()]
}

func computeAvailableBoxes() [][]Box {
	result := make([][]Box, MaxGame)
	for game := GameState(0); game < MaxGame; game++ {
		available := make([]Box, 0)
		for box := Ones; box <= Yahtzee; box++ {
			if !game.BoxFilled(box) {
				available = append(available, box)
			}
		}

		result[game] = available
	}

	return result
}

func (game GameState) SetBoxFilled(box Box) GameState {
	return game | (1 << box)
}

func (game GameState) AddUpperHalfScore(score int) GameState {
	newGame := game + GameState(score<<shiftUHS)

	// Cap upper half score at bonus threshold since all values > threshold
	// are equivalent in terms of getting the bonus.
	prevUHS := game.UpperHalfScore()
	if prevUHS+score >= UpperHalfBonusThreshold {
		newGame -= GameState((prevUHS + score - UpperHalfBonusThreshold) << shiftUHS)
	}

	return newGame
}

func (game GameState) SetBonusEligible() GameState {
	return game | (1 << bonusBit)
}

func (game GameState) AddScore(score int) GameState {
	return game + GameState(score<<shiftScore)
}

func (game GameState) FillBox(box Box, roll Roll) (GameState, int) {
	if game.BoxFilled(box) {
		panic(fmt.Errorf("trying to play already filled box %v", box))
	} else if roll.NumDice() != NDice {
		panic(fmt.Errorf("trying to play incomplete roll with %v dice", roll.NumDice()))
	}

	value := box.Score(roll)

	newGame := game.SetBoxFilled(box)
	if box == Yahtzee && value != 0 {
		newGame = newGame.SetBonusEligible()
	}

	prevUHS := game.UpperHalfScore()
	if value != 0 && box.IsUpperHalf() && prevUHS < UpperHalfBonusThreshold {
		newGame = newGame.AddUpperHalfScore(value)
		if newGame.UpperHalfScore() >= UpperHalfBonusThreshold {
			value += UpperHalfBonus
		}
	}

	// Joker rule: Roll can be played in any box for points,
	// if the corresponding upper half box is already filled.
	if game.BonusEligible() && IsYahtzee(roll) {
		value += YahtzeeBonus

		nativeBox := nativeUpperHalfBox(roll)
		if game.BoxFilled(nativeBox) {
			switch box {
			case FullHouse:
				value += 25
			case SmallStraight:
				value += 30
			case LargeStraight:
				value += 40
			}
		}
	}

	newGame = newGame.AddScore(value)
	return newGame, value
}

func (game GameState) String() string {
	return fmt.Sprintf("{ID: %d, Available: %v, BonusEligible: %v, UpperHalf: %v}",
		game, game.AvailableBoxes(), game.BonusEligible(), game.UpperHalfScore())
}

func nativeUpperHalfBox(yahtzeeRoll Roll) Box {
	side := yahtzeeRoll.One()
	return Box(side - 1)
}

type TurnStep int

const (
	Begin TurnStep = iota
	Hold1
	Hold2
	FillBox
)
