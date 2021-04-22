package service

import (
	"encoding/binary"
	"errors"
	"mime/multipart"
	"net/http"
	"regexp"

	"github.com/Confialink/wallet-files/internal/service/syssettings"

	"github.com/Confialink/wallet-files/internal/config"
	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-files/internal/errcodes"
	"github.com/Confialink/wallet-files/internal/storage"
	errorsPkg "github.com/Confialink/wallet-pkg-errors"
)

const MaxStorageSizePerUserBytes = 5e+7
const MaxFileSizeBytes = 5e+6

type StorageService struct {
	pool       map[string]storage.Storage
	config     *config.Config
	repository *database.Repository
}

func NewStorageService(
	pool map[string]storage.Storage,
	config *config.Config,
	repository *database.Repository,
) *StorageService {
	return &StorageService{
		pool:       pool,
		config:     config,
		repository: repository,
	}
}

func (s *StorageService) Upload(
	file multipart.File,
	header *multipart.FileHeader,
	userId string,
	isAdminOnly bool,
	isPrivate bool,
	contentTypeRegexpValidator *regexp.Regexp,
) (*database.FileModel, errorsPkg.TypedError) {
	st, ok := s.pool[s.config.Storage]
	if !ok {
		return nil, &errorsPkg.PrivateError{Message: "can't find storage"}
	}

	if isPrivate || isAdminOnly {
		totalSize, err := s.repository.GetTotalSizeOfUserFiles(userId)
		if err != nil {
			pErr := &errorsPkg.PrivateError{Message: "can't get total size of user files"}
			pErr.AddLogPair("err", err)
			return nil, pErr
		}

		limits, err := syssettings.GetUserFilesStorageLimits()
		if err != nil {
			pErr := &errorsPkg.PrivateError{Message: "can't get storage limits from settings service"}
			pErr.AddLogPair("err", err)
			return nil, pErr
		}

		if header.Size > limits.FileSizeLimitBytes {
			return nil, &errorsPkg.PublicError{
				Title:      "File is too large",
				Code:       errcodes.CodeFileTooLarge,
				HttpStatus: http.StatusRequestEntityTooLarge,
			}
		}

		if totalSize+float64(header.Size) > float64(limits.TotalLimitBytes) {
			return nil, &errorsPkg.PublicError{
				Title:      "Not enough space in your files storage",
				Code:       errcodes.CodeNotEnoughSpaceInFilesStorage,
				HttpStatus: http.StatusBadRequest,
			}
		}
	}

	res, err := st.Upload(file, header, userId, isAdminOnly, isPrivate, contentTypeRegexpValidator)
	if err != nil {
		pErr := &errorsPkg.PrivateError{Message: "can't upload file"}
		pErr.AddLogPair("err", err)

		return nil, pErr
	}

	return res, nil
}

func (s *StorageService) UploadBytes(
	bytes []byte,
	fileName string,
	userId string,
	isAdminOnly bool,
	isPrivate bool,
	category *string,
) (*database.FileModel, errorsPkg.TypedError) {
	st, ok := s.pool[s.config.Storage]
	if !ok {
		return nil, &errorsPkg.PrivateError{Message: "can't find storage"}
	}

	size := binary.Size(bytes)
	if isPrivate || isAdminOnly {
		totalSize, err := s.repository.GetTotalSizeOfUserFiles(userId)
		if err != nil {
			pErr := &errorsPkg.PrivateError{Message: "can't get total size of user files"}
			pErr.AddLogPair("err", err)
			return nil, pErr
		}

		limits, err := syssettings.GetUserFilesStorageLimits()
		if err != nil {
			pErr := &errorsPkg.PrivateError{Message: "can't get storage limits from settings service"}
			pErr.AddLogPair("err", err)
			return nil, pErr
		}

		if int64(size) > limits.FileSizeLimitBytes {
			return nil, &errorsPkg.PublicError{
				Title:      "File is too large",
				Code:       errcodes.CodeFileTooLarge,
				HttpStatus: http.StatusRequestEntityTooLarge,
			}
		}

		if totalSize+float64(int64(size)) > float64(limits.TotalLimitBytes) {
			return nil, &errorsPkg.PublicError{
				Title:      "Not enough space in your files storage",
				Code:       errcodes.CodeNotEnoughSpaceInFilesStorage,
				HttpStatus: http.StatusBadRequest,
			}
		}
	}

	res, err := st.UploadBytes(bytes, fileName, userId, isAdminOnly, isPrivate, category)
	if err != nil {
		pErr := &errorsPkg.PrivateError{Message: "can't upload bytes"}
		pErr.AddLogPair("err", err)

		return nil, pErr
	}

	return res, nil
}

// Delete deletes file from bucket and database
func (s *StorageService) Delete(file *database.FileModel) error {
	st, ok := s.pool[file.Storage]
	if !ok {
		return errors.New("storage not found")
	}

	return st.Delete(file)
}

func (s *StorageService) Download(file *database.FileModel) []byte {
	st, ok := s.pool[file.Storage]
	if !ok {
		return nil
	}

	return st.Download(file)
}
