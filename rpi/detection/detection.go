package detection

import (
	"gocv.io/x/gocv"
)

type Roll [5]int

func GetCurrentRoll() (Roll, error) {
	return Roll{1, 2, 3, 4, 5}, nil
}
