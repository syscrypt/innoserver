package handler

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	_ "gitlab.com/innoserver/pkg/model"
)

// Uploads a file through http MultipartForm
//
// @param Request: the current request
// @param maxSize: the maximum file size
// @param file:    the multiparts file key which is the name of the uploaded
//                 file
// @param fType:   determine if filetype is wether image or video
func (s *Handler) UploadFile(r *http.Request, maxSize int64, file string, fType string) error {
	var outDir string
	logrus.Infoln("file upload initialized")
	r.ParseMultipartForm(maxSize)
	upFile, handler, err := r.FormFile(file)
	if err != nil {
		return err
	}
	defer upFile.Close()
	logrus.Println("Fileheader:", handler.Header)
	logrus.Println("Filesize:", handler.Size)
	hash := sha256.New()

	if fType == "image" {
		outDir = "./assets/images"
	} else {
		outDir = "./assets/videos"
	}

	io.WriteString(hash, handler.Filename+strconv.Itoa(rand.Int()))
	temp, err := ioutil.TempFile("", "upload")
	if err != nil {
		return err
	}
	defer temp.Close()

	f, err := os.OpenFile(outDir+string(hash.Sum(nil)), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, temp)

	return nil
}
