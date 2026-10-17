package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

type loginRequest struct {
	GrantType string `json:"grant_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	ClientID  string `json:"client_id"`
}

type refreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
}

type ssoLoginRequest struct {
	GrantType    string                 `json:"grant_type"`
	ProviderName models.SSOProviderName `json:"provider_name"`
	AuthCode     string                 `json:"auth_code"`
	ClientID     string                 `json:"client_id"`
	AppTarget    *models.SSOAppTarget   `json:"app_target,omitempty"`
}

const authTokenPath = "auth_token"

// postAuthToken marshals payload as JSON, sends an unauthenticated POST to
// the auth_token endpoint, decodes the response, and stores the access token.
func (c *Client) postAuthToken(ctx context.Context, payload any) (*models.TokenResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoUnauthenticated(ctx, http.MethodPost, authTokenPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var tok models.TokenResponse
	if err := DecodeJSON(resp, &tok); err != nil {
		return nil, err
	}

	c.token = tok.AccessToken
	return &tok, nil
}

// Login authenticates with username and password credentials.
func (c *Client) Login(ctx context.Context, username, password, clientID string) (*models.TokenResponse, error) {
	c.clientID = clientID
	return c.postAuthToken(ctx, loginRequest{
		GrantType: "password",
		Username:  username,
		Password:  password,
		ClientID:  clientID,
	})
}

// Refresh exchanges a refresh token for a new access token using the stored
// client ID.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	return c.postAuthToken(ctx, refreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     c.clientID,
	})
}

// LoginSSO authenticates using an SSO authorization code. appTarget may be nil
// to omit the field from the request.
func (c *Client) LoginSSO(ctx context.Context, provider models.SSOProviderName, authCode, clientID string, appTarget *models.SSOAppTarget) (*models.TokenResponse, error) {
	c.clientID = clientID
	return c.postAuthToken(ctx, ssoLoginRequest{
		GrantType:    "sso_auth_code",
		ProviderName: provider,
		AuthCode:     authCode,
		ClientID:     clientID,
		AppTarget:    appTarget,
	})
}

// RevokeTokens revokes the current authentication tokens.
func (c *Client) RevokeTokens(ctx context.Context) error {
	_, err := c.Delete(ctx, authTokenPath)
	return err
}
