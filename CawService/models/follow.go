package models

import "gopkg.in/mgo.v2/bson"

// Follow represents followed of follower user
type Follow struct {
	UserID bson.ObjectId `json:"user_id" bson:"user_id"`
	Name   string        `json:"name"`
}

// Equal returns true if provided user and current user has the same Name and UserId
// otherwhise false
func (u *Follow) Equal(other Follow) bool {
	return u.Name == other.Name && u.UserID.Hex() == other.UserID.Hex()
}
