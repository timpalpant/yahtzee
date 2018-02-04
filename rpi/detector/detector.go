package detector

import ()

type Roll [5]int

type YahtzeeDetector struct {
}

func NewYahtzeeDetector() *YahtzeeDetector {
	return &YahtzeeDetector{}
}

func (d *YahtzeeDetector) GetCurrentRoll() (Roll, error) {
	return Roll{1, 2, 3, 4, 5}, nil
}
