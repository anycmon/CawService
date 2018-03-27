package models

// Auth represents credentials required to authorize
type Auth struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
