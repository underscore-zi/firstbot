package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	// BroadcasterID is the Twitch ID of the broadcaster to watch for online/offline events
	BroadcasterID string `json:"broadcaster"`
	// Chat contains all the configuration for the Twitch Chat bot
	Chat     ChatConfig `json:"chat"`
	filename string
}

type ChatConfig struct {
	// Username is the Twitch IRC username to connect using
	Username string `json:"username"`
	// Token is the token that starts with "oauth:" to authenticate to Twitch IRC
	Token string `json:"token"`
	// Channel is the channel for the bot to listen and send messages in
	Channel string `json:"channel"`
}

func (c *Config) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	c.filename = filename
	return nil
}
