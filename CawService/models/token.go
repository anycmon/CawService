package models

import "time"

// Token represents session
type Token struct {
	AccessToken string    `json:"access_token"`
	Type        string    `json:"type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
