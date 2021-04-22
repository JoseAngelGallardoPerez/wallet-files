package http

import (
	"github.com/Confialink/wallet-files/internal/srvdiscovery"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-files/internal/auth"
	"github.com/Confialink/wallet-files/internal/errcodes"
	"github.com/Confialink/wallet-pkg-env_config"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

const TmpAuthHeader = "X-Tmp-Auth"

// CorsMiddleware cors middleware
func CorsMiddleware(config *env_config.Cors) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()

	corsConfig.AllowMethods = config.Methods
	for _, origin := range config.Origins {
		if origin == "*" {
			corsConfig.AllowAllOrigins = true
		}
	}
	if !corsConfig.AllowAllOrigins {
		corsConfig.AllowOrigins = config.Origins
	}
	corsConfig.AllowHeaders = config.Headers

	return cors.New(corsConfig)
}

func OwnerOrAdminOrRoot(ctx *gin.Context) {
	// Retrieve current user
	user, ok := ctx.Get("_user")
	if !ok {
		// Returns a "403 StatusForbidden" response
		errcodes.AddError(ctx, errcodes.Forbidden)
		ctx.Abort()
		return
	}

	if ctx.Params.ByName("uid") != user.(*userpb.User).UID {
		rolename := user.(*userpb.User).RoleName
		if rolename != "admin" && rolename != "root" {
			// Returns a "403 StatusForbidden" response
			errcodes.AddError(ctx, errcodes.Forbidden)
			ctx.Abort()
			return
		}
	}
}

// Authentication authentication middleware
func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := c.ServiceLogger()
		accessToken, ok := ExtractToken(ctx)
		if ok == true {
			client := userpb.NewUserHandlerProtobufClient(getRPCUserServerAddr(), &http.Client{})
			res, err := client.ValidateAccessToken(context.Background(), &userpb.Request{AccessToken: accessToken})

			if nil != err {
				logger.Error("can't validate token", "err", err)
			}

			if nil != res {
				ctx.Set("AccessToken", accessToken)
				ctx.Set("_user", res.User)
			}
		}
	}
}

func TmpAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := c.ServiceLogger()
		tmpAuthToken := ctx.Request.Header.Get(TmpAuthHeader)

		if tmpAuthToken == "" {
			ctx.Abort()
			ctx.Status(http.StatusUnauthorized)
			return
		}

		client := userpb.NewUserHandlerProtobufClient(getRPCUserServerAddr(), &http.Client{})
		res, err := client.ValidateTmpAuthToken(context.Background(), &userpb.Request{TmpAuthToken: tmpAuthToken})

		if err != nil {
			logger.Error("can't validate token", "err", err)
			ctx.Abort()
			ctx.Status(http.StatusUnauthorized)
			return
		}

		if res != nil {
			ctx.Set("TmpAuthHeader", tmpAuthToken)
			ctx.Set("_user", res.User)
		}
	}
}

// ExtractToken extracts jwt token from the header "Authorization" field with Bearer
func ExtractToken(c *gin.Context) (string, bool) {
	tokens := c.Request.Header.Get("Authorization")
	if len(tokens) < 8 || !strings.EqualFold(tokens[0:7], "Bearer ") {
		return "", false // empty token
	}

	return tokens[7:], true
}

func getRPCUserServerAddr() string {
	usersUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNameUsers)
	if nil != err {
		log.Fatalf(err.Error())
	}

	return usersUrl.String()
}

type PermissionChecker struct {
	authService auth.ServiceInterface
}

func NewMwPermissionChecker(authService auth.ServiceInterface) *PermissionChecker {
	return &PermissionChecker{authService}
}

// check dynamic permission to file resource
func (s *PermissionChecker) CanWithFile(action string) func(*gin.Context) {
	return func(c *gin.Context) {
		user, exist := c.Get("_user")
		if !exist {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		file, exist := c.Get("_requested_file")
		if !exist {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		if !s.authService.Can(user.(*userpb.User), action, auth.FilesResource, file) {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}
	}
}

// check dynamic permission to user's files
func (s *PermissionChecker) CanWithUser(action string, resourceName string) func(*gin.Context) {
	return func(c *gin.Context) {
		user, exist := c.Get("_user")
		if !exist {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		requestedUser, exist := c.Get("_requested_user")
		if !exist {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		if !s.authService.Can(user.(*userpb.User), action, resourceName, requestedUser) {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}
	}
}

// check dynamic permission to file resource
func (s *PermissionChecker) Can(action string, resourceName string) func(*gin.Context) {
	return func(c *gin.Context) {
		user, exist := c.Get("_user")
		if !exist {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}

		if !s.authService.Can(user.(*userpb.User), action, resourceName, nil) {
			errcodes.AddError(c, errcodes.Forbidden)
			c.Abort()
			return
		}
	}
}
