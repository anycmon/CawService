package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// GET /v1/users/{userID}
// getUserHandler handle HTTP GET method and returns requested user
func (app *App) getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	ds := app.newUserDataStore()
	defer ds.Close()
	user, err := ds.GetUser(userID)
	if err != nil {
		app.logger.Errorf("Cannot to get user. userID: %s, err: %s", userID, err)
		if err == infrastructure.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js, err := user.ToPublic().ToJSON()
	if err != nil {
		app.logger.Errorf("Cannot to convert user to json. err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(js)
}

// POST /v1/users
// postUserHandler handle HTTP POST method by creating new user and returns his URI identifier
// and data in JSON format
func (app *App) postUserHandler(w http.ResponseWriter, r *http.Request) {
	//TODO check if body is null
	user, err := models.UserFromJSON(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.logger.Errorf("Cannot to convert payload to User. err: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !user.IsValid() {
		app.logger.Errorf("User is invalid. err: %s", user)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		app.logger.Errorf("Cannot to hash password. err: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.Password = passwordHash
	user.CreatedAt = time.Now()
	ds := app.newUserDataStore()
	defer ds.Close()
	storedUser, err := ds.StoreUser(*user)
	if err != nil {
		app.logger.Errorf("Cannot to store user. name: %s, email: %s, err: %s", user.Name, user.Email, err)
		if err == infrastructure.ErrUserExists {
			app.logger.Errorf("User already exists. name: %s, email: %s", user.Name, user.Email)
			writeErrMsg(http.StatusConflict, "User already exists", w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js, err := storedUser.ToPublic().ToJSON()
	if err != nil {
		app.logger.Errorf("Cannot to convert user to js. err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uriBuilder := utils.NewUriBuilder()
	w.Header().Add("Location", uriBuilder.User().WithUser(storedUser.ID.Hex()).Done())
	w.WriteHeader(http.StatusCreated)
	w.Write(js)
}

// DELETE /v1/users/{userID}
// deleteUserHandler deletes requested user only if requestor has access to it
func (app *App) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	userClaims, ok := r.Context().Value("userClaims").(*utils.UserClaims)
	if !ok {
		app.logger.Error("Cannot retrieve user claims")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if userClaims.UserId != userID {
		app.logger.Errorf("Cannot to delete user %v by user %v", userID, userClaims.UserId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ds := app.newUserDataStore()
	defer ds.Close()
	if err := ds.DeleteUser(userID); err != nil {
		app.logger.Errorf("Cannot to delete user. userID: %s, err: %s", userID, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

// GET /v1/users/{userID}/followers
// getUserFollowersHandler handle HTTP GET method and returns requested user followers
func (app *App) getUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	ds := app.newUserDataStore()
	defer ds.Close()
	followers, err := ds.GetUserFollowers(userID)
	if err != nil {
		app.logger.Errorf("Cannot to get user followers. userID: %s, err: %s", userID, err)
		if err == infrastructure.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(followers)
	if err != nil {
		app.logger.Errorf("Cannot to marshal followers. err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(js)
}

// GET /v1/users/{userID}/following
// getUserFollowingHandler handle HTTP GET method and returns following users by requested user
func (app *App) getUserFollowingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	ds := app.newUserDataStore()
	defer ds.Close()
	following, err := ds.GetUserFollowing(userID)
	if err != nil {
		app.logger.Errorf("Cannot to get user following. userID: %s, err: %s", userID, err)
		if err == infrastructure.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(following)
	if err != nil {
		app.logger.Errorf("Cannot to marshal user following. err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(js)
}

// POST /v1/users/{userID}/following
// postUserFollowingHandler handle HTTP POST method by adding new following user by requested user
func (app *App) postUserFollowingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followerID := vars["userID"]
	var followingUser models.Follow
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&followingUser)
	if err != nil {
		app.logger.Errorf("Cannot to convert payload to Following. err: %s", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()

	ds := app.newUserDataStore()
	defer ds.Close()
	err = ds.AddFollowingUser(followerID, followingUser.UserID.Hex())
	if err != nil {
		app.logger.Errorf("Cannot to add following user. followerID: %s, followingUser: %s, err: %s", followerID, followingUser.UserID.Hex(), err)
		if err == infrastructure.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uriBuilder := utils.NewUriBuilder()
	w.Header().Set("Location", uriBuilder.User().WithUser(followerID).Following().WithFollowing(followingUser.UserID.Hex()).Done())
	w.WriteHeader(http.StatusCreated)
}

// DELETE /v1/users/{userID}/following/{fid}
// deleteUserFollowingHandler handle HTTP DELETE method by removing follow repationship
func (app *App) deleteUserFollowingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	followingID := vars["fid"]

	userClaims, ok := r.Context().Value("userClaims").(*utils.UserClaims)
	if !ok {
		app.logger.Errorf("Cannot retrieve user claims")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if userClaims.UserId != userID {
		app.logger.Errorf("Cannot delete user following %v by user %v", userID, userClaims.UserId)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	ds := app.newUserDataStore()
	defer ds.Close()

	err := ds.UnfollowUser(userID, followingID)
	if err != nil {
		app.logger.Errorf("Cannot to unfollow user: %s, followingID: %s, err %s", userID, followingID, err)
		if err == infrastructure.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
