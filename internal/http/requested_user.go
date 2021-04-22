package http

import (
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-files/internal/errcodes"
	"github.com/Confialink/wallet-files/internal/service"
)

// put requested user to the Context
func RequestedUser(usersService *service.Users) gin.HandlerFunc {
	return func(c *gin.Context) {
		filesOwner, err := usersService.GetByUID(c.Params.ByName("uid"))
		if err != nil {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		c.Set("_requested_user", filesOwner)
	}
}
