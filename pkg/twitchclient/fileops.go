package twitchclient

import (
	"encoding/json"
	"os"
)

func (c *Client) Load(filename string, shouldRefresh bool) error {
	c.filename = filename
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	if shouldRefresh && c.shouldRefresh() {
		if err = c.RefreshTokens(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Save() error {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(c.filename, jsonData, 0644)
}
