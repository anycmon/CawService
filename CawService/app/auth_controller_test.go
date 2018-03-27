package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2/bson"
)

func TestAuthenticationHandler(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		Name               string
		AuthData           models.Auth
		User               models.User
		UserDataStoreError error
		ExpectedStatusCode int
	}{
		{
			Name:     "SuccessfullyAuthenticatedUserTest",
			AuthData: models.Auth{Name: "anycmon", Password: "foobar"},
			User: models.User{
				ID:       bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
				Name:     "anycmon",
				Email:    "anycmon@gmail.com",
				Password: "$2a$14$KIRkInExkvkQdy71k1hkcOh9WtOqqQYNCFsKsr7hmTLPJko2LTXU6"},
			UserDataStoreError: nil,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:     "WrongPasswordTest",
			AuthData: models.Auth{Name: "anycmon", Password: "wrongPassword"},
			User: models.User{
				ID:       bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
				Name:     "anycmon",
				Email:    "anycmon@gmail.com",
				Password: "$2a$14$KIRkInExkvkQdy71k1hkcOh9WtOqqQYNCFsKsr7hmTLPJko2LTXU6"},
			UserDataStoreError: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:     "UserDoesNotExistsTest",
			AuthData: models.Auth{Name: "anycmon", Password: "foobar"},
			User: models.User{
				ID:       bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
				Name:     "anycmon",
				Email:    "anycmon@gmail.com",
				Password: "$2a$14$KIRkInExkvkQdy71k1hkcOh9WtOqqQYNCFsKsr7hmTLPJko2LTXU6"},
			UserDataStoreError: infrastructure.ErrNotFound,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:     "StorageInternalErrorTest",
			AuthData: models.Auth{Name: "anycmon", Password: "foobar"},
			User: models.User{
				ID:       bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
				Name:     "anycmon",
				Email:    "anycmon@gmail.com",
				Password: "$2a$14$KIRkInExkvkQdy71k1hkcOh9WtOqqQYNCFsKsr7hmTLPJko2LTXU6"},
			UserDataStoreError: errors.New("StorageInternalError"),
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		js, err := json.Marshal(testCase.AuthData)
		if err != nil {
			t.Fatal("Cannot marshal follow", err)
		}

		request, err := http.NewRequest("POST", uriBuilder.Auth().Done(), bytes.NewBuffer(js))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")

		userDataStoreMock := &UserDataStoreMock{
			OnGetUserByName: func(userName string) (*models.User, error) {
				return &testCase.User, testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)

		assert.Equal(t, testCase.ExpectedStatusCode, recorder.Code)

		if testCase.ExpectedStatusCode == http.StatusOK {
			var token models.Token
			err = json.NewDecoder(recorder.Body).Decode(&token)
			assert.Nil(t, err,
				"cannot to decode token payload body %v", err)

			claims, err := utils.DecodeUserToken(token.AccessToken)
			assert.Nil(t, err,
				"cannot to decode access token %v", err)

			err = claims.Valid()
			assert.Nil(t, err,
				"claims are not valid %v, %v", claims, err)

			assert.Equal(t, claims.UserId, testCase.User.ID.Hex(),
				"user id from claims and test case are not equal %v %v", testCase.User.ID.Hex(), claims.UserId)

			assert.Equal(t, token.ExpiresAt.Unix(), claims.ExpiresAt,
				"expireAt in token and claim are differen, %v, %v", token.ExpiresAt.Unix(), claims.ExpiresAt)
		}
	}
}

func TestAuthenticationOptionsHandler(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest("OPTIONS", uriBuilder.Auth().Done(), nil)
	if err != nil {
		t.Fatal(err)
	}
	userDataStoreMock := &UserDataStoreMock{}
	cawDataStoreMock := &CawDataStoreMock{}

	app := createApp(userDataStoreMock, cawDataStoreMock)

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, request)

	expectedContentType := "application/json"
	expectedAllowOrigin := "*"
	expectedAllowHeader := "*"

	contentType := recorder.Header().Get("Content-Type")
	allowOrigin := recorder.Header().Get("Access-Control-Allow-Origin")
	allowHeaders := recorder.Header().Get("Access-Control-Allow-Headers")

	assert.Equal(t, expectedContentType, contentType, "different content type %v %v", expectedContentType, contentType)
	assert.Equal(t, expectedAllowHeader, allowHeaders, "different allow headers %v %v", expectedAllowHeader, allowHeaders)
	assert.Equal(t, expectedAllowOrigin, allowOrigin, "different allow origin %v %v", expectedAllowOrigin, allowOrigin)

}
