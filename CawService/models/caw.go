package models

import (
	"encoding/json"
	"io"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Caw represents message or response to message created by user
type Caw struct {
	ID           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	UserID       bson.ObjectId `json:"user_id" bson:"user_id"`
	UserName     string        `json:"user_name" bson:"user_name"`
	ParentID     bson.ObjectId `json:"parent_id" bson:"parent_id,omitempty"`
	Message      string        `json:"message" bson:"message"`
	CreatedAt    time.Time     `json:"created_at" bson:"created_at"`
	LikeCount    int           `json:"like_count" bson:"like_count"`
	RecawCount   int           `json:"recaw_count" bson:"recaw_count"`
	RepliesCount int           `json:"replies_count" bson:"replies_count"`
}

// CawFromJson parse JSON payload to Caw model
func CawFromJSON(jCaw io.Reader) (*Caw, error) {
	var caw Caw
	decoder := json.NewDecoder(jCaw)
	if err := decoder.Decode(&caw); err != nil {
		return nil, err
	}

	return &caw, nil
}

// ToJSON converts current Caw to JSON payload
func (c Caw) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}
