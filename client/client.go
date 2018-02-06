package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/timpalpant/yahtzee"
	"github.com/timpalpant/yahtzee/server"
)

type Client struct {
	client *http.Client
	uri    string
}

func NewClient(uri string) *Client {
	return &Client{http.DefaultClient, uri}
}

func (c *Client) GetOptimalMove(game yahtzee.GameState, step server.TurnStep, roll yahtzee.Roll, scoreToBeat int) (*server.OptimalMoveResponse, error) {
	req := &server.OptimalMoveRequest{
		GameState: server.FromYahtzeeGameState(game),
		TurnState: server.TurnState{
			Step: step,
			Dice: roll.Dice(),
		},
		ScoreToBeat: scoreToBeat,
	}

	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}

	endpoint := c.uri + "/rest/v1/optimal_move"
	resp, err := http.Post(endpoint, "application/json; charset=utf-8", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned: %v", resp.Status)
	}

	result := &server.OptimalMoveResponse{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(result); err != nil {
		return nil, err
	}

	return result, err
}
