package main

import (
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/handler"
	"gitlab.com/innoserver/pkg/repository"
)

func main() {
	db, err := sqlx.Open("mysql", "ip:password@tcp(127.0.0.1:3306)/innovision?parseTime=true")
	if err != nil {
		logrus.Fatalln(err)
	}
	defer db.Close()
	logrus.Infoln("server started")

	userRepository, err := repository.NewUserRepository(db)
	if err != nil {
		logrus.Errorln("error creating the user repository: ", err)
	}

	// TODO load adress and port from config
	srv := &http.Server{
		Addr:         "0.0.0.0:5000",
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		Handler: handler.NewHandler(
			userRepository,
		),
	}

	if err := srv.ListenAndServe(); err != nil {
		logrus.Errorln("error during server setup: ", err)
	}
}
