package models

import (
	"encoding/json"
	"io"
	"time"

	valid "github.com/asaskevich/govalidator"
	"gopkg.in/mgo.v2/bson"
)

// PublicUser represents User without sensitive data
type PublicUser struct {
	User
	Password string `json:"password,omitempty"`
}

// ToJSON converts current PublicUser to JSON payload
func (u PublicUser) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// User represents caw user
type User struct {
	ID             bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name           string        `json:"name"`
	Email          string        `json:"email"`
	Password       string        `json:"password"`
	FollowersCount uint64        `json:"followers_count" bson:"followers_count"`
	FollowingCount uint64        `json:"following_count" bson:"following_count"`
	CreatedAt      time.Time     `json:"created_at" bson:"created_at"`
}

// Equal return true if UserId, Name and Email are equal otherwhise false
func (u *User) Equal(other User) bool {
	return u.ID.Hex() == other.ID.Hex() &&
		u.Name == other.Name &&
		u.Email == other.Email
}

// ToPublic converts current user to PublicUser
func (u User) ToPublic() PublicUser {
	return PublicUser{User: u}
}

// IsValid returns false if name, email, password is empty or email is in wrong format
func (u User) IsValid() bool {
	if u.Name == "" || u.Email == "" || u.Password == "" || !valid.IsEmail(u.Email) {
		return false
	}

	return true
}

// UserFromJson parse JSON payload to User model
func UserFromJSON(jUser io.Reader) (*User, error) {
	var user User
	decoder := json.NewDecoder(jUser)
	if err := decoder.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
