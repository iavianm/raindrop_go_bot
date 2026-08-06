package raindrop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const tokenURL = "https://raindrop.io/oauth/access_token"
const createBookmarkURL = "https://api.raindrop.io/rest/v1/raindrop"

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type authorizeRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

type bookmarkRequest struct {
	Link        string      `json:"link"`
	PleaseParse pleaseParse `json:"pleaseParse"`
}

type refreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type pleaseParse struct{}

func New(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Authorize(code string) (*Tokens, error) {
	body, err := json.Marshal(authorizeRequest{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		RedirectURI:  c.redirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("raindrop: marshal request: %w", err)
	}

	tokens, _ := c.requestTokens(body)

	return tokens, nil
}

func (c *Client) AddBookmark(accessToken, link string) error {
	body, err := json.Marshal(bookmarkRequest{
		Link:        link,
		PleaseParse: pleaseParse{},
	})
	if err != nil {
		return fmt.Errorf("raindrop: marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, createBookmarkURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("raindrop: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("raindrop: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("raindrop: unexpected status %d: %s", resp.StatusCode, errBody)
	}
	return nil
}

func (c *Client) Refresh(refreshToken string) (*Tokens, error) {
	body, err := json.Marshal(refreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("raindrop: marshal request: %w", err)
	}

	tokens, _ := c.requestTokens(body)

	return tokens, nil
}

func (c *Client) requestTokens(payload []byte) (*Tokens, error) {
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("raindrop: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("raindrop: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("raindrop: unexpected status %d: %s", resp.StatusCode, errBody)
	}

	var tokens Tokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("raindrop: decode response: %w", err)
	}

	if tokens.AccessToken == "" {
		return nil, fmt.Errorf("raindrop: empty access token in response")
	}

	return &tokens, nil
}
