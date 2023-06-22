package main

import (
	twitch "github.com/gempir/go-twitch-irc/v4"
	"strings"
)

// TwitchChat connects to the twitch chat and returns two channels, one for receiving messages and one for sending messages.
// if it is unable to connect to chat, it will panic
func TwitchChat(username, token, channel string) (chan twitch.PrivateMessage, chan string) {
	channel = strings.ToLower(channel)
	client := twitch.NewClient(username, token)
	client.Join(channel)

	messages := make(chan twitch.PrivateMessage)
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if strings.ToLower(message.Channel) != channel {
			return
		}
		messages <- message
	})

	out := make(chan string)
	go func() {
		for {
			msg := <-out
			client.Say(channel, msg)
		}
	}()

	go func() {
		err := client.Connect()
		if err != nil {
			panic("Unable to connect to twitch: " + err.Error())
		}
	}()

	return messages, out
}
