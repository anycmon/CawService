package app

import (
	"Caw/UserService/models"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

type ExtendedCaw models.Caw

func (caw *ExtendedCaw) Equal(other *models.Caw) bool {
	return caw.ID == other.ID &&
		caw.UserID == other.UserID &&
		caw.ParentID == other.ParentID &&
		caw.Message == other.Message &&
		caw.RecawCount == other.RecawCount &&
		caw.RepliesCount == caw.RepliesCount &&
		caw.CreatedAt.Equal(other.CreatedAt)
}

func TestPostCawHandler(t *testing.T) {
	t.Parallel()

	userID := bson.NewObjectId()
	cawID := bson.NewObjectId()
	caw := &models.Caw{
		UserID:  userID,
		Message: "test message",
	}
	cawDiffUserId := &models.Caw{
		UserID:  bson.NewObjectId(),
		Message: "test message",
	}
	testCases := []struct {
		Name               string
		UserID             string
		Caw                *models.Caw
		CawDataStoreErr    error
		ExpectedStoredCaw  *models.Caw
		ExpectedLocation   string
		ExpectedStatusCode int
	}{
		{
			Name:               "PostCawSuccessfullyTest",
			UserID:             userID.Hex(),
			Caw:                caw,
			CawDataStoreErr:    nil,
			ExpectedStoredCaw:  caw,
			ExpectedLocation:   uriBuilder.User().WithUser(userID.Hex()).Caws().WithCaw(cawID.Hex()).Done(),
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "PostCawDifferentUserIDInURIAndPayloadTest",
			UserID:             userID.Hex(),
			Caw:                cawDiffUserId,
			CawDataStoreErr:    nil,
			ExpectedStoredCaw:  cawDiffUserId,
			ExpectedLocation:   "",
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "PostCawDataStoreErrorTest",
			UserID:             userID.Hex(),
			Caw:                caw,
			CawDataStoreErr:    errors.New("Unknow error"),
			ExpectedStoredCaw:  caw,
			ExpectedLocation:   "",
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		jsCaw, err := testCase.Caw.ToJSON()

		assert.Nil(t, err, "cannot to encode caw to JSON %v", err)

		request := NewTestRequest(
			t, "POST", uriBuilder.User().WithUser(testCase.UserID).Caws().Done(),
			bytes.NewBuffer(jsCaw)).WithAuthorization(testCase.UserID)

		assert.Nil(t, err, "cannot to prepare request %v", err)

		userDataStoreMock := UserDataStoreMock{}
		cawDataStoreMock := CawDataStoreMock{
			OnStore: func(caw models.Caw) (*models.Caw, error) {
				if testCase.ExpectedStoredCaw != nil {
					testCase.ExpectedStoredCaw.ID = cawID
				}
				return testCase.ExpectedStoredCaw, testCase.CawDataStoreErr
			},
		}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		assert.Equal(t, recorder.Code, testCase.ExpectedStatusCode,
			"status codes are different %v %v", recorder.Code, testCase.ExpectedStatusCode)

		result := recorder.Result()
		location := result.Header.Get("Location")

		assert.Equal(t, testCase.ExpectedLocation, location,
			"different location %v %v", testCase.ExpectedLocation, location)

		if recorder.Code == http.StatusCreated {
			cawFromBody, err := models.CawFromJSON(result.Body)
			assert.Nil(t, err,
				"cannot to convert JSON to caw")

			testCaw := ExtendedCaw(*cawFromBody)
			assert.True(t, testCaw.Equal(testCase.Caw),
				"returned caw is different %v %v", testCaw, testCase.Caw)
		}
	}
}

func TestGetUserCaws(t *testing.T) {
	t.Parallel()

	userID := bson.NewObjectId()
	testCases := []struct {
		Name string
	}{
		{
			Name: "IfExists",
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)

		request := NewTestRequest(
			t, "GET",
			"/v1/users/"+userID.Hex()+"/caws?page",
			nil).WithAuthorization(userID.Hex())

		userDataStoreMock := UserDataStoreMock{}
		cawDataStoreMock := CawDataStoreMock{}

		app := createApp(userDataStoreMock, cawDataStoreMock)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request.Request)

		assert.Equal(t, recorder.Code, http.StatusOK, "status code are different %v %v", recorder.Code)
	}
}
