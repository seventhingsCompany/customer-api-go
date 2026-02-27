package models

// SSOProviderName identifies an SSO identity provider.
type SSOProviderName string

const (
	// SSOProviderAzure is the Azure OpenID Connect identity provider.
	SSOProviderAzure SSOProviderName = "azure-open-id-connect"
	// SSOProviderGoogle is the Google OpenID Connect identity provider.
	SSOProviderGoogle SSOProviderName = "google-open-id-connect"
	// SSOProviderOneLogin is the OneLogin OpenID Connect identity provider.
	SSOProviderOneLogin SSOProviderName = "one-login-open-id-connect"
)

// SSOAppTarget identifies the application target for SSO login.
type SSOAppTarget string

const (
	// SSOAppTargetWeb targets the web application.
	SSOAppTargetWeb SSOAppTarget = "web"
	// SSOAppTargetMobile targets the mobile application.
	SSOAppTargetMobile SSOAppTarget = "mobile"
)

// LoginDeniedReason describes why a login attempt was denied.
type LoginDeniedReason string

const (
	// LoginDeniedDeactivated indicates the account has been deactivated.
	LoginDeniedDeactivated LoginDeniedReason = "LoginDeactivated"
	// LoginDeniedBanned indicates the account has been banned.
	LoginDeniedBanned LoginDeniedReason = "Banned"
	// LoginDeniedEmailUnconfirmed indicates the email has not been confirmed.
	LoginDeniedEmailUnconfirmed LoginDeniedReason = "EmailUnconfirmed"
	// LoginDeniedInactive indicates the account is inactive.
	LoginDeniedInactive LoginDeniedReason = "Inactive"
	// LoginDeniedOnlySSOAllowed indicates only SSO login is permitted.
	LoginDeniedOnlySSOAllowed LoginDeniedReason = "OnlySSOLoginAllowed"
)

// TokenResponse is the response from a successful authentication request.
type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    int     `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	Scope        *string `json:"scope"`
	RefreshToken string  `json:"refresh_token"`
	UserID       int     `json:"user_id"`
}

// LoginDeniedDetail is the response body when login is denied (403).
type LoginDeniedDetail struct {
	Detail LoginDeniedReason `json:"detail"`
}
