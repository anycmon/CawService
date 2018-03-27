package infrastructure

import (
	"Caw/UserService/models"
	"testing"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const mongoAddress = "database:27017"

var logger = logrus.New()

func InitializeDataBase() *mgo.Session {
	session, err := mgo.Dial(mongoAddress)
	if err != nil {
		panic(err)
	}

	err = session.DB(db).DropDatabase()
	if err != nil {
		panic(err)
	}
	return session
}

func TestStoreUser(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	user := models.User{
		ID:    bson.NewObjectId(),
		Name:  "StoreUser",
		Email: "StoreUser@email.com",
	}

	storedUser, err := userDataStore.StoreUser(user)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", user)
	}

	if storedUser.ID.Hex() == "" {
		t.Fatal("StoreUser does not fill User ID field")
	}
}

func TestStoreDuplicatedUser(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	user := models.User{
		ID:    bson.NewObjectId(),
		Name:  "StoreUser",
		Email: "StoreUser@email.com",
	}

	storedUser, err := userDataStore.StoreUser(user)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", user)
	}

	if storedUser.ID.Hex() == "" {
		t.Fatal("StoreUser does not fill User ID field")
	}

	storedUser, err = userDataStore.StoreUser(user)

	if err != ErrUserExists {
		t.Fatalf("Should return ErrUserExists")
	}
}

func TestGetUser(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	user := models.User{
		ID:    bson.NewObjectId(),
		Name:  "GetUser",
		Email: "GetUser@email.com",
	}

	storedUser, err := userDataStore.StoreUser(user)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", user)
	}

	retrievedUser, err := userDataStore.GetUser(storedUser.ID.Hex())
	if err != nil {
		t.Fatal(err)
	}

	if !retrievedUser.Equal(*storedUser) {
		t.Fatal("Retrieved user is not equal to stored user\n", retrievedUser, storedUser)
	}
}

func TestDeleteUser(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	testUser := models.User{
		Name:     "Delete Test",
		Email:    "Delete@test.com",
		Password: "password",
	}

	storedUser, err := userDataStore.StoreUser(testUser)
	if err != nil {
		t.Errorf("%v", err)
	}

	anotherUser := models.User{
		Name:     "Another User",
		Email:    "anotheruser@test.com",
		Password: "password",
	}

	_, err = userDataStore.StoreUser(anotherUser)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = userDataStore.DeleteUser(storedUser.ID.Hex())
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestAddFollowedUser(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	follower := models.User{
		ID:    bson.NewObjectId(),
		Name:  "sadistic",
		Email: "follower@email.com",
	}

	following := models.User{
		ID:    bson.NewObjectId(),
		Name:  "following user",
		Email: "following@email.com",
	}

	storedfollower, err := userDataStore.StoreUser(follower)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", follower)
	}

	storedFollowing, err := userDataStore.StoreUser(following)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", following)
	}

	err = userDataStore.AddFollowingUser(storedfollower.ID.Hex(), storedFollowing.ID.Hex())
	if err != nil {
		t.Fatal("AddFollowingUser err", err)
	}
}

func TestGetFollowedUsers(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	userDataStore := NewMgoUserDataStore(session, logger)
	defer userDataStore.Close()

	follower := models.User{
		ID:    bson.NewObjectId(),
		Name:  "follower user",
		Email: "follower@email.com",
	}

	following := models.User{
		ID:    bson.NewObjectId(),
		Name:  "following user",
		Email: "following@email.com",
	}

	storedFollower, err := userDataStore.StoreUser(follower)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", follower)
	}

	storedFollowing, err := userDataStore.StoreUser(following)

	if err != nil {
		t.Fatalf("StoreUser error: ", err, "user", following)
	}

	err = userDataStore.AddFollowingUser(storedFollower.ID.Hex(), storedFollowing.ID.Hex())
	if err != nil {
		t.Fatal("AddFollowingUser err", err)
	}

	followingUsers, err := userDataStore.GetUserFollowing(storedFollower.ID.Hex())
	if err != nil {
		t.Fatal("GetUserFollowed err", err)
	}

	if len(followingUsers) != 1 {
		t.Fatal("GetUserFollowed returns wrong users count", len(followingUsers), 1)
	}

	expectedFollowingUser := models.Follow{
		Name:   storedFollowing.Name,
		UserID: storedFollowing.ID,
	}

	if !followingUsers[0].Equal(expectedFollowingUser) {
		t.Fatal("Following user is not equal to expected user", followingUsers, expectedFollowingUser)
	}
}
