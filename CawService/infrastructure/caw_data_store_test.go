package infrastructure

import (
	"Caw/UserService/models"
	"testing"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func DropCawCollection(session *mgo.Session) {
	session.DB(db).C(cawCollection).DropCollection()
}

func TestStoreCaw(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	cawDataStore := mgoCawDataStore{session.Clone(), logger}
	defer cawDataStore.Close()

	caw := models.Caw{
		UserID:  bson.NewObjectId(),
		Message: "message",
	}

	storedCaw, err := cawDataStore.Store(caw)

	ValidateStore(t, err, storedCaw)
}

func ValidateStore(t *testing.T, err error, storedCaw *models.Caw) {
	if err != nil {
		t.Fatalf("Store Caw error: ", err)
	}

	if storedCaw == nil {
		t.Fatalf("Store does not returns stored Caw")
	}
	if storedCaw.ID.Hex() == "" {
		t.Fatal("Store does not fill Caw ID field")
	}
}

func TestGetByUserIDCaw(t *testing.T) {
	session := InitializeDataBase()
	defer session.Close()
	cawDataStore := &mgoCawDataStore{session.Clone(), logger}
	defer cawDataStore.Close()

	userID := bson.NewObjectId()

	caw := models.Caw{
		UserID:  userID,
		Message: "message",
	}
	testCases := []struct {
		Name              string
		UserID            bson.ObjectId
		Caws              []models.Caw
		ExpectedPageCount int
		ExpectedCaws      []models.Caw
	}{
		{
			Name:              "OnlyOneCawTest",
			UserID:            userID,
			Caws:              []models.Caw{caw},
			ExpectedPageCount: 1,
			ExpectedCaws:      []models.Caw{caw},
		},
		{
			Name:              "MoreThanOneCawTest",
			UserID:            userID,
			Caws:              []models.Caw{caw, caw, models.Caw{UserID: bson.NewObjectId(), Message: "Foobar"}},
			ExpectedPageCount: 1,
			ExpectedCaws:      []models.Caw{caw, caw},
		},
		{
			Name:              "MoreThanOnePageTest",
			UserID:            userID,
			Caws:              []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, models.Caw{UserID: bson.NewObjectId(), Message: "Foobar"}},
			ExpectedPageCount: 3,
			ExpectedCaws:      []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw},
		},
		{
			Name:              "CawsCountEqualToPageSizeTest",
			UserID:            userID,
			Caws:              []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, models.Caw{UserID: bson.NewObjectId(), Message: "Foobar"}},
			ExpectedPageCount: 1,
			ExpectedCaws:      []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw},
		},
		{
			Name:              "CawsCountEqualToMultiplicityOfPageSizeTest",
			UserID:            userID,
			Caws:              []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, models.Caw{UserID: bson.NewObjectId(), Message: "Foobar"}},
			ExpectedPageCount: 2,
			ExpectedCaws:      []models.Caw{caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw, caw},
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Name)
		DropCawCollection(session)
		for _, caw := range testCase.Caws {
			storedCaw, err := cawDataStore.Store(caw)

			ValidateStore(t, err, storedCaw)
		}

		for page := 0; page < testCase.ExpectedPageCount; page++ {
			storedCaws, err := cawDataStore.GetByUserID(userID.Hex(), page)
			if err != nil {
				t.Fatal(err)
			}

			lenExpectedCaws := len(testCase.ExpectedCaws)
			expectedCawsOnPage := 0
			if lenExpectedCaws-pageSize*(page+1) >= 0 {
				expectedCawsOnPage = pageSize
			} else {
				expectedCawsOnPage = lenExpectedCaws % pageSize
			}
			if len(storedCaws) != expectedCawsOnPage {
				t.Fatalf("Caws count are different expected %d, given %d", expectedCawsOnPage, len(storedCaws))
			}

			for _, storedCaw := range storedCaws {
				if storedCaw.UserID != testCase.UserID {
					t.Fatal("Caw UserID is not equal to expected. Expected %s, given %s", testCase.UserID.Hex(), storedCaw.UserID.Hex())
				}
			}

			//check if GetByUserID does not returns more pages
			if page+1 == testCase.ExpectedPageCount {
				storedCaws, err := cawDataStore.GetByUserID(userID.Hex(), page+1)
				if err != nil {
					t.Fatal(err)
				}
				if len(storedCaws) != 0 {
					t.Fatal("GetByUserID returns more pages than expected")
				}
			}
		}
	}
}
