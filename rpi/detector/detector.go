package detector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
	"gocv.io/x/gocv"

	"github.com/timpalpant/yahtzee"
)

const (
	imageWidth  = 1280
	imageHeight = 720
	fps = 5
	mJPGFormat = 1196444237

	retryLimit = 3
)

var stdin = bufio.NewReader(os.Stdin)

type YahtzeeDetector struct {
	client *Client

	outDir   string
	annotate bool

	// mu protects currentImg
	mu         sync.Mutex
	currentImg gocv.Mat
	closeCh    chan error
}

func NewYahtzeeDetector(device int, uri, outDir string, annotate bool) (*YahtzeeDetector, error) {
	cam, err := gocv.VideoCaptureDevice(device)
	if err != nil {
		return nil, err
	}

	cam.Set(gocv.VideoCaptureFPS, fps)
	cam.Set(gocv.VideoCaptureFrameWidth, imageWidth)
	cam.Set(gocv.VideoCaptureFrameHeight, imageHeight)
	cam.Set(gocv.VideoCaptureFormat, mJPGFormat)

	if !cam.IsOpened() {
		cam.Close()
		return nil, fmt.Errorf("error opening device %v", device)
	}

	if outDir != "" {
		os.MkdirAll(outDir, os.ModePerm)
	}

	var client *Client
	if uri != "" {
		client = NewClient(uri)
	}

	d := &YahtzeeDetector{
		client:     client,
		outDir:     outDir,
		annotate:   annotate,
		currentImg: gocv.NewMat(),
		closeCh:    make(chan error),
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
	for retry := 0; retry < retryLimit; retry++ {
		dice, err := d.getRoll()
		if err != nil {
			glog.Error(err)
		} else {
			return dice, nil
		}
	}

	return nil, fmt.Errorf("failed to get dice after %d tries", retryLimit)
}

func (d *YahtzeeDetector) getRoll() ([]int, error) {
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

	id, err := d.saveImage(frame)
	if err != nil {
		glog.Error(err)
	}

	var dice []int
	if d.client != nil {
		glog.V(1).Infof("Sending image %v to image processing service", id)
		dice, err = d.client.GetDiceFromImage(frame)
		if err != nil {
			return nil, err
		}
		glog.Infof("Detected dice: %v", dice)
	}

	if d.annotate {
		roll := promptRoll()
		if err := d.saveAnnotation(id, roll); err != nil {
			glog.Error(err)
		}

		dice = roll
	}

	if len(dice) != yahtzee.NDice {
		return nil, fmt.Errorf("incorrect number of dice: got %d, expected %d",
			len(dice), yahtzee.NDice)
	}

	return dice, nil
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

func (d *YahtzeeDetector) saveImage(frame []byte) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	outputFile := path.Join(d.outDir, fmt.Sprintf("%s.jpg", id.String()))
	glog.V(1).Infof("Saving: %v", outputFile)
	f, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.Write(frame)
	return id.String(), err
}

func (d *YahtzeeDetector) saveAnnotation(id string, roll []int) error {
	rollsFile := path.Join(d.outDir, fmt.Sprintf("%s.json", id))
	f, err := os.OpenFile(rollsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(roll)
}

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
