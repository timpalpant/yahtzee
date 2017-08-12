package yahtzee

type Box uint

const (
	Ones Box = iota
	Twos
	Threes
	Fours
	Fives
	Sixes
	ThreeOfAKind
	FourOfAKind
	FullHouse
	SmallStraight
	LargeStraight
	Chance
	Yahtzee
)

func (b Box) IsUpperHalf() bool {
	return b <= Sixes
}

// BoxScore returns the score that would be received for
// playing the given roll in the given box.
//
// Note it does not include any bonuses.
func (b Box) Score(roll Roll) int {
	if b.IsUpperHalf() {
		side := int(b + 1)
		return side * roll.CountOf(side)
	}

	switch b {
	case ThreeOfAKind:
		if roll.HasNOfAKind(3) {
			return roll.SumOfDice()
		}
	case FourOfAKind:
		if roll.HasNOfAKind(4) {
			return roll.SumOfDice()
		}
	case FullHouse:
		if roll.IsFullHouse() {
			return 25
		}
	case SmallStraight:
		if roll.HasNInARow(4) {
			return 30
		}
	case LargeStraight:
		if roll.HasNInARow(5) {
			return 40
		}
	case Chance:
		return roll.SumOfDice()
	case Yahtzee:
		if IsYahtzee(roll) {
			return 50
		}
	}

	return 0
}

func IsYahtzee(roll Roll) bool {
	return roll.HasNOfAKind(5)
}
