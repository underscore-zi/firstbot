package eventsub

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
	"time"
)

var SocketURL = "wss://eventsub.wss.twitch.tv/ws"
var SubscriptionURL = "https://api.twitch.tv/helix/eventsub/subscriptions"

func init() {
	if v, found := os.LookupEnv("EVENTSUB_SOCKET_URL"); found {
		SocketURL = v
	}

	if v, found := os.LookupEnv("EVENTSUB_SUBSCRIPTION_URL"); found {
		SubscriptionURL = v
	}
}

type Client struct {
	Config
	conn            *websocket.Conn
	logger          *logrus.Logger
	session         *Session
	keepaliveTicker *time.Ticker
	subscriptions   map[string]SubscriptionHandler
	once            sync.Once
}

type Config struct {
	TwitchClientID    string
	TwitchAccessToken string
	SocketURL         string
	SubcriptionURL    string
}

func (c *Client) Connect() (*websocket.Conn, *http.Response, error) {
	c.session = nil
	conn, resp, err := websocket.DefaultDialer.Dial(c.SocketURL, nil)
	if err == nil {
		c.conn = conn
	}

	c.once.Do(func() { go c.watchdog() })
	go c.reader()

	return conn, resp, err
}

func (c *Client) reader() {
	for {
		if c.conn == nil {
			return
		}

		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var parsed WebsocketMessage
		switch messageType {
		case websocket.TextMessage:
			if err = json.Unmarshal(message, &parsed); err != nil {
				c.logger.WithError(err).Error("Failed to parse message from twitch")
				continue
			}
			go c.dispatch(parsed)
		default:
			c.logger.WithField("type", messageType).Error("Unhandled message type")
		}
	}
}

func (c *Client) Close() error {
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) ListSubscriptions() ([]Subscription, error) {
	hClient := http.Client{}
	req, err := http.NewRequest("GET", c.SubcriptionURL, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create subscription list request")
		return nil, err
	}

	req.Header.Add("Client-ID", c.TwitchClientID)
	req.Header.Add("Authorization", "Bearer "+c.TwitchAccessToken)

	type responseStruct struct {
		Data []Subscription `json:"data"`
	}
	var data responseStruct

	resp, err := hClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get subscription list")
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	// read resp.Body into data
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	if err != nil {
		c.logger.WithError(err).Error("Failed to decode subscription list")
		return nil, err
	}
	return data.Data, nil
}

func NewClient(config Config) *Client {
	c := &Client{
		Config:          config,
		logger:          logrus.StandardLogger(),
		keepaliveTicker: time.NewTicker(10 * time.Second),
		subscriptions:   make(map[string]SubscriptionHandler),
	}

	// Twich has been sending a keep alive value of 0 so, we stop the ticker by default
	// and if we ever recieve a value > 0 it'll start back up.
	c.keepaliveTicker.Stop()

	return c
}
