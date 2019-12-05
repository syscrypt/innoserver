package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, _ := r.Context().Value("config").(*model.Config)
		w.Header().Set("Access-Control-Allow-Origin", config.AccessControlAllowOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", config.AccessControlAllowCredentials)
		w.Header().Set("Access-Control-Allow-Methods", config.AccessControlAllowMethods)
		w.Header().Set("Access-Control-Allow-Headers", config.AccessControlAllowHeaders)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func keyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, _ := r.Context().Value("config").(*model.Config)
		if r.Header.Get("API_KEY") != config.ApiKey && config.ApiKey != "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func authenticationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("X-Auth-Token")
		if tokenStr != "" {
			claims := &model.Claims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					errStr := "unexpected signing method"
					logrus.Errorln(errStr)
					return nil, errors.New(errStr)
				}
				if config, ok := r.Context().Value("config").(*model.Config); ok {
					return []byte(config.JwtSecret), nil
				}
				return nil, nil
			})
			if err != nil {
				logrus.Errorln("parsing incoming jw-token failed:", err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if userRepo, ok := r.Context().Value("user_repository").(*userRepository); ok {
				user, err := (*userRepo).GetByEmail(r.Context(), claims.Email)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r = r.WithContext(context.WithValue(r.Context(), "user", user))
			}
			h.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
}

func groupMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func errorWrapper(f func(http.ResponseWriter, *http.Request) (error, int)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, ok := r.Context().Value("config").(*model.Config)
		err, status := f(w, r)
		if err != nil {
			logrus.Error(r.URL.String() + ": " + err.Error())
			if ok && config.RunLevel == "debug" {
				w.Header().Set("content-type", "application/json")
				errResp := &model.ErrorResponse{}
				errResp.Message = r.URL.String() + ": " + err.Error()
				errStr, _ := json.Marshal(errResp)
				w.WriteHeader(status)
				w.Write([]byte(errStr))
			}
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
		}
	})
}
