package utils

import "testing"

func TestUriBuilder(t *testing.T) {
	uriBuilder := NewUriBuilder()
	userID := "123"
	cawID := "456"
	followingID := "789"
	followersID := "987"
	var testCases = []struct {
		Name        string
		URI         string
		ExpectedURI string
	}{
		{
			Name:        "UserURITest",
			URI:         uriBuilder.User().Done(),
			ExpectedURI: "/v1/users",
		},
		{
			Name:        "UserURIWithUserIDTest",
			URI:         uriBuilder.User().WithUser(userID).Done(),
			ExpectedURI: "/v1/users/" + userID,
		},
		{
			Name:        "UserURIWithUserIDAndCawIDTest",
			URI:         uriBuilder.User().WithUser(userID).Caws().WithCaw(cawID).Done(),
			ExpectedURI: "/v1/users/" + userID + "/caws/" + cawID,
		},
		{
			Name:        "UserURIWithUserIDAndFollowingIDTest",
			URI:         uriBuilder.User().WithUser(userID).Following().WithFollowing(followingID).Done(),
			ExpectedURI: "/v1/users/" + userID + "/following/" + followingID,
		},
		{
			Name:        "UserURIWithUserIDAndFollowersIDTest",
			URI:         uriBuilder.User().WithUser(userID).Followers().WithFollowers(followersID).Done(),
			ExpectedURI: "/v1/users/" + userID + "/followers/" + followersID,
		},
		{
			Name:        "UserURIWithUserIDAndFollowingTest",
			URI:         uriBuilder.User().WithUser(userID).Following().Done(),
			ExpectedURI: "/v1/users/" + userID + "/following",
		},
		{
			Name:        "UserURIWithUserIDFollowersTest",
			URI:         uriBuilder.User().WithUser(userID).Followers().Done(),
			ExpectedURI: "/v1/users/" + userID + "/followers",
		},
		{
			Name:        "AuthUriTest",
			URI:         uriBuilder.Auth().Done(),
			ExpectedURI: "/v1/authentication",
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		if testCase.URI != testCase.ExpectedURI {
			t.Errorf("Given URI %v is not equal to expected URI %v", testCase.URI, testCase.ExpectedURI)
		}
	}
}
