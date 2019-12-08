package handler

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
func (s *Handler) UploadFile(r *http.Request, maxSize int64, file string, fType int) (string, error) {
	var outDir string
	parseError := r.ParseMultipartForm(maxSize)
	if parseError != nil {
		return "", parseError
	}
	upFile, handler, err := r.FormFile(file)
	if err != nil {
		return "", err
	}
	defer upFile.Close()
	hash := sha256.New()
	contentType := handler.Header.Get("Content-Type")
	extension := "." + contentType[strings.LastIndex(contentType, "/")+1:]
	if fType == model.PostTypeImage {
		outDir = "./assets/images/"
	} else {
		outDir = "./assets/videos/"
	}
	rand.Seed(time.Now().UnixNano())
	hash.Write([]byte(strconv.FormatInt(int64(rand.Uint64()), 10)))
	sha := base64.URLEncoding.EncodeToString(hash.Sum(nil))
	outFile := outDir + sha[:len(sha)-1] + extension
	path := sha[:len(sha)-1] + extension
	s.log.WithFields(logrus.Fields{
		"file": path,
		"path": outDir,
	}).Infoln("writing uploaded file")
	f, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, upFile)

	return path, nil
}
