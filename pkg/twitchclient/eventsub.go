package twitchclient

import "github.com/underscore-zi/firstbot/pkg/eventsub"

// EventSubClient provides an eventsub.Client but you'll still need to call .Connect on it to start it
// after the first call it always returns the same client instance
func (c *Client) EventSubClient() *eventsub.Client {
	if c.eventsubClient == nil {
		c.eventsubClient = eventsub.NewClient(eventsub.Config{
			TwitchClientID:    c.ClientID,
			TwitchAccessToken: c.AccessToken,
			SocketURL:         eventsub.SocketURL,
			SubcriptionURL:    eventsub.SubscriptionURL,
		})
	}
	return c.eventsubClient
}
