package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"

	"gopkg.in/mgo.v2/bson"
)

var Logger = logrus.New()
var uriBuilder = utils.NewUriBuilder()

type CawDataStoreMock struct {
	OnStore func(caw models.Caw) (*models.Caw, error)
}

func (m CawDataStoreMock) Store(caw models.Caw) (*models.Caw, error) {
	return m.OnStore(caw)
}

func (m CawDataStoreMock) GetByUserID(userID string, page int) ([]models.Caw, error) {
	return nil, nil
}

func (m CawDataStoreMock) GetByID(cawID string) (*models.Caw, error) {
	return nil, nil
}

func (m CawDataStoreMock) Delete(cawID string) error {
	return nil
}

func (m CawDataStoreMock) Close() {
}

type UserDataStoreMock struct {
	OnGetUser          func(userID string) (*models.User, error)
	OnGetUserByName    func(userID string) (*models.User, error)
	OnStoreUser        func(user models.User) (*models.User, error)
	OnDeleteUser       func(userID string) error
	OnGetUserFollowers func(userID string) ([]models.Follow, error)
	OnGetUserFollowing func(userID string) ([]models.Follow, error)
	OnAddFollowingUser func(followerID, followingID string) error
	OnUnfollowUser     func(followerID, followingID string) error
}

func (m UserDataStoreMock) GetUser(userID string) (*models.User, error) {
	return m.OnGetUser(userID)
}

func (m UserDataStoreMock) GetUserByName(userID string) (*models.User, error) {
	return m.OnGetUserByName(userID)
}

func (m UserDataStoreMock) StoreUser(user models.User) (*models.User, error) {
	return m.OnStoreUser(user)
}

func (m UserDataStoreMock) DeleteUser(userID string) error {
	return m.OnDeleteUser(userID)
}

func (m UserDataStoreMock) GetUserFollowers(userID string) ([]models.Follow, error) {
	return m.OnGetUserFollowers(userID)
}

func (m UserDataStoreMock) GetUserFollowing(userID string) ([]models.Follow, error) {
	return m.OnGetUserFollowing(userID)
}

func (m UserDataStoreMock) AddFollowingUser(followerID string, followingID string) error {
	return m.OnAddFollowingUser(followerID, followingID)
}

func (m UserDataStoreMock) UnfollowUser(followerID, followingID string) error {
	return m.OnUnfollowUser(followerID, followingID)
}

func (m UserDataStoreMock) Close() {
}

type DataStoreFactoryMock struct {
	OnCreateUserDataStore func() infrastructure.UserDataStore
	OnCreateCawDataStore  func() infrastructure.CawDataStore
}

func (m DataStoreFactoryMock) CreateUserDataStore() infrastructure.UserDataStore {
	return m.OnCreateUserDataStore()
}

func (m DataStoreFactoryMock) CreateCawDataStore() infrastructure.CawDataStore {
	return m.OnCreateCawDataStore()
}

func (m DataStoreFactoryMock) Close() {
}

func TestGetUserHandler(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		Name               string
		UserID             string
		UserDataStoreError error
		ExpectedStatusCode int
	}{
		{
			Name:               "GetExistingUserReturnsOKTest",
			UserID:             "1",
			UserDataStoreError: nil,
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name:               "GetNotExistingUserReturnsNotFoundTest",
			UserID:             "2",
			UserDataStoreError: infrastructure.ErrNotFound,
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Name:               "GetExistingUserDuringDataStoreOutageReturnInternalServerErrorTest",
			UserID:             "1",
			UserDataStoreError: errors.New("Unknow error"),
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		request := NewTestRequest(t, "GET", uriBuilder.User().WithUser(testCase.UserID).Done(), nil).
			WithAuthorization(testCase.UserID)

		userDataStoreMock := UserDataStoreMock{
			OnGetUser: func(string) (*models.User, error) {
				return &models.User{}, testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}
	}
}

func TestPostUserHandler(t *testing.T) {
	t.Parallel()
	userID := "597bd1d34ac00c75e9280ae4"
	var testCases = []struct {
		Name               string
		User               models.User
		StoreError         error
		ExpectedStoredUser *models.User
		ExpectedLocation   string
		ExpectedStatusCode int
	}{
		{
			Name: "SuccessfullyCreatedUserTest",
			User: models.User{Name: "foobar", Email: "foobar@gmail.com", Password: "foobar"},
			ExpectedStoredUser: &models.User{
				ID:    bson.ObjectIdHex(userID),
				Name:  "foobar",
				Email: "foobar@gmail.com",
			},
			StoreError:         nil,
			ExpectedLocation:   uriBuilder.User().WithUser("597bd1d34ac00c75e9280ae4").Done(),
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "UserWithoutRequiredFiledsReturnsErrorTest",
			User:               models.User{},
			ExpectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:               "WrongEmailFormatReturnsErrorTest",
			User:               models.User{Name: "foobar", Email: "foobargmail.com", Password: "foobar"},
			ExpectedStoredUser: nil,
			ExpectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			Name:               "StorageSystemOutageReturnsErrorTest",
			User:               models.User{Name: "foobar", Email: "foobar@gmail.com", Password: "foobar"},
			StoreError:         errors.New("Unknow error"),
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			Name:               "StorageSystemOutageReturnsErrorTest",
			User:               models.User{Name: "foobar", Email: "foobar@gmail.com", Password: "foobar"},
			StoreError:         infrastructure.ErrUserExists,
			ExpectedStatusCode: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		request := NewTestRequest(t, "POST", uriBuilder.User().Done(), marshalUser(testCase.User, t)).
			WithAuthorization(testCase.User.ID.Hex())

		userDataStoreMock := &UserDataStoreMock{
			OnStoreUser: func(user models.User) (*models.User, error) {
				return testCase.ExpectedStoredUser, testCase.StoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		result := recorder.Result()

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}

		location := result.Header.Get("Location")
		if testCase.ExpectedLocation != location {
			t.Errorf("Locations are not equal: expected %s given %s", testCase.ExpectedLocation, location)
		}

		if testCase.ExpectedStoredUser != nil {

			var bodyUser models.User
			decoder := json.NewDecoder(result.Body)
			if err := decoder.Decode(&bodyUser); err != nil {
				t.Fatal(err)
			}

			if !testCase.ExpectedStoredUser.Equal(bodyUser) {
				t.Errorf("Users are not equal: expected %v given %v", *testCase.ExpectedStoredUser, bodyUser)
			}
		}
	}
}

func TestGetUserFollowersHandler(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		Name               string
		UserID             string
		UserDataStoreError error
		ExpectedStatusCode int
		ExpectedFollowers  []models.Follow
	}{
		{
			Name:               "SuccessfullyRetrivedUserFollowersTest",
			UserID:             "1",
			UserDataStoreError: nil,
			ExpectedStatusCode: http.StatusOK,
			ExpectedFollowers: []models.Follow{models.Follow{
				bson.NewObjectId(), "first following"},
				models.Follow{bson.NewObjectId(), "second following"},
			}},
		{Name: "UserNotFoundTest",
			UserID:             "2",
			UserDataStoreError: infrastructure.ErrNotFound,
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedFollowers:  []models.Follow{}},
		{Name: "UnknowInternalErrorTest",
			UserID:             "3",
			UserDataStoreError: errors.New("Unknow error"),
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedFollowers:  []models.Follow{}},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		request := NewTestRequest(t, "GET", uriBuilder.User().WithUser(testCase.UserID).Followers().Done(), nil).
			WithAuthorization(testCase.UserID)

		userDataStoreMock := &UserDataStoreMock{
			OnGetUserFollowers: func(userID string) ([]models.Follow, error) {
				return testCase.ExpectedFollowers, testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}

		if testCase.ExpectedStatusCode == http.StatusCreated {
			var followers []models.Follow
			err := json.NewDecoder(recorder.Body).Decode(&followers)
			if err != nil {
				t.Errorf("Error while decoding body: ", err)
			}

			if len(followers) != len(testCase.ExpectedFollowers) {
				t.Errorf("Followers count does not meet requirements: expected %d given %d", len(testCase.ExpectedFollowers), len(followers))
			}

			for i, follow := range followers {
				expectedFollower := testCase.ExpectedFollowers[i]
				if expectedFollower.Equal(follow) {
					t.Errorf("Follow is not equal expected follower: expected %d given %d", expectedFollower, follow)
				}
			}
		}
	}
}

func TestGetUserFollowingHandler(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		Name               string
		UserID             string
		UserDataStoreError error
		ExpectedStatusCode int
		ExpectedFollowing  []models.Follow
	}{
		{
			Name:               "SuccessfullyRetrivedUserFollowingTest",
			UserID:             "1",
			UserDataStoreError: nil,
			ExpectedStatusCode: http.StatusOK,
			ExpectedFollowing: []models.Follow{models.Follow{
				bson.NewObjectId(), "first following"},
				models.Follow{bson.NewObjectId(), "second following"},
			}},
		{
			Name:               "UserDoesNotExistsTest",
			UserID:             "2",
			UserDataStoreError: infrastructure.ErrNotFound,
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedFollowing:  []models.Follow{}},
		{
			Name:               "UnknowInternalErrorTest",
			UserID:             "3",
			UserDataStoreError: errors.New("Unknow error"),
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedFollowing:  []models.Follow{}},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		request := NewTestRequest(t, "GET", uriBuilder.User().WithUser(testCase.UserID).Following().Done(), nil).
			WithAuthorization(testCase.UserID)
		userDataStoreMock := &UserDataStoreMock{
			OnGetUserFollowing: func(userID string) ([]models.Follow, error) {
				return testCase.ExpectedFollowing, testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}

		if testCase.ExpectedStatusCode == http.StatusOK {
			var following []models.Follow
			err := json.NewDecoder(recorder.Body).Decode(&following)
			if err != nil {
				t.Errorf("Error while decoding body: ", err)
			}

			if len(following) != len(testCase.ExpectedFollowing) {
				t.Errorf("Following count does not meet requirements: expected %v given %v", len(testCase.ExpectedFollowing), len(following))
			}

			for i, follow := range following {
				expectedFollowing := testCase.ExpectedFollowing[i]
				if !expectedFollowing.Equal(follow) {
					t.Errorf("Follow is not equal expected follow: expected %v given %v", expectedFollowing, follow)
				}
			}
		}
	}
}

func TestDeleteUserHandler(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name                 string
		UserIDForDeletion    bson.ObjectId
		AuthenticateAsUserID bson.ObjectId
		UserDataStoreError   error
		ExpectedStatusCode   int
	}{
		{
			Name:                 "SuccessfullyDeletedUserTest",
			UserIDForDeletion:    bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
			AuthenticateAsUserID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
			UserDataStoreError:   nil,
			ExpectedStatusCode:   http.StatusOK,
		},
		{
			Name:                 "DeleteUserDifferentThanFromUserClaimsTest",
			UserIDForDeletion:    bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
			AuthenticateAsUserID: bson.NewObjectId(),
			UserDataStoreError:   nil,
			ExpectedStatusCode:   http.StatusForbidden,
		},
		{
			Name:                 "InvalidUserIDTest",
			UserIDForDeletion:    bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
			AuthenticateAsUserID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"),
			UserDataStoreError:   errors.New("Unknow error"),
			ExpectedStatusCode:   http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		request := NewTestRequest(t, "DELETE", uriBuilder.User().WithUser(testCase.UserIDForDeletion.Hex()).Done(), nil).
			WithAuthorization(testCase.AuthenticateAsUserID.Hex())

		userDataStoreMock := &UserDataStoreMock{
			OnDeleteUser: func(userID string) error {
				return testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if recorder.Code != testCase.ExpectedStatusCode {
			t.Errorf("Wrong status code expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}
	}
}

func TestPostUserFollowingHandler(t *testing.T) {
	t.Parallel()
	followerID := bson.ObjectIdHex("597bd1d34ac00c75e9280ae4")
	followingID := bson.ObjectIdHex("5982201ebaa07aa3c968ce57")

	var testCases = []struct {
		Name                   string
		Follower               models.Follow
		Following              models.Follow
		UserDataStoreError     error
		ExpectedLocationHeader string
		ExpectedStatusCode     int
	}{
		{
			Name:                   "SuccessfullyCreatedFollowingTest",
			Follower:               models.Follow{UserID: followerID, Name: "follower user"},
			Following:              models.Follow{UserID: followingID, Name: "followed user"},
			UserDataStoreError:     nil,
			ExpectedLocationHeader: uriBuilder.User().WithUser(followerID.Hex()).Following().WithFollowing(followingID.Hex()).Done(),
			ExpectedStatusCode:     http.StatusCreated,
		},
		{
			Name:                   "UserDoesNotExistsTest",
			Follower:               models.Follow{UserID: followerID, Name: "follower user"},
			Following:              models.Follow{UserID: followingID, Name: "followed user"},
			UserDataStoreError:     infrastructure.ErrNotFound,
			ExpectedLocationHeader: "",
			ExpectedStatusCode:     http.StatusNotFound,
		},
		{
			Name:                   "UnknowInternalServerErrorTest",
			Follower:               models.Follow{UserID: followerID, Name: "follower user"},
			Following:              models.Follow{UserID: followingID, Name: "followed user"},
			UserDataStoreError:     errors.New("Unknow error"),
			ExpectedLocationHeader: "",
			ExpectedStatusCode:     http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		js, err := json.Marshal(testCase.Following)
		if err != nil {
			t.Fatal("Cannot marshal follow", err)
		}

		request := NewTestRequest(t, "POST", uriBuilder.User().WithUser(testCase.Follower.UserID.Hex()).Following().Done(), bytes.NewBuffer(js)).
			WithAuthorization(testCase.Follower.UserID.Hex())

		userDataStoreMock := &UserDataStoreMock{
			OnAddFollowingUser: func(followerID string, followedID string) error {
				return testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}

		if testCase.ExpectedStatusCode == http.StatusCreated {
			location := recorder.Header().Get("Location")
			if location != testCase.ExpectedLocationHeader {
				t.Errorf("Created resource location %v does not equals expected %v", location, testCase.ExpectedLocationHeader)
			}
		}
	}
}

func TestDeleteUserFollowing(t *testing.T) {
	t.Parallel()
	followerID := bson.NewObjectId().Hex()
	followingID := bson.NewObjectId().Hex()

	testCases := []struct {
		Name                 string
		FollowerID           string
		FollowingID          string
		UserDataStoreError   error
		AuthenticateAsUserID string
		ExpectedStatusCode   int
	}{
		{
			Name:                 "SuccessfullyDeletedFollowingTest",
			FollowerID:           followerID,
			FollowingID:          followingID,
			UserDataStoreError:   nil,
			AuthenticateAsUserID: followerID,
			ExpectedStatusCode:   http.StatusOK,
		},
		{
			Name:                 "UserNotFoundTest",
			FollowerID:           followerID,
			FollowingID:          followingID,
			UserDataStoreError:   infrastructure.ErrNotFound,
			AuthenticateAsUserID: followerID,
			ExpectedStatusCode:   http.StatusNotFound,
		},
		{
			Name:                 "UnsupportedErrorTest",
			FollowerID:           followerID,
			FollowingID:          followingID,
			UserDataStoreError:   errors.New("Unknow error"),
			AuthenticateAsUserID: followerID,
			ExpectedStatusCode:   http.StatusInternalServerError,
		},
		{
			Name:                 "UnfollowByUserThatDoesNotOwnFollowingRelationTest",
			FollowerID:           followerID,
			FollowingID:          followingID,
			UserDataStoreError:   nil,
			AuthenticateAsUserID: bson.NewObjectId().Hex(),
			ExpectedStatusCode:   http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		request := NewTestRequest(t, "DELETE", uriBuilder.User().WithUser(testCase.FollowerID).Following().WithFollowing(testCase.FollowingID).Done(), nil).
			WithAuthorization(testCase.AuthenticateAsUserID)

		userDataStoreMock := UserDataStoreMock{
			OnUnfollowUser: func(followerID, followingID string) error {
				return testCase.UserDataStoreError
			},
		}
		cawDataStoreMock := &CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)
		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		if testCase.ExpectedStatusCode != recorder.Code {
			t.Errorf("Status codes are not equal: expected %d given %d", testCase.ExpectedStatusCode, recorder.Code)
		}
	}
}

func createApp(userDataStore infrastructure.UserDataStore,
	cawDataStore infrastructure.CawDataStore) *App {
	dataStoreFactoryMock := DataStoreFactoryMock{
		OnCreateUserDataStore: func() infrastructure.UserDataStore {
			return userDataStore
		},
		OnCreateCawDataStore: func() infrastructure.CawDataStore {
			return cawDataStore
		},
	}

	return New(&utils.AppConfig{}, dataStoreFactoryMock, Logger)
}

func marshalUser(user models.User, t *testing.T) *bytes.Buffer {
	js, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Cannot marshal user", user, err)
	}
	return bytes.NewBuffer(js)
}

type TestRequest struct {
	*http.Request
	t *testing.T
}

func NewTestRequest(t *testing.T, method, uri string, body io.Reader) *TestRequest {
	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	return &TestRequest{Request: request, t: t}
}

func (r *TestRequest) WithAuthorization(userID string) *TestRequest {
	expiresAt := time.Now().Add(time.Duration(1 * time.Minute)).Unix()
	token, err := utils.NewUserToken(userID, expiresAt)
	if err != nil {
		r.t.Fatal(err)
	}
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}
