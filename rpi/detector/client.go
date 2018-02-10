package detector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type ImageProcessingError struct {
	Message string
}

func (e ImageProcessingError) Error() string {
	return e.Message
}

type Client struct {
	client *http.Client
	uri    string
}

func NewClient(uri string) *Client {
	return &Client{http.DefaultClient, uri}
}

func (c *Client) GetDiceFromImage(image []byte) ([]int, error) {
	var req struct {
		Image string
	}

	req.Image = base64.StdEncoding.EncodeToString(image)
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}

	endpoint := c.uri + "/rest/v1/process_image"
	resp, err := http.Post(endpoint, "application/json; charset=utf-8", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned: %v", resp.Status)
	}

	var result struct {
		Dice  []int
		Error *ImageProcessingError
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return result.Dice, result.Error
}
