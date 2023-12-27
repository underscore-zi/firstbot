package eventsub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type SubscriptionHandler interface {
	OnSubscribed(sub Subscription)
	OnEvent(sub Subscription, event Event)
	OnRevoke(sub Subscription)
}

type SubscriptionRequest struct {
	Type      string                 `json:"type"`
	Version   string                 `json:"version"`
	Condition map[string]interface{} `json:"condition"`
	Transport struct {
		Method    string `json:"method"`
		SessionID string `json:"session_id"`
	} `json:"transport"`
}

type SubscriptionResponse struct {
	Data         []Subscription `json:"data"`
	Total        int            `json:"total"`
	TotalCost    int            `json:"total_cost"`
	MaxTotalCost int            `json:"max_total_cost"`
}

func (c *Client) Subscribe(subscriptionType, version string, condition map[string]interface{}, handler SubscriptionHandler) error {
	for c.session == nil {
		c.logger.Info("Waiting for session to be created...")
		time.Sleep(1 * time.Second)
	}

	request := SubscriptionRequest{
		Type:      subscriptionType,
		Version:   version,
		Condition: condition,
	}
	request.Transport.Method = "websocket"
	request.Transport.SessionID = c.session.ID

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(request); err != nil {
		c.logger.WithError(err).Error("Failed to encode subscription request")
	}

	req, err := http.NewRequest("POST", c.SubscriptionURL, &buf)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create subscription request")
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Client-ID", c.TwitchClientID)
	req.Header.Add("Authorization", "Bearer "+c.TwitchAccessToken)

	hClient := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       10 * time.Second,
	}

	resp, err := hClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Failed to send subscription request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 202 {
		c.logger.WithField("status", resp.StatusCode).Error("Failed to subscribe")
		return err
	}

	var subscriptionResponse SubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&subscriptionResponse); err != nil {
		c.logger.WithError(err).Error("Failed to parse subscription response")
		return err
	}

	c.subscriptions[subscriptionResponse.Data[0].ID] = handler
	handler.OnSubscribed(subscriptionResponse.Data[0])

	return nil
}
