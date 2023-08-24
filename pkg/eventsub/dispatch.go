package eventsub

import (
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func (c *Client) watchdogTriggered() {
	c.logger.Info("watchdog closed, restarting")
	os.Exit(0)
}

func (c *Client) watchdog() {
	defer c.watchdogTriggered()

	for {
		select {
		case <-c.keepaliveTicker.C:
			c.logger.Info("Missed keepalive, closing connection")
			_ = c.Close()
			return
		}
	}
}

func (c *Client) welcome(msg WebsocketMessage) {
	c.session = msg.Payload.Session
	c.logger.WithField("session", c.session.ID).Info("Connected to twitch")

	subs, _ := c.ListSubscriptions()
	for _, sub := range subs {
		if sub.Status != "enabled" {
			continue
		}

		c.logger.WithFields(logrus.Fields{
			"subscription": sub.ID,
			"type":         sub.Type,
			"status":       sub.Status,
		}).Info("Reconnected to subscription")
	}
}

func (c *Client) reconnect(msg WebsocketMessage) {
	oldURL := c.SocketURL
	oldconn := c.conn

	c.SocketURL = msg.Payload.Session.ReconnectURL
	c.logger.WithField("url", c.SocketURL).Info("Reconnecting to twitch")
	_, _, err := c.Connect()
	if err != nil {
		c.logger.WithError(err).Error("Failed to reconnect")
	}

	// Waiting for session to be established
	var count int
	for c.session != nil {
		time.Sleep(1 * time.Second)
		count++
		if count > 10 {
			c.logger.WithField("url", c.SocketURL).Error("Failed to reconnect")
			return
		}
	}

	c.SocketURL = oldURL
	_ = oldconn.Close()
}

func (c *Client) dispatch(msg WebsocketMessage) {
	switch msg.Metadata.Type {
	case "session_welcome":
		c.welcome(msg)
	case "session_reconnect":
		c.reconnect(msg)
	case "revocation":
		if handler, found := c.subscriptions[msg.Payload.Subscription.ID]; found {
			delete(c.subscriptions, msg.Payload.Subscription.ID)
			if msg.Payload.Event != nil {
				handler.OnRevoke(*msg.Payload.Subscription)
			} else {
				c.logger.WithField("subscription", msg.Payload.Subscription.ID).Error("Received event with no payload, this shouldn't happen")
			}
		}

	case "notification":
		if handler, found := c.subscriptions[msg.Payload.Subscription.ID]; found {
			if msg.Payload.Event != nil {
				handler.OnEvent(*msg.Payload.Subscription, *msg.Payload.Event)
			} else {
				c.logger.WithField("subscription", msg.Payload.Subscription.ID).Error("Received event with no payload, this shouldn't happen")
			}
		} else {
			c.logger.WithField("subscription", msg.Payload.Subscription.ID).Error("Received event for unknown subscription")
		}
	}

	// Every message should reset the timer
	if c.session != nil && c.session.KeepaliveTimeoutSeconds > 0 {
		c.keepaliveTicker.Reset(time.Duration(c.session.KeepaliveTimeoutSeconds) * time.Second * 2)
	}
}
