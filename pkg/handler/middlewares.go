package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
	"gitlab.com/innoserver/pkg/writer"
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
		groupUid := r.URL.Query().Get("group_uid")
		if groupUid == "" {
			h.ServeHTTP(w, r)
			return
		}
		user, err := GetCurrentUser(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if groupRepo, ok := r.Context().Value("group_repository").(*groupRepository); ok {
			group, err := (*groupRepo).GetByUid(r.Context(), groupUid)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			inGroup, err := (*groupRepo).IsUserInGroup(r.Context(), user, group)
			if !inGroup {
				logrus.Error("user " + user.Name + " is not in group " + group.Title)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func errorWrapper(f func(http.ResponseWriter, *http.Request) (error, int)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew := writer.New(w)
		config, ok := r.Context().Value("config").(*model.Config)
		err, status := f(ew, r)
		if err != nil {
			log, _ := r.Context().Value("log").(*logrus.Logger)
			log.WithField("url", r.URL.String()).Error(err)
			if ok && config.RunLevel == "debug" {
				rlog, _ := r.Context().Value("rlog").(*logrus.Logger)
				rlog.SetOutput(ew)
				SetJsonHeader(ew)
				ew.WriteHeader(status)
				rlog.WithFields(logrus.Fields{
					"url": r.URL.String(),
				}).Error(err)
				return
			}
		}
		if status != http.StatusOK && ew.Status == 0 {
			ew.WriteHeader(status)
		}
	})
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if log, ok := r.Context().Value("log").(*logrus.Logger); ok {
			log.WithField("url", r.URL.String()).Debugln("request made")
		}
		if rlog, ok := r.Context().Value("rlog").(*logrus.Logger); ok {
			rlog.SetOutput(w)
		}
		h.ServeHTTP(w, r)
	})
}
