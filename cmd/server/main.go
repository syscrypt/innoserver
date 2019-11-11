package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func main() {
	db, err := sqlx.Open("mysql", "ip:password@tcp(127.0.0.1:3306)/innovision?parseTime=true")
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Infoln("server started")

	defer db.Close()
	for {

	}
}
