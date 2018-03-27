package infrastructure

import (
	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

type DataStoreFactory interface {
	CreateUserDataStore() UserDataStore
	CreateCawDataStore() CawDataStore
	Close()
}

type mgoDataStoreFactory struct {
	session *mgo.Session
	logger  *logrus.Logger
}

func NewFactory(session *mgo.Session, logger *logrus.Logger) DataStoreFactory {
	return &mgoDataStoreFactory{
		session: session,
		logger:  logger,
	}
}

func (f mgoDataStoreFactory) CreateUserDataStore() UserDataStore {
	return &mgoUserDataStore{
		session: f.session.Clone(),
		logger:  f.logger,
	}
}

func (f mgoDataStoreFactory) CreateCawDataStore() CawDataStore {
	return &mgoCawDataStore{
		session: f.session.Clone(),
		logger:  f.logger,
	}
}

func (f mgoDataStoreFactory) Close() {
	f.session.Close()
}
