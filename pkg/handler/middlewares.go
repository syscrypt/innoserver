package handler

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
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
			if !inGroup && !group.Public {
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
		ew := writer.
		config, ok := r.Context().Value("config").(*model.Config)
		err, status := f(ew, r)
		if err != nil {
			log, _ := r.Context().Value("log").(*logrus.Entry)
			log.Error(err)
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
		if log, ok := r.Context().Value("log").(*logrus.Entry); ok {
			logger := log.WithField("url", r.URL.String())
			logger.Debugln("incoming request")
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.WithError(err).Debugln("error while decode request body")
				h.ServeHTTP(w, r)
				return
			}
			rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
			rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
			r.Body = rdr2
			if _, _, err := r.FormFile("file"); err == nil {
				log.WithField("file", "true").Debugln("request body")
				h.ServeHTTP(w, r)
				return
			}
			log.WithFields(logrus.Fields{
				"body":   rdr1,
				"ip":     r.RemoteAddr,
				"method": r.Method,
				"header": r.Header,
			}).Debugln("request body")
		}
		if rlog, ok := r.Context().Value("rlog").(*logrus.Logger); ok {
			rlog.SetOutput(w)
		}
		h.ServeHTTP(w, r)
	})
}

func adminMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, _ := r.Context().Value("config").(*model.Config)
		rlog, _ := r.Context().Value("rlog").(*logrus.Logger)
		ew := writer.New(w)
		rlog.SetOutput(ew)
		groupUid := r.URL.Query().Get("group_uid")
		user, err := GetCurrentUser(r)
		if err != nil {
			if config.RunLevel == "debug" {
				SetJsonHeader(ew)
				w.WriteHeader(http.StatusUnauthorized)
				rlog.WithFields(logrus.Fields{
					"url": r.URL.String(),
				}).WithError(err).Error("current user couldn't be fetched")
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if groupRepo, ok := r.Context().Value("group_repository").(*groupRepository); ok {
			group, err := (*groupRepo).GetByUid(r.Context(), groupUid)
			if err != nil {
				if config.RunLevel == "debug" {
					w.WriteHeader(http.StatusInternalServerError)
					SetJsonHeader(ew)
					rlog.WithFields(logrus.Fields{
						"url": r.URL.String(),
					}).WithError(err).Error("error fetching group")
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if group.AdminID != user.ID {
				if config.RunLevel == "debug" {
					SetJsonHeader(ew)
					w.WriteHeader(http.StatusUnauthorized)
					rlog.WithFields(logrus.Fields{
						"url":   r.URL.String(),
						"user":  user.Name,
						"group": group.Title,
					}).Error("operation is not allowed")
				}
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(ew, r)
	})
}
