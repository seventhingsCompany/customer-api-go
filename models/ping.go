package models

// PingResponse represents the response from the ping endpoint.
type PingResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}
