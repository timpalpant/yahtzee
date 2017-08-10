package tricks

const nSides = 6

type Position int

const (
	Ones          Position = 0
	Twos                   = 1
	Threes                 = 2
	Fours                  = 3
	Fives                  = 4
	Sixes                  = 5
	ThreeOfAKind           = 6
	FourOfAKind            = 7
	FullHouse              = 8
	SmallStraight          = 9
	LargeStraight          = 10
	Chance                 = 11
	Yahtzee                = 12
)

func (p Position) IsUpperHalf() bool {
	return p < nSides
}

func (p Position) IsLowerHalf() bool {
	return p >= nSides
}

func isNOfAKind(dice []int, n int) bool {
	counts := make([]int, nSides)
	for _, die := range dice {
		counts[die-1]++
		if counts[die-1] >= n {
			return true
		}
	}

	return false
}

func IsThreeOfAKind(dice []int) bool {
	return isNOfAKind(dice, 3)
}

func IsFourOfAKind(dice []int) bool {
	return isNOfAKind(dice, 4)
}

func IsFullHouse(dice []int) bool {
	counts := make([]int, nSides)
	for _, die := range dice {
		counts[die-1]++

	}

	for _, count := range counts {
		if count != 0 && count != 2 && count != 3 {
			return false
		}
	}

	return true
}

func hasNInARow(dice []int, n int) bool {
	var haveDie int
	for _, die := range dice {
		haveDie |= (1 << uint(die))
	}

	nInARow := 0
	for i := uint(1); i <= nSides; i++ {
		bit := haveDie & (1 << i)
		if bit != 0 {
			nInARow++
		} else {
			nInARow = 0
		}

		if nInARow >= n {
			return true
		} else if n-nInARow > nSides-int(i) {
			// Not enough remaining dice that we could possibly
			// have n in a row.
			return false
		}
	}

	return false
}

func IsSmallStraight(dice []int) bool {
	return hasNInARow(dice, 4)
}

func IsLargeStraight(dice []int) bool {
	return hasNInARow(dice, 5)
}

func IsYahtzee(dice []int) bool {
	for _, die := range dice[1:] {
		if die != dice[0] {
			return false
		}
	}

	return true
}
