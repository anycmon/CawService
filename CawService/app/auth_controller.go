package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// POST /v1/authentication
// authenticateHandler authenticate user by password
func (app *App) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var auth models.Auth
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&auth); err != nil {
		errMsg := "Illformated payload"
		app.logger.Error(errMsg + fmt.Sprintf(". err: %s", err))
		writeErrMsg(http.StatusBadRequest, errMsg, w)
		return
	}
	defer r.Body.Close()

	ds := app.newUserDataStore()
	defer ds.Close()
	user, err := ds.GetUserByName(auth.Name)
	if err != nil {
		app.logger.Errorf("Cannot to get user by name. name: %s, err: %s", auth.Name, err)
		if err == infrastructure.ErrNotFound {
			errMsg := "Incorrect user name or password"
			app.logger.Error(errMsg)
			writeErrMsg(http.StatusUnauthorized, errMsg, w)
			return
		}
		errMsg := http.StatusText(http.StatusInternalServerError)
		app.logger.Error(errMsg)
		writeErrMsg(http.StatusInternalServerError, errMsg, w)
		return

	}

	if !utils.CheckPasswordHash(auth.Password, user.Password) {
		errMsg := "Incorrect user name or password"
		app.logger.Error(errMsg, err)
		writeErrMsg(http.StatusUnauthorized, errMsg, w)
		return
	}

	expiresAt := time.Now().Add(time.Duration(time.Duration(app.TokenExpiresInMinutes) * time.Minute)).Unix()
	tokenString, err := utils.NewUserToken(user.ID.Hex(), expiresAt)
	if err != nil {
		errMsg := "Token is invalid"
		app.logger.Errorf(errMsg + fmt.Sprintf(". err: %s", err))
		writeErrMsg(http.StatusUnauthorized, errMsg, w)
		return
	}

	token := models.Token{
		AccessToken: tokenString,
		Type:        "Bearer",
		ExpiresAt:   time.Unix(expiresAt, 0),
	}

	js, err := json.Marshal(token)
	if err != nil {
		app.logger.Errorf("Cannot marshal to json. err: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Write(js)
}
