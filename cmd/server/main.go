package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/handler"
	"gitlab.com/innoserver/pkg/model"
	"gitlab.com/innoserver/pkg/repository"
)

func main() {
	config := &model.Config{}
	configPtr := flag.String("config", "./init/config.json", "path to the json config file")
	flag.Parse()

	if configJson, err := ioutil.ReadFile(*configPtr); err == nil {
		if err = json.Unmarshal(configJson, config); err != nil {
			logrus.Println("error parsing config file", *configPtr)
		}
	}
	connectionStr := config.DatabaseUser + ":" + config.DatabasePassword + "@tcp(" +
		config.DatabaseAddress + ":" + config.DatabasePort + ")/" + config.Database +
		"?parseTime=true"
	db, err := sqlx.Open("mysql", connectionStr)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer db.Close()
	logrus.Infoln("server started")

	userRepository, err := repository.NewUserRepository(db)
	if err != nil {
		logrus.Errorln("error creating the user repository: ", err)
	}
	postRepository, err := repository.NewPostRepository(db)
	if err != nil {
		logrus.Errorln("error creating the user repository: ", err)
	}

	defer func() {
		logrus.Println("closing database statements")
		if err = userRepository.Close(); err != nil {
			logrus.Errorln("user repository:", err.Error())
		}
		if err = postRepository.Close(); err != nil {
			logrus.Errorln("post repository:", err.Error())
		}
	}()

	srvStr := config.ServerAddress + ":" + config.ServerPort
	srv := &http.Server{
		Addr:         srvStr,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		Handler: handler.NewHandler(
			userRepository,
			postRepository,
			config,
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		logrus.Errorln("error during server setup: ", err)
	}
}
