package twitchclient

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func (c *Client) shouldRefresh() bool {
	//return c.UpdatedAt.Before(time.Now().Add(-1 * time.Hour * 24 * 30))
	return true
}

func (c *Client) RefreshTokens() error {
	type RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Scopes       []string
		TokenType    string `json:"token_type"`
	}

	l := c.logger()

	l.WithFields(logrus.Fields{
		"access_token":  c.AccessToken,
		"refresh_token": c.RefreshToken,
	}).Info("Refreshing tokens")

	TokenURL := "https://id.twitch.tv/oauth2/token"
	resp, err := c.HttpClient().PostForm(TokenURL, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.RefreshToken},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var tokenResp RefreshTokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.AccessToken = tokenResp.AccessToken
	c.RefreshToken = tokenResp.RefreshToken
	c.UpdatedAt = time.Now()

	l.WithFields(logrus.Fields{
		"access_token":  c.AccessToken,
		"refresh_token": c.RefreshToken,
	}).Info("Refreshed tokens")

	if err := c.Save(); err != nil {
		return err
	}

	return nil
}
