package neo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	SessionState     string `json:"session_state"`
}

func (c *Client) auth(ctx context.Context, body url.Values) (*Token, error) {
	if len(body) == 0 {
		return nil, ErrInvalidAuthRequest
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/auth/realms/sandbox/protocol/openid-connect/token",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("neo: failed to create a new http.Request: %w", err)
	}
	req.Header.Set("content-type", ContentTypeFormURLEncoded)
	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("neo: invalid auth request: %w", err)
	}
	defer closeBody(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("neo: unexpected HTTP response: %s", resp.Status)
	}
	t := &Token{}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return nil, fmt.Errorf("neo: unable to decode an access token: %w", err)
	}
	return t, nil
}

// AccessToken returns an access token and a refresh token from the authentication server.
func (c *Client) AccessToken(ctx context.Context) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", c.client)
	body.Set("client_secret", c.secret)
	body.Set("grant_type", "client_credentials")

	return c.auth(ctx, body)
}

// RefreshToken returns a fresh access token.
func (c *Client) RefreshToken(ctx context.Context, t *Token) (*Token, error) {
	if t == nil || t.RefreshToken == "" {
		return nil, fmt.Errorf("auth: invalid refresh token")
	}

	body := url.Values{}
	body.Set("client_id", c.client)
	body.Set("client_secret", c.secret)
	body.Set("refresh_token", t.RefreshToken)
	body.Set("grant_type", "refresh_token")

	return c.auth(ctx, body)
}
