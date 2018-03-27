package utils

import (
	"testing"
	"time"
)

func TestUserToken(t *testing.T) {
	t.Parallel()
	userId := "foobar"
	expiresAt := time.Now().Add(time.Duration(10 * time.Minute))
	tokenString, err := NewUserToken("foobar", expiresAt.Unix())
	if err != nil {
		t.Error(err)
	}

	if tokenString == "" {
		t.Error("Empty token")
	}

	claims, err := DecodeUserToken(tokenString)
	if err != nil {
		t.Error(err)
	}

	if claims.UserId != userId {
		t.Errorf("Invalid user_id expected %v given %v", userId, claims.UserId)
	}

	if claims.ExpiresAt != expiresAt.Unix() {
		t.Errorf("Invalid expires_at expected %v given %v", expiresAt.Unix(), claims.ExpiresAt)
	}
}

func TestCheckHashPassword(t *testing.T) {
	t.Parallel()
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}

	testCases := []struct {
		Name           string
		Password       string
		Hash           string
		ExpectedResult bool
	}{
		{
			"SuccessfullyCheckedPasswordTest",
			password,
			hash,
			true,
		},
		{
			"PasswordDoesNotMatchToHashTest",
			"DifferentPassword",
			hash,
			false,
		},
		{
			"InvalidHashTest",
			password,
			"InvalidHash",
			false,
		},
	}

	for _, testCase := range testCases {
		if CheckPasswordHash(testCase.Password, testCase.Hash) != testCase.ExpectedResult {
			t.Error("For data: password %v hash %v expected %v given %v",
				testCase.Password, testCase.Hash, testCase.ExpectedResult, !testCase.ExpectedResult)
		}
	}
}

func TestHashPassword(t *testing.T) {
	t.Parallel()
	_, err := HashPassword("1")
	if err != nil {
		t.Error(err)
	}
}
