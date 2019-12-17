package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"gitlab.com/innoserver/pkg/model"
)

// Uploads a file through http MultipartForm
//
// @param Request: the current request
// @param maxSize: the maximum file size
// @param file:    the multiparts file key which is the name of the uploaded file
//
// @param fType:   determine if filetype is wether image or video
func (s *Handler) UploadFile(r *http.Request, maxSize int64, file string, fType int) (string, string, error) {
	parseError := r.ParseMultipartForm(maxSize)
	if parseError != nil {
		return "", "", parseError
	}
	upFile, handler, err := r.FormFile(file)
	if err != nil {
		return "", "", err
	}
	defer upFile.Close()
	contentType := handler.Header.Get("Content-Type")
	extension := "." + contentType[strings.LastIndex(contentType, "/")+1:]
	outDir := initOutputDir(s, fType)
	fileName := generateFileName()

	s.log.WithFields(logrus.Fields{
		"file": fileName + extension,
		"path": outDir,
	}).Infoln("writing uploaded file")

	f, err := os.OpenFile(outDir+fileName+extension, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	io.Copy(f, upFile)
	if err != nil {
		return "", "", err
	}
	return fileName + extension, fileName + "_thumbnail" + extension, nil
}

func initOutputDir(s *Handler, fType int) string {
	if fType == model.PostTypeImage {
		return "." + s.config.ImagePath
	}
	return "." + s.config.VideoPath
}

func generateFileName() string {
	hash := sha256.New()
	hash.Write([]byte(strconv.FormatInt(int64(rand.Uint64()), 10)))
	sha := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	return sha[:len(sha)-1]
}

func initNewPostUpload(post *model.Post, s *Handler, r *http.Request) error {
	var err error
	post.Title = r.FormValue("title")
	if post.Title == "" {
		return errors.New("missing parameter title in request query")
	}
	parentUid := r.FormValue("parent_uid")
	post.Method, err = strconv.Atoi(r.FormValue("method"))
	if err != nil {
		return err
	}
	post.Type, err = strconv.Atoi(r.FormValue("type"))
	if err != nil {
		return err
	}
	gUid := r.URL.Query().Get("group_uid")
	if gUid != "" {
		if group, err := s.groupRepo.GetByUid(r.Context(), gUid); err == nil {
			post.GroupID.Int32 = int32(group.ID)
			post.GroupID.Valid = true
		}
	}
	parent, err := s.postRepo.GetByUid(r.Context(), parentUid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if parent != nil && parent.ID != 0 {
		post.ParentID.Int32 = int32(parent.ID)
		post.ParentID.Valid = true
	}
	return nil
}

func determineMaxPostSize(post *model.Post, s *Handler) int64 {
	if post.Type == model.PostTypeImage {
		return s.config.MaxImageSize
	}
	return s.config.MaxVideoSize
}
