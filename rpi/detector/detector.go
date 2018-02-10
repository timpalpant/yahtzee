package detector

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"gocv.io/x/gocv"

	"github.com/timpalpant/yahtzee"
)

const (
	imageWidth  = 1280
	imageHeight = 960
)

type YahtzeeDetector struct {
	client *Client

	outDir string
	n      int

	// mu protects currentImg
	mu sync.Mutex
	currentImg gocv.Mat
	closeCh chan error
}

func NewYahtzeeDetector(device int, uri, outDir string) (*YahtzeeDetector, error) {
	cam, err := gocv.VideoCaptureDevice(device)
	if err != nil {
		return nil, err
	}

	cam.Set(gocv.VideoCaptureFrameWidth, imageWidth)
	cam.Set(gocv.VideoCaptureFrameHeight, imageHeight)

	if !cam.IsOpened() {
		cam.Close()
		return nil, fmt.Errorf("error opening device %v", device)
	}

	if outDir != "" {
		os.MkdirAll(outDir, os.ModePerm)
	}

	d := &YahtzeeDetector{
		client: NewClient(uri),
		currentImg: gocv.NewMat(),
		outDir: outDir,
		closeCh: make(chan error),
	}

	go d.streamFrames(cam)
	return d, nil
}

func (d *YahtzeeDetector) Close() error {
	d.closeCh <- nil
	err := <-d.closeCh
	d.currentImg.Close()
	return err
}

func (d *YahtzeeDetector) GetCurrentRoll() ([]int, error) {
	glog.V(2).Infof("Getting current webcam image")
	img := gocv.NewMat()
	defer img.Close()
	d.getCurrentImage(img)
	if img.Empty() {
		return nil, fmt.Errorf("current image is empty")
	}

	frame, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Sending image to image processing service")
	dice, err := d.client.GetDiceFromImage(frame)
	if err != nil {
		return nil, err
	}

	// TODO: Implement image extraction of dice.
	// For now, they have to be entered manually.
	glog.Infof("Detected dice: %v", dice)
	roll := promptRoll()

	if err := d.saveTrainingData(img, roll); err != nil {
		glog.Error(err)
	}

	return roll, nil
}

func (d *YahtzeeDetector) streamFrames(webcam *gocv.VideoCapture) error {
	img := gocv.NewMat()
	defer img.Close()

Loop:
	for {
		select {
		case <-d.closeCh:
			break Loop
		default:
		}

		if ok := webcam.Read(img); !ok {
			glog.Error("error reading frame from webcam")
			continue
		}

		if img.Empty() {
			glog.Warning("got an empty image from webcam")
			continue
		}

		d.setCurrentImage(img)
	}

	return webcam.Close()
}

func (d *YahtzeeDetector) setCurrentImage(img gocv.Mat) {
	d.mu.Lock()
	defer d.mu.Unlock()
	img.CopyTo(d.currentImg)
}

func (d *YahtzeeDetector) getCurrentImage(img gocv.Mat) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.currentImg.CopyTo(img)
}

func (d *YahtzeeDetector) saveTrainingData(img gocv.Mat, roll []int) error {
	outputFile := path.Join(d.outDir, fmt.Sprintf("%d.jpg", d.n))
	glog.V(1).Infof("Saving: %v", outputFile)
	if !gocv.IMWrite(outputFile, img) {
		return fmt.Errorf("error saving mage to output file")
	}

	rollsFile := path.Join(d.outDir, "rolls.txt")
	f, err := os.OpenFile(rollsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%d\t%v\n", d.n, roll)
	d.n++
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
