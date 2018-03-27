package models

import (
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestUserModelValidation(t *testing.T) {
	testCases := []struct {
		Name            string
		UserModel       User
		ExcpectedResult bool
	}{
		{
			"ValidUserTest",
			User{Name: "user", Email: "user@email.com", Password: "foobar"},
			true,
		},
		{
			"EmptyUserNameTest",
			User{Name: "", Email: "user@email.com", Password: "foobar"},
			false,
		},
		{
			"InvalidEmailTest",
			User{Name: "user", Email: "useremail.com", Password: "foobar"},
			false,
		},
		{
			"EmptyPasswordTest",
			User{Name: "user", Email: "user@email.com", Password: ""},
			false,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		if result := testCase.UserModel.IsValid(); result != testCase.ExcpectedResult {
			t.Errorf("User %v validation returned %t expected %t", testCase.UserModel, result, testCase.ExcpectedResult)
		}
	}
}

func TestUserModelEquality(t *testing.T) {
	testCases := []struct {
		Name            string
		First           User
		Second          User
		ExcpectedResult bool
	}{
		{
			"EqualUsersTest",
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "user", Email: "user@email.com", Password: "foobar"},
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "user", Email: "user@email.com", Password: "foobar"},
			true,
		},
		{
			"DifferentIDTest",
			User{ID: bson.ObjectIdHex("59c2ee426c63c2024d00d684"), Name: "user", Email: "user@email.com", Password: "foobar"},
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "user", Email: "user@email.com", Password: "foobar"},
			false,
		},
		{
			"DifferentNameTest",
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "first", Email: "user@email.com", Password: "foobar"},
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "second", Email: "user@email.com", Password: "foobar"},
			false,
		},
		{
			"DifferentEmailAddressTest",
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "user", Email: "first@email.com", Password: "foobar"},
			User{ID: bson.ObjectIdHex("597bd1d34ac00c75e9280ae4"), Name: "user", Email: "second@email.com", Password: "foobar"},
			false,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		if result := testCase.First.Equal(testCase.Second); result != testCase.ExcpectedResult {
			t.Errorf("Equal returns %t expected %t for users %v, %v", result, testCase.ExcpectedResult, testCase.First, testCase.Second)
		}
	}
}
