package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuth handles Google OAuth authentication
type GoogleOAuth struct {
	config *oauth2.Config
	states map[string]bool
	mu     sync.RWMutex
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewGoogleOAuth creates a new GoogleOAuth instance
func NewGoogleOAuth(clientID, clientSecret, redirectURL string) *GoogleOAuth {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuth{
		config: config,
		states: make(map[string]bool),
	}
}

// GetAuthURL returns the Google OAuth authorization URL
func (g *GoogleOAuth) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// SaveState saves a state for CSRF protection
func (g *GoogleOAuth) SaveState(state string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.states[state] = true
}

// ValidateState validates a state for CSRF protection
func (g *GoogleOAuth) ValidateState(state string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.states[state]; exists {
		delete(g.states, state) // Remove state after validation
		return true
	}
	return false
}

// Exchange exchanges authorization code for access token
func (g *GoogleOAuth) Exchange(code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo retrieves user information from Google
func (g *GoogleOAuth) GetUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := g.config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
