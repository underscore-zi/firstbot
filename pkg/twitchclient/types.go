package twitchclient

import (
	"github.com/sirupsen/logrus"
	"github.com/underscore-zi/firstbot/pkg/eventsub"
	"net/http"
	"time"
)

type Application struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AccessTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Client struct {
	Application  `json:"application"`
	AccessTokens `json:"access_tokens"`

	httpClient     *http.Client
	filename       string
	log            *logrus.Logger
	eventsubClient *eventsub.Client
}

func (c *Client) HttpClient() *http.Client {
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return c.httpClient
}

func (c *Client) logger() *logrus.Logger {
	if c.log == nil {
		c.log = logrus.StandardLogger()
	}
	return c.log
}

func New(client *http.Client, logger *logrus.Logger, filename string) (*Client, error) {
	c := &Client{
		httpClient: client,
		log:        logger,
	}
	err := c.Load(filename, true)
	return c, err
}
