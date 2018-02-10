package detector

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"gocv.io/x/gocv"

	"github.com/timpalpant/yahtzee"
)

const (
	imageWidth  = 1280
	imageHeight = 960
)

var (
	dhtMarker = []byte{255, 196}
	dht       = []byte{1, 162, 0, 0, 1, 5, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 1, 0, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 16, 0, 2, 1, 3, 3, 2, 4, 3, 5, 5, 4, 4, 0, 0, 1, 125, 1, 2, 3, 0, 4, 17, 5, 18, 33, 49, 65, 6, 19, 81, 97, 7, 34, 113, 20, 50, 129, 145, 161, 8, 35, 66, 177, 193, 21, 82, 209, 240, 36, 51, 98, 114, 130, 9, 10, 22, 23, 24, 25, 26, 37, 38, 39, 40, 41, 42, 52, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 17, 0, 2, 1, 2, 4, 4, 3, 4, 7, 5, 4, 4, 0, 1, 2, 119, 0, 1, 2, 3, 17, 4, 5, 33, 49, 6, 18, 65, 81, 7, 97, 113, 19, 34, 50, 129, 8, 20, 66, 145, 161, 177, 193, 9, 35, 51, 82, 240, 21, 98, 114, 209, 10, 22, 36, 52, 225, 37, 241, 23, 24, 25, 26, 38, 39, 40, 41, 42, 53, 54, 55, 56, 57, 58, 67, 68, 69, 70, 71, 72, 73, 74, 83, 84, 85, 86, 87, 88, 89, 90, 99, 100, 101, 102, 103, 104, 105, 106, 115, 116, 117, 118, 119, 120, 121, 122, 130, 131, 132, 133, 134, 135, 136, 137, 138, 146, 147, 148, 149, 150, 151, 152, 153, 154, 162, 163, 164, 165, 166, 167, 168, 169, 170, 178, 179, 180, 181, 182, 183, 184, 185, 186, 194, 195, 196, 197, 198, 199, 200, 201, 202, 210, 211, 212, 213, 214, 215, 216, 217, 218, 226, 227, 228, 229, 230, 231, 232, 233, 234, 242, 243, 244, 245, 246, 247, 248, 249, 250}
	sosMarker = []byte{255, 218}
)

type YahtzeeDetector struct {
	cam    *gocv.VideoCapture
	client *Client

	outDir string
	n      int
}

func NewYahtzeeDetector(device int, uri, outDir string) (*YahtzeeDetector, error) {
	cam, err := gocv.VideoCaptureDevice(device)
	if err != nil {
		return nil, err
	}

	cam.Grab(5)

	if outDir != "" {
		os.MkdirAll(outDir, os.ModePerm)
	}

	return &YahtzeeDetector{
		cam:    cam,
		client: NewClient(uri),
		outDir: outDir,
	}, nil
}

func (d *YahtzeeDetector) Close() error {
	return d.cam.Close()
}

func (d *YahtzeeDetector) GetCurrentRoll() ([]int, error) {
	img := gocv.NewMat()
	defer img.Close()

	if ok := d.cam.Read(img); !ok {
		return nil, fmt.Errorf("error reading image from webcam")
	}
	if img.Empty() {
		return nil, fmt.Errorf("received an empty image from webcam")
	}

	frame, err := gocv.IMEncode(gocv.JPEGFileExt, img)
	if err != nil {
		return nil, err
	}

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

func (d *YahtzeeDetector) saveTrainingData(img gocv.Mat, roll []int) error {
	outputFile := path.Join(d.outDir, fmt.Sprintf("%d.jpg", d.n))
	glog.V(1).Infof("Saving: %v", outputFile)
	if ok := gocv.IMWrite(outputFile, img); !ok {
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
