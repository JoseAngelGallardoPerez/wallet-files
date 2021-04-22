package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Confialink/wallet-files/internal/config"
	"github.com/Confialink/wallet-files/internal/database"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3
type S3 struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	s3         *s3.S3
	config     config.AwsConfig
	repo       *database.Repository
}

func NewS3(
	uploader *s3manager.Uploader,
	downloader *s3manager.Downloader,
	s3 *s3.S3,
	config config.AwsConfig,
	repo *database.Repository,
) *S3 {
	return &S3{uploader, downloader, s3, config, repo}
}

// Upload uploads file to bucket and create record in database
func (s *S3) Upload(
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
	file.Read(b)
	file.Seek(0, 0)

	contentType := http.DetectContentType(b)
	if contentTypeRegexpValidator != nil && !contentTypeRegexpValidator.MatchString(contentType) {
		return nil, errors.New("content type is not allowed")
	}

	// retrieve extension from filename and remove dot from extension
	extDir := strings.Replace(filepath.Ext(header.Filename), ".", "", -1)
	if len(extDir) == 0 {
		extDir = "others"
	}

	path := extDir + "/" + time.Now().Format("2006-01-02")
	filename := strconv.FormatInt(time.Now().Unix(), 10) + "-" + header.Filename

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(s.config.S3Bucket),
		Key:                  aws.String(path + "/" + filename),
		Body:                 file,
		ContentType:          aws.String(contentType),
		ACL:                  aws.String("private"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		return nil, err
	}

	fileModel := database.FileModel{
		Filename:    filename,
		Path:        path,
		Size:        size,
		ContentType: contentType,
		Bucket:      s.config.S3Bucket,
		UserId:      userId,
		IsAdminOnly: isAdminOnly,
		IsPrivate:   isPrivate,
		Storage:     StorageS3,
	}

	createdFile, err := s.repo.Create(&fileModel)

	if err != nil {
		s.deleteFromS3(s.config.S3Bucket, path+"/"+filename)
		return nil, err
	}

	return createdFile, nil
}

// UploadBytes uploads bytes to bucket and create record in database
func (s *S3) UploadBytes(
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

	path := extDir + "/" + time.Now().Format("2006-01-02")
	filename := strconv.FormatInt(time.Now().Unix(), 10) + "-" + fileName

	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(s.config.S3Bucket),
		Key:                  aws.String(path + "/" + filename),
		Body:                 bytes.NewReader(b),
		ContentType:          aws.String(contentType),
		ACL:                  aws.String("private"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		return nil, err
	}

	fileModel := database.FileModel{
		Filename:    filename,
		Path:        path,
		Size:        int64(size),
		ContentType: contentType,
		Bucket:      s.config.S3Bucket,
		UserId:      userId,
		IsAdminOnly: isAdminOnly,
		IsPrivate:   isPrivate,
		Storage:     StorageS3,
		Category:    category,
	}

	createdFile, err := s.repo.Create(&fileModel)

	if err != nil {
		s.deleteFromS3(s.config.S3Bucket, path+"/"+filename)
		return nil, err
	}

	return createdFile, nil
}

// Delete deletes file from bucket and database
func (s *S3) Delete(file *database.FileModel) error {
	err := s.deleteFromS3(file.Bucket, file.Path+"/"+file.Filename)

	if err != nil {
		return err
	}

	err = s.repo.Delete(file)

	return err
}

func (s *S3) Download(file *database.FileModel) []byte {
	b := &aws.WriteAtBuffer{}
	s.downloader.Download(b, &s3.GetObjectInput{
		Bucket: aws.String(file.Bucket),
		Key:    aws.String(file.Path + "/" + file.Filename),
	})
	return b.Bytes()
}

// deleteFromS3 deletes file from bucket
func (s *S3) deleteFromS3(bucket string, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3.DeleteObject(input)

	return err
}
