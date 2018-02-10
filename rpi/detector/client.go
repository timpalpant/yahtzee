package detector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	client *http.Client
	uri    string
}

func NewClient(uri string) *Client {
	return &Client{http.DefaultClient, uri}
}

func (c *Client) GetDiceFromImage(image []byte) ([]int, error) {
	var req struct {
		Image string `json:"image"`
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
		Error struct {
			Message string
		}
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	if result.Error.Message != "" {
		err = fmt.Errorf(result.Error.Message)
	}

	return result.Dice, err
}
