package utils

import "bytes"

type UriBuilder interface {
	User() UriBuilder
	Auth() UriBuilder
	Following() UriBuilder
	Followers() UriBuilder
	Caws() UriBuilder
	WithUser(userID string) UriBuilder
	WithFollowing(followingID string) UriBuilder
	WithFollowers(followersID string) UriBuilder
	WithCaw(cawID string) UriBuilder
	Done() string
}

type uriBuilder struct {
	buffer bytes.Buffer
}

func (ub uriBuilder) User() UriBuilder {
	ub.buffer.WriteString("/v1/users")
	return ub
}

func (ub uriBuilder) WithUser(userID string) UriBuilder {
	ub.buffer.WriteString("/" + userID)
	return ub
}

func (ub uriBuilder) Auth() UriBuilder {
	ub.buffer.WriteString("/v1/authentication")
	return ub
}

func (ub uriBuilder) Following() UriBuilder {
	ub.buffer.WriteString("/following")
	return ub
}

func (ub uriBuilder) WithFollowing(followingID string) UriBuilder {
	ub.buffer.WriteString("/" + followingID)
	return ub
}

func (ub uriBuilder) Followers() UriBuilder {
	ub.buffer.WriteString("/followers")
	return ub
}

func (ub uriBuilder) WithFollowers(followersID string) UriBuilder {
	ub.buffer.WriteString("/" + followersID)
	return ub
}

func (ub uriBuilder) Caws() UriBuilder {
	ub.buffer.WriteString("/caws")
	return ub
}

func (ub uriBuilder) WithCaw(cawID string) UriBuilder {
	ub.buffer.WriteString("/" + cawID)
	return ub
}

func (ub uriBuilder) Done() string {
	result := ub.buffer.String()
	ub.buffer.Reset()
	return result
}

func NewUriBuilder() UriBuilder {
	return uriBuilder{}
}
