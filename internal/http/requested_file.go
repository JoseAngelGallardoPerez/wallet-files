package http

import (
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-files/internal/errcodes"
)

// put requested file to the Context
func RequestedFile(repo *database.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, typedErr := getIdParam(c)
		if typedErr != nil {
			errors.AddErrors(c, typedErr)
			return
		}

		file, err := repo.FindByID(id)
		if err != nil {
			errcodes.AddError(c, errcodes.FileNotFound)
			c.Abort()
			return
		}
		c.Set("_requested_file", file)
	}
}

// getIdParam returns id or nil
func getIdParam(c *gin.Context) (uint64, errors.TypedError) {
	id := c.Params.ByName("id")
	// convert string to uint
	id64, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		typedError := &errors.PublicError{Title: "id param must be an integer"}
		return 0, typedError
	}

	return uint64(id64), nil
}
