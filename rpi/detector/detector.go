package detector

import (
	"fmt"
	"time"

	"github.com/blackjack/webcam"
)

const frameWaitTimeout = time.Second

type Roll [5]int

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

func (d *YahtzeeDetector) GetCurrentRoll() (Roll, error) {
	err := d.cam.WaitForFrame(uint32(frameWaitTimeout.Seconds()))
	if err != nil {
		return Roll{}, err
	}


	frame, err := d.cam.ReadFrame()
	if err != nil {
		return Roll{}, err
	}

	fmt.Println(frame)

	return Roll{1, 2, 3, 4, 5}, nil
}
