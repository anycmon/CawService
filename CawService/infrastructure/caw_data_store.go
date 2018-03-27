package infrastructure

import (
	"Caw/UserService/models"
	"time"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	cawCollection = "caw"
	pageSize      = 10
)

type CawDataStore interface {
	Store(caw models.Caw) (*models.Caw, error)
	GetByID(cawID string) (*models.Caw, error)
	GetByUserID(userID string, page int) ([]models.Caw, error)
	Delete(cawID string) error
	Close()
}

type mgoCawDataStore struct {
	session *mgo.Session
	logger  *logrus.Logger
}

func (ds *mgoCawDataStore) caw() *mgo.Collection {
	return ds.session.DB(db).C(cawCollection)
}

func (ds *mgoCawDataStore) Store(caw models.Caw) (*models.Caw, error) {
	caw.ID = bson.NewObjectId()
	caw.CreatedAt = time.Now()
	err := ds.caw().Insert(&caw)
	if err != nil {
		ds.logger.Error(err)
		return nil, err
	}

	return &caw, nil
}

func (ds *mgoCawDataStore) GetByUserID(userID string, page int) ([]models.Caw, error) {
	var storedCaws []models.Caw
	err := ds.caw().
		Find(bson.M{"user_id": bson.ObjectIdHex(userID)}).
		Sort("created_at").
		Skip(page * pageSize).
		Limit(pageSize).
		All(&storedCaws)
	if err != nil {
		return nil, err
	}

	return storedCaws, nil
}

func (ds *mgoCawDataStore) GetByID(cawID string) (*models.Caw, error) {
	var caw models.Caw
	if !bson.IsObjectIdHex(cawID) {
		ds.logger.Error("Caw Id is not mongo ObjectId")
		return nil, ErrNotFound
	}

	err := ds.caw().
		FindId(bson.ObjectIdHex(cawID)).One(&caw)

	if err == mgo.ErrNotFound {
		ds.logger.Error(err)
		return nil, ErrNotFound
	} else if err != nil {
		ds.logger.Error(err)
		return nil, err
	}
	return &caw, nil
}

func (ds *mgoCawDataStore) Delete(cawID string) error {
	err := ds.caw().RemoveId(bson.ObjectIdHex(cawID))
	if err != nil {
		ds.logger.Error(err)
	}
	return err
}

func (ds *mgoCawDataStore) Close() {
	ds.session.Clone()
}
