package main

import (
		"encoding/json" 	
		"net/http" 	
		"time" 	
		"github.com/dgrijalva/jwt-go" 	
		"github.com/sirupsen/logrus" 	
		"gitlab.com/innoserver/pkg/model"

	
		//https://github.com/auth0/go-jwt-middleware	
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandartClass
}


// Create Testuser, for response porposes only
func (s *Handler) createTestuser(w http.ResposeWriter, r *http.Request) {
	
	testUser := &model.User{
		Name: "foobar"
		Email: "foo@bar.com"
		Imei: "asdf1234"
		password: "secretPassword"
	}
	
	if err != nil {
		logrus.ErrorLn("Login: error decoding json body", err)
		w.WriterHeader(http.StatusBadRequest)
		return
	}

	epirationTime := time.Now().Add(5* time.Minute)

	claim_user := &Claims{
		Username: testUser.Name,
		StandartClaims: jwt.StandartClaims {
			ExpiresAt: expirationTime.Unix()
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriterHeader(http.StatusInternalServerError
		return 
	}

	if _, err := w.Write([]byte(tokenString)); err != nil {
		w.WriterHeader(http.StatusInternalServerError)
		return 
	}
	
}
