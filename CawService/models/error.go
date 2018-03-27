package models

import "encoding/json"

type Error struct {
	Message string `json:"message"`
}

// ToJSON converts current Error to JSON payload
func (e Error) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
