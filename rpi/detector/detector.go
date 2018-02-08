package detector

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/blackjack/webcam"
	"github.com/golang/glog"

	"github.com/timpalpant/yahtzee"
)

const (
	frameWaitTimeout = time.Second

	mJPGPixelFormat = webcam.PixelFormat(1196444237)
	imageWidth      = 1280
	imageHeight     = 960
)

type YahtzeeDetector struct {
	cam *webcam.Webcam
}

func NewYahtzeeDetector() (*YahtzeeDetector, error) {
	cam, err := webcam.Open("/dev/video0") // Open webcam
	if err != nil {
		return nil, err
	}

	_, w, h, err := cam.SetImageFormat(mJPGPixelFormat, imageWidth, imageHeight)
	if err != nil {
		return nil, err
	} else {
		glog.Infof("Webcam image format: %dx%d", w, h)
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

func (d *YahtzeeDetector) GetCurrentRoll() ([]int, error) {
	err := d.cam.WaitForFrame(uint32(frameWaitTimeout.Seconds()))
	if err != nil {
		return nil, err
	}

	frame, err = d.cam.ReadFrame()
	if err != nil {
		return nil, err
	}

	// TODO: Implement image extraction of dice.
	// For now, they have to be entered manually.
	roll := promptRoll()
	if err := saveTrainingData(frame, roll); err != nil {
		glog.Error(err)
	}
	return roll, nil
}

const outputDataDir = "training-data"

var nRolls int

func saveTrainingData(frame []byte, roll []int) error {
	outputFile := path.Join(outputDataDir, fmt.Sprintf("%d.jpg", nRolls))
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(frame); err != nil {
		return err
	}

	f, err := os.OpenFile("rolls.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%d\t%v", nRolls, roll)
	nRolls++
	return err
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

func parseRoll(s string) ([]int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	dice := make([]int, 0, yahtzee.NDice)
	for ; i > 0; i /= 10 {
		die := i % 10
		dice = append(dice, die)
	}

	// Reverse so that dice are in the order they were entered.
	reverse(dice)

	if len(dice) != yahtzee.NDice {
		return dice, fmt.Errorf("Invalid number of dice: %v != %v",
			len(dice), yahtzee.NDice)
	}

	return dice, nil
}

func reverse(a []int) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

func promptRoll() []int {
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
