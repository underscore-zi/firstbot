package twitchclient

import "net/http"

// Do is a wrapper for the internal client's Do method. It adds the Authorization header to the request.
// and refreshes the tokens as necessary
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	l := c.logger()
	if c.shouldRefresh() {
		if err := c.RefreshTokens(); err != nil {
			l.WithError(err).Error("Failed to refresh tokens, will try again next request")
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	return c.HttpClient().Do(req)
}
