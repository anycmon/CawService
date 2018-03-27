package main

import (
	"Caw/UserService/app"
	"Caw/UserService/infrastructure"
	"Caw/UserService/utils"
	"context"
	"net/http"
	"os"
	"os/signal"

	mgo "gopkg.in/mgo.v2"

	"github.com/Sirupsen/logrus"
)

func main() {
	appConfig := utils.New()
	logger := logrus.New()
	logger.Infof("Connecting to mongo %s", appConfig.Mongo)
	session, err := mgo.Dial(appConfig.Mongo)
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	logger.Info("Connected to mongo")

	mgoDataStoreFactory := infrastructure.NewFactory(session, logger)
	defer mgoDataStoreFactory.Close()

	app := app.New(appConfig, mgoDataStoreFactory, logger)

	appConfig.WithWatchConfig(func(appConfig *utils.AppConfig) {
		app.UpdateConfig(appConfig)
	})

	srv := http.Server{Addr: appConfig.Address, Handler: app}
	stopChan := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		logger.Info("Shutting down server...")
		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			logger.Errorf("HTTP server Shutdown: %v", err)
		}
		close(stopChan)
	}()

	logger.Infof("Listening at %s", appConfig.Address)
	srv.ListenAndServe()
	<-stopChan

	logger.Info("Server gracefully stopped")
}
