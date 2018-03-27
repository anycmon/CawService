package infrastructure

import (
	"Caw/UserService/models"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	userCollection           = "user"
	followRelationCollection = "followRelation"
)

// UserDataStore represents data access layer
type UserDataStore interface {
	GetUser(userID string) (*models.User, error)
	GetUserByName(userID string) (*models.User, error)
	StoreUser(user models.User) (*models.User, error)
	DeleteUser(userID string) error
	GetUserFollowers(userID string) ([]models.Follow, error)
	GetUserFollowing(userID string) ([]models.Follow, error)
	AddFollowingUser(followerID string, followingID string) error
	UnfollowUser(followerID string, followingID string) error
	Close()
}

// mgoUserDataStore implements UserDataStore and provides access to mongodb store
type mgoUserDataStore struct {
	session *mgo.Session
	logger  *logrus.Logger
}

// NewMgoUserDataStore creates mgoUserDataStore
func NewMgoUserDataStore(session *mgo.Session, logger *logrus.Logger) *mgoUserDataStore {
	dataStore := mgoUserDataStore{session: session, logger: logger}
	index := mgo.Index{
		Key:        []string{"name", "email"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := dataStore.user().EnsureIndex(index)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	return &dataStore
}

func (ds *mgoUserDataStore) Close() {
	ds.session.Close()
}

func (ds *mgoUserDataStore) user() *mgo.Collection {
	return ds.session.DB(db).C(userCollection)
}

func (ds *mgoUserDataStore) followRelation() *mgo.Collection {
	return ds.session.DB(db).C(followRelationCollection)
}

// GetUser gets and returns User with requested ID
func (ds *mgoUserDataStore) GetUser(userID string) (*models.User, error) {
	var user models.User
	if !bson.IsObjectIdHex(userID) {
		ds.logger.Error("User Id is not mongo ObjectId")
		return nil, ErrNotFound
	}

	err := ds.user().
		FindId(bson.ObjectIdHex(userID)).One(&user)

	if err == mgo.ErrNotFound {
		ds.logger.Error(err)
		return nil, ErrNotFound
	} else if err != nil {
		ds.logger.Error(err)
		return nil, err
	}

	return &user, nil
}

// GetUserByName gets and returns User with requested name
func (ds *mgoUserDataStore) GetUserByName(userName string) (*models.User, error) {
	var user models.User
	err := ds.user().Find(bson.M{"name": userName}).One(&user)

	if err == mgo.ErrNotFound {
		ds.logger.Error(err)
		return nil, ErrNotFound
	} else if err != nil {
		ds.logger.Error(err)
		return nil, err
	}

	return &user, nil
}

// StoreUser persist provided User
func (ds *mgoUserDataStore) StoreUser(user models.User) (*models.User, error) {
	user.ID = bson.NewObjectId()
	err := ds.user().Insert(&user)
	if err != nil {
		if mgo.IsDup(err) {
			return nil, ErrUserExists
		}
		ds.logger.Error(err)
		return nil, err
	}

	return &user, nil
}

// DeleteUser deletes user with provider ID
func (ds *mgoUserDataStore) DeleteUser(userID string) error {
	err := ds.user().RemoveId(bson.ObjectIdHex(userID))
	if err != nil {
		ds.logger.Error(err)
	}
	return err
}

// GetUserFollowers gets end returns followers of user with provided ID
func (ds *mgoUserDataStore) GetUserFollowers(userID string) ([]models.Follow, error) {
	var queryResult = []struct {
		Follower models.Follow `json:"follower"`
	}{}
	err := ds.followRelation().
		Find(bson.M{"following.user_id": bson.ObjectIdHex(userID)}).
		Select(bson.M{"follower": true}).All(&queryResult)

	if err != nil {
		ds.logger.Error(err)
		return []models.Follow{}, err
	}

	var followers []models.Follow
	for _, f := range queryResult {
		followers = append(followers, f.Follower)
	}

	return followers, nil
}

// GetUserFollowing gets and returns users followed by user with provided ID
func (ds *mgoUserDataStore) GetUserFollowing(userID string) ([]models.Follow, error) {
	var queryResult = []struct {
		Following models.Follow `json:"following"`
	}{}
	err := ds.followRelation().
		Find(bson.M{"follower.user_id": bson.ObjectIdHex(userID)}).
		Select(bson.M{"following": true}).All(&queryResult)

	if err != nil {
		ds.logger.Error(err)
		return []models.Follow{}, err
	}

	following := []models.Follow{}
	for _, f := range queryResult {
		following = append(following, f.Following)
	}

	return following, nil
}

// AddFollowingUser creates relationships between follower and following user
func (ds *mgoUserDataStore) AddFollowingUser(followerID string, followingID string) error {
	followerUser, err := ds.GetUser(followerID)
	if err != nil {
		ds.logger.Error(err)
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	followingUser, err := ds.GetUser(followingID)
	if err != nil {
		ds.logger.Error(err)
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	followRelation := models.FollowRelation{
		Follower: models.Follow{
			UserID: followerUser.ID,
			Name:   followerUser.Name,
		},
		Following: models.Follow{
			UserID: followingUser.ID,
			Name:   followingUser.Name,
		},
	}
	err = ds.followRelation().Insert(&followRelation)
	if err != nil {
		ds.logger.Error(err)
		return err
	}

	err = ds.incrementFollowStat(followerID, followingID)
	if err != nil {
		ds.logger.Error(err)
		return err
	}
	return nil
}

func (ds *mgoUserDataStore) UnfollowUser(followerID, followingID string) error {
	return nil
}

func (ds *mgoUserDataStore) incrementFollowStat(followerID, followingID string) error {
	err := ds.user().Update(bson.M{"_id": bson.ObjectIdHex(followingID)}, bson.M{"$inc": bson.M{"followers_count": 1}})
	if err != nil {
		ds.logger.Error(err)
		return err
	}

	err = ds.user().Update(bson.M{"_id": bson.ObjectIdHex(followerID)}, bson.M{"$inc": bson.M{"following_count": 1}})
	if err != nil {
		ds.logger.Error(err)
		return err
	}

	return nil
}
