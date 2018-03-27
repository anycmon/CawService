package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// POST /v1/users/{id}/caws
func (app *App) postCawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userClaims, ok := r.Context().Value("userClaims").(*utils.UserClaims)
	if !ok {
		app.logger.Errorf("Cannot retrieve user claims")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	userID := vars["userID"]
	if userClaims.UserId != userID {
		app.logger.Errorf("Cannot to post caw by not authorized user: %v", userID, userClaims.UserId)
		w.WriteHeader(http.StatusUnauthorized)
	}

	caw, err := models.CawFromJSON(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.logger.Errorf("Cannot decode payload. err: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if caw.UserID.Hex() != userID {
		errMsg := "User id in URI and JSON have to be the same"
		app.logger.Error(errMsg + fmt.Sprintf("%s %s", caw.UserID.Hex(), userID))
		writeErrMsg(http.StatusBadRequest, errMsg, w)
		return
	}

	userDataStore := app.newUserDataStore()
	defer userDataStore.Close()
	user, err := userDataStore.GetUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	caw.UserName = user.Name
	cawDataStore := app.newCawDataStore()
	defer cawDataStore.Close()

	storedCaw, err := cawDataStore.Store(*caw)
	if err != nil {
		app.logger.Errorf("Cannot store caw. err: %s", err)
		writeErrMsg(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), w)
		return
	}
	if storedCaw == nil {
		app.logger.Errorf("Stored caw is nil. err: %s", err)
		writeErrMsg(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), w)
		return
	}
	jsCaw, err := storedCaw.ToJSON()
	if err != nil {
		app.logger.Errorf("Cannot convert caw into js. err: %s", err)
		writeErrMsg(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), w)
		return
	}

	uriBuilder := utils.NewUriBuilder()
	w.Header().Add("Location", uriBuilder.User().WithUser(userID).Caws().WithCaw(storedCaw.ID.Hex()).Done())
	w.WriteHeader(http.StatusCreated)
	w.Write(jsCaw)
}

// DELETE /v1/users/{userID}/caws/{cawId}
func (app *App) deleteCawHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	cawID := vars["cawId"]

	cawDataStore := app.newCawDataStore()
	defer cawDataStore.Close()

	caw, err := cawDataStore.GetByID(cawID)
	if err != nil {
		app.logger.Errorf("Cannot get caw. cawID %s, err: %s", cawID, err)
		writeErrMsg(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), w)
		return
	}

	if caw.UserID.Hex() != userID {
		writeErrMsg(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), w)
		return
	}

	if err = cawDataStore.Delete(cawID); err != nil {
		if err == infrastructure.ErrNotFound {
			writeErrMsg(http.StatusNotFound, http.StatusText(http.StatusNotFound), w)
			return
		}
		writeErrMsg(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /v1/users/{userID}/caws?page=$
func (app *App) getUserCawsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	page := 0
	vals := r.URL.Query()
	qPage, ok := vals["page"]
	if ok && len(qPage) >= 1 {
		if v, err := strconv.Atoi(qPage[0]); err != nil {
			page = v
		}
	}
	cawDataStore := app.newCawDataStore()
	defer cawDataStore.Close()
	caws, err := cawDataStore.GetByUserID(userID, page)
	if err != nil {
		return
	}
	jsCaws, err := json.Marshal(caws)

	w.Write(jsCaws)
}
