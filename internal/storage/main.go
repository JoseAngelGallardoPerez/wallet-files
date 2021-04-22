package storage

import (
	"mime/multipart"
	"regexp"

	"github.com/Confialink/wallet-files/internal/database"
)

const StorageS3 = "s3"
const StorageLocal = "local"

// Storage
type Storage interface {
	Upload(
		file multipart.File,
		header *multipart.FileHeader,
		userId string,
		isAdminOnly bool,
		isPrivate bool,
		contentTypeRegexpValidator *regexp.Regexp,
	) (*database.FileModel, error)
	UploadBytes(
		b []byte,
		fileName string,
		userId string,
		isAdminOnly bool,
		isPrivate bool,
		category *string,
	) (*database.FileModel, error)
	Delete(file *database.FileModel) error
	Download(file *database.FileModel) []byte
}
