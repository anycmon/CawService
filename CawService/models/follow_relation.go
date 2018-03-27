package models

// FollowRelation represents users relationship
type FollowRelation struct {
	Follower  Follow `json:"follower"`
	Following Follow `json:"following"`
}
