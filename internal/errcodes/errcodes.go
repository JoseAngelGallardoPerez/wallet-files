package errcodes

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
)

const (
	Forbidden                        = "FORBIDDEN"
	FileNotFound                     = "FILE_NOT_FOUND"
	CodeNotEnoughSpaceInFilesStorage = "NOT_ENOUGH_SPACE_IN_FILES_STORAGE"
	CodeFileTooLarge                 = "FILE_TOO_LARGE"
)

var StatusCodes = map[string]int{
	Forbidden:    http.StatusForbidden,
	FileNotFound: http.StatusNotFound,
}

func AddError(c *gin.Context, code string) {
	publicErr := &errors.PublicError{
		Code:       code,
		HttpStatus: StatusCodes[code],
	}
	errors.AddErrors(c, publicErr)
}

func AddErrorMeta(c *gin.Context, code string, meta interface{}) {
	publicErr := &errors.PublicError{
		Code:       code,
		HttpStatus: StatusCodes[code],
		Meta:       meta,
	}
	errors.AddErrors(c, publicErr)
}
