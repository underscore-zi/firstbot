package main

import (
	"FirstBot/pkg/eventsub"
	"FirstBot/pkg/eventsub/subscriptions"
	"FirstBot/pkg/twitchclient"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

var logger = logrus.StandardLogger()

func loadFiles(twitchFile, configFile, stateFile string) (state *State, twitch *twitchclient.Client, config *Config) {
	var err error

	state = &State{Filename: stateFile}
	err = state.Load()
	if err != nil {
		logger.WithError(err).WithField("filename", stateFile).Fatal("Failed to load state")
	}

	twitch, err = twitchclient.New(nil, nil, twitchFile)
	if err != nil {
		logger.WithError(err).WithField("filename", twitchFile).Fatal("Failed to load Twitch config")
	}

	config = &Config{}
	err = config.Load(configFile)
	if err != nil {
		logger.WithError(err).WithField("filename", configFile).Fatal("Failed to load config")
	}

	if config.Chat.Username == "" || config.Chat.Channel == "" {
		logger.Fatal("Chat username and channel must be set")
	}

	if config.Chat.Token == "" {
		logger.Info("Chat token not set, will use the Twitch access token. This only works if the eventsub account is the same as the twitch chat account")
		config.Chat.Token = fmt.Sprintf("oauth:%s", twitch.AccessToken)
	}

	return
}

func main() {
	var twitchFile = flag.String("twitch", "twitch.json", "The file to store/load the Twitch app/token information from")
	var configFile = flag.String("config", "config.json", "The file load general configuration information from")
	var stateFile = flag.String("state", "state.json", "Contains the bot state like claim history and such")
	flag.Parse()
	state, twitch, config := loadFiles(*twitchFile, *configFile, *stateFile)

	streamState := make(chan bool)

	if _, _, err := twitch.EventSubClient().Connect(); err != nil {
		logger.WithError(err).Fatal("Failed to connect to Twitch EventSub")
	}

	if err := (subscriptions.StreamOnline{
		EventSub:  twitch.EventSubClient(),
		Logger:    logger,
		ChannelID: config.BroadcasterID,
		Callback: func(_ eventsub.Event) {
			streamState <- true
		},
	}).Register(); err != nil {
		logger.WithError(err).Fatal("Failed to subscribe to stream.online")
	}

	if err := (subscriptions.StreamOffline{
		EventSub:  twitch.EventSubClient(),
		Logger:    logger,
		ChannelID: config.BroadcasterID,
		Callback: func(_ eventsub.Event) {
			streamState <- false
		},
	}).Register(); err != nil {
		logger.WithError(err).Fatal("Failed to subscribe to stream.offline")
	}

	chatRecv, chatSend := TwitchChat(config.Chat.Username, config.Chat.Token, config.Chat.Channel)
	logger.WithFields(logrus.Fields{
		"username": config.Chat.Username,
		"channel":  config.Chat.Channel,
	}).Info("Connected to Twitch chat")

	for {
		select {
		case isLive := <-streamState:
			if isLive {
				logger.Info("Stream is live")
				state.SetOnline()
			} else {
				logger.Info("Stream is offline")
				state.SetOffline()
			}
		case msg := <-chatRecv:
			firstWord := strings.ToLower(strings.Split(msg.Message, " ")[0])
			if firstWord == "!first" {
				switch state.TryClaim(msg.User.Name) {
				case nil:
					claimedBy, total, streak := state.ClaimedBy()
					logger.WithFields(logrus.Fields{
						"claimed_by": claimedBy,
						"total":      total,
						"streak":     streak,
					}).Info("First claimed")

					out := strings.Builder{}
					out.WriteString(fmt.Sprintf("Congratulations @%s! You are the first! ", claimedBy))

					if total == 1 {
						out.WriteString(fmt.Sprintf("This is your first first! "))
					} else {
						out.WriteString(fmt.Sprintf("You've been first %d times, ", total))
						if streak == 1 {
							out.WriteString(fmt.Sprintf("and this is the start of a new streak! "))
						} else {
							var suffix string
							switch streak % 10 {
							case 1:
								suffix = "st"
							case 2:
								suffix = "nd"
							case 3:
								suffix = "rd"
							default:
								suffix = "th"
							}

							out.WriteString(fmt.Sprintf("and this is the %d%s time in a row.", streak, suffix))
						}
					}

					chatSend <- out.String()
				case ErrAlreadyClaimed:
					claimedBy, _, _ := state.ClaimedBy()
					if msg.User.Name == claimedBy {
						chatSend <- fmt.Sprintf("@%s, you already claimed first!", msg.User.Name)
					} else {
						chatSend <- fmt.Sprintf("Sorry, @%s, too late! @%s already beat you to it.", msg.User.Name, claimedBy)
					}
				case ErrNotLive:
					chatSend <- fmt.Sprintf("Nice try @%s but the stream is not live yet!", msg.User.Name)
				}
			}

			if msg.User.Name == "underscorezi" {
				if firstWord == "!forceoffline" {
					go func() { streamState <- false }()
				} else if firstWord == "!forceonline" {
					go func() { streamState <- true }()
				}
			}

		}
	}
}
