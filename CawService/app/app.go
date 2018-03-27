package app

import (
	"Caw/UserService/infrastructure"
	"Caw/UserService/middleware"
	"Caw/UserService/models"
	"Caw/UserService/utils"
	"net/http"

	"github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
)

var (
	appJSON              = "application/json"
	supportedAccept      = appJSON
	supportedContentType = appJSON
)

type App struct {
	router                *mux.Router
	dataStoreFactory      infrastructure.DataStoreFactory
	logger                *logrus.Logger
	TokenExpiresInMinutes int
}

func New(appConfig *utils.AppConfig, dataStoreFactory infrastructure.DataStoreFactory, logger *logrus.Logger) *App {
	app := App{dataStoreFactory: dataStoreFactory, logger: logger, TokenExpiresInMinutes: appConfig.TokenExpiresInMinutes}
	app.createRoute()
	return &app
}

func (app App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}

func (app *App) UpdateConfig(appConfig *utils.AppConfig) {
	app.TokenExpiresInMinutes = appConfig.TokenExpiresInMinutes
	logrus.Infof("TokenExpiresInMinutes updated to %v", app.TokenExpiresInMinutes)
}

func (app App) newUserDataStore() infrastructure.UserDataStore {
	return app.dataStoreFactory.CreateUserDataStore()
}

func (app App) newCawDataStore() infrastructure.CawDataStore {
	return app.dataStoreFactory.CreateCawDataStore()
}

func writeErrMsg(statusCode int, msg string, w http.ResponseWriter) {
	errMsg := models.Error{Message: msg}
	js, err := errMsg.ToJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(js)
}

// createRoute creates route
func (app *App) createRoute() {

	router := mux.NewRouter()
	app.addUserEndpoint(router)
	app.addCawEndpoint(router)
	app.addAuthEndpoint(router)
	app.router = router
}

func (app App) addUserEndpoint(router *mux.Router) {
	uriBuilder := utils.NewUriBuilder()

	router.HandleFunc(
		uriBuilder.User().Done(),
		middleware.Chain(app.postUserHandler,
			middleware.CQRS(),
			middleware.Consume(supportedContentType),
			middleware.Produce(supportedAccept),
			middleware.Logging(app.logger))).
		Methods("POST")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Done(),
		middleware.Chain(app.commonMiddleware(app.getUserHandler),
			middleware.Produce(supportedAccept))).
		Methods("GET")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Done(),
		app.commonMiddleware(app.deleteUserHandler)).
		Methods("DELETE")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Followers().Done(),
		app.commonMiddleware(app.getUserFollowersHandler)).
		Methods("GET")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Following().Done(),
		app.commonMiddleware(app.getUserFollowingHandler)).
		Methods("GET")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Following().Done(),
		app.commonMiddleware(app.postUserFollowingHandler)).
		Methods("POST")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Following().WithFollowing("{fid}").Done(),
		app.commonMiddleware(app.deleteUserFollowingHandler)).
		Methods("DELETE")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Caws().Done(),
		app.commonMiddleware(app.postCawHandler)).
		Methods("POST")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Caws().WithCaw("{cawId}").Done(),
		app.commonMiddleware(app.deleteCawHandler)).
		Methods("DELETE")
	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Caws().Done(),
		app.commonMiddleware(app.getUserCawsHandler)).
		Queries("page", "{[0-9]+}").
		Methods("GET")

	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Done(),
		middleware.Chain(CQRS, middleware.Logging(app.logger))).Methods("OPTIONS")

	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Caws().Done(),
		middleware.Chain(CQRS, middleware.Logging(app.logger))).Methods("OPTIONS")

	router.HandleFunc(
		uriBuilder.User().WithUser("{userID}").Caws().WithCaw("{cawId}").Done(),
		middleware.Chain(CQRS, middleware.Logging(app.logger))).Methods("OPTIONS")
}

func (app *App) addCawEndpoint(router *mux.Router) {
	//router.HandleFunc("/v1/caws/{cawID}/replies?page={page}",
	//	app.commonMiddleware(app.PostCawHandler)).
	//	Methods("GET")
}

func (app *App) addAuthEndpoint(router *mux.Router) {
	uriBuilder := utils.NewUriBuilder()

	router.HandleFunc(
		uriBuilder.Auth().Done(),
		middleware.Chain(app.authenticateHandler,
			middleware.Consume(supportedContentType),
			middleware.Produce(supportedAccept),
			middleware.Logging(app.logger))).
		Methods("POST")
	router.HandleFunc(
		uriBuilder.Auth().Done(),
		middleware.Chain(CQRS, middleware.Logging(app.logger))).Methods("OPTIONS")
}

func (app App) commonMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return middleware.Chain(f,
		middleware.MustAuth(app.logger),
		middleware.CQRS(),
		middleware.Logging(app.logger))
}

func CQRS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
}
