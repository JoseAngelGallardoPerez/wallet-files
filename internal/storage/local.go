package storage

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Confialink/wallet-files/internal/database"
)

// Local
type Local struct {
	repo *database.Repository
}

const StorageDir = "files"

func NewLocal(
	repo *database.Repository,
) *Local {
	return &Local{repo}
}

// Upload uploads file to bucket and create record in database
func (s *Local) Upload(
	file multipart.File,
	header *multipart.FileHeader,
	userId string,
	isAdminOnly bool,
	isPrivate bool,
	contentTypeRegexpValidator *regexp.Regexp,
) (*database.FileModel, error) {
	defer file.Close()

	size := header.Size
	b := make([]byte, size)
	_, _ = file.Read(b)
	_, _ = file.Seek(0, 0)

	contentType := http.DetectContentType(b)
	if contentTypeRegexpValidator != nil && !contentTypeRegexpValidator.MatchString(contentType) {
		return nil, errors.New("content type is not allowed")
	}

	// retrieve extension from filename and remove dot from extension
	extDir := strings.Replace(filepath.Ext(header.Filename), ".", "", -1)
	if len(extDir) == 0 {
		extDir = "others"
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := StorageDir + "/" + extDir + "/" + time.Now().Format("2006-01-02")
	dirPath := wd + "/" + path
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}

	filename := strconv.FormatInt(time.Now().Unix(), 10) + "-" + header.Filename

	err = ioutil.WriteFile(dirPath+"/"+filename, b, 0644)
	if err != nil {
		return nil, err
	}

	fileModel := database.FileModel{
		Filename:    filename,
		Path:        path,
		Size:        size,
		ContentType: contentType,
		UserId:      userId,
		IsAdminOnly: isAdminOnly,
		IsPrivate:   isPrivate,
		Storage:     StorageLocal,
	}

	createdFile, err := s.repo.Create(&fileModel)

	if err != nil {
		_ = s.deleteFromLocalStorage(path, filename)
		return nil, err
	}

	return createdFile, nil
}

// Upload uploads file to bucket and create record in database
func (s *Local) UploadBytes(
	b []byte,
	fileName string,
	userId string,
	isAdminOnly bool,
	isPrivate bool,
	category *string,
) (*database.FileModel, error) {
	size := binary.Size(b)
	contentType := http.DetectContentType(b)

	// retrieve extension from filename and remove dot from extension
	extDir := strings.Replace(filepath.Ext(fileName), ".", "", -1)
	if len(extDir) == 0 {
		extDir = "others"
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := StorageDir + "/" + extDir + "/" + time.Now().Format("2006-01-02")
	dirPath := wd + "/" + path
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}
	filename := strconv.FormatInt(time.Now().Unix(), 10) + "-" + fileName

	err = ioutil.WriteFile(dirPath+"/"+filename, b, 0644)
	if err != nil {
		return nil, err
	}

	fileModel := database.FileModel{
		Filename:    filename,
		Path:        path,
		Size:        int64(size),
		ContentType: contentType,
		UserId:      userId,
		IsAdminOnly: isAdminOnly,
		IsPrivate:   isPrivate,
		Storage:     StorageLocal,
		Category:    category,
	}

	createdFile, err := s.repo.Create(&fileModel)

	if err != nil {
		_ = s.deleteFromLocalStorage(path, filename)
		return nil, err
	}

	return createdFile, nil
}

// Delete deletes file from bucket and database
func (s *Local) Delete(file *database.FileModel) error {
	err := s.deleteFromLocalStorage(file.Path, file.Filename)

	if err != nil {
		return err
	}

	err = s.repo.Delete(file)

	return err
}

func (s *Local) Download(file *database.FileModel) []byte {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}

	b, err := ioutil.ReadFile(wd + "/" + file.Path + "/" + file.Filename)
	if err != nil {
		return nil
	}

	return b
}

// deleteFromLocalStorage deletes file from local storage
func (s *Local) deleteFromLocalStorage(path string, filename string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return os.Remove(wd + "/" + path + "/" + filename)
}
