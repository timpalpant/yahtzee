package detector

import (
	"fmt"
	"time"

	"github.com/blackjack/webcam"

	"github.com/timpalpant/yahtzee"
)

const frameWaitTimeout = time.Second

type YahtzeeDetector struct {
	cam *webcam.Webcam
}

func NewYahtzeeDetector() (*YahtzeeDetector, error) {
	cam, err := webcam.Open("/dev/video0") // Open webcam
	if err != nil {
		return nil, err
	}

	err = cam.StartStreaming()
	if err != nil {
		return nil, err
	}

	return &YahtzeeDetector{cam}, nil
}

func (d *YahtzeeDetector) Close() error {
	return d.cam.Close()
}

func (d *YahtzeeDetector) GetCurrentRoll() (yahtzee.Roll, error) {
	err := d.cam.WaitForFrame(uint32(frameWaitTimeout.Seconds()))
	if err != nil {
		return yahtzee.NewRoll(), err
	}

	frame, err := d.cam.ReadFrame()
	if err != nil {
		return yahtzee.NewRoll(), err
	}

	// TODO: Implement image extraction of dice.
	// For now, they have to be entered manually.
	roll := promptRoll()
	return roll, nil
}

var stdin = bufio.NewReader(os.Stdin)

func prompt(msg string) string {
	fmt.Print(msg)
	result, err := stdin.ReadString('\n')
	if err != nil {
		panic(err)
	}

	return strings.TrimRight(result, "\n")
}

func parseRoll(s string) (yahtzee.Roll, error) {
	roll := yahtzee.NewRoll()

	i, err := strconv.Atoi(s)
	if err != nil {
		return roll, err
	}

	for ; i > 0; i /= 10 {
		die := i % 10
		roll = roll.Add(die)
	}

	if roll.NumDice() != yahtzee.NDice {
		return roll, fmt.Errorf("Invalid number of dice: %v != %v",
			roll.NumDice(), yahtzee.NDice)
	}

	return roll, nil
}

func promptRoll() yahtzee.Roll {
	for {
		rollStr := prompt("Enter roll: ")
		roll, err := parseRoll(rollStr)
		if err != nil {
			fmt.Printf("Invalid roll: %v\n", err)
			continue
		}

		return roll
	}
}
