package routes

import (
	"github.com/Confialink/wallet-files/internal/auth"
	"github.com/Confialink/wallet-files/internal/authentication"
	"github.com/Confialink/wallet-files/internal/di"
	"github.com/Confialink/wallet-files/internal/http"
	"github.com/Confialink/wallet-files/internal/version"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-service_names"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine
var c = di.Container

func initRoutes() {
	r = gin.New()

	fileHandler := http.Factory.FilesHandlerFactory()

	r.Use(
		http.CorsMiddleware(c.Config().Cors),
	)

	apiGroup := r.Group(service_names.Files.Internal)
	apiGroup.Use(errors.ErrorHandler(c.ServiceLogger().New("Middleware", "Errors")))

	apiGroup.GET("/health-check", fileHandler.HealthCheckHandler)

	r.GET("/files/build", func(c *gin.Context) {
		c.JSON(200, version.BuildInfo)
	})

	mwRequestedFile := http.RequestedFile(c.Repository())

	privateGroup := apiGroup.Group("/private")
	{
		permChecker := http.NewMwPermissionChecker(c.AuthService())

		v1Group := privateGroup.Group("/v1", authentication.Middleware(c.ServiceLogger().New("Middleware", "Authentication")))
		{
			mwRequestedUser := http.RequestedUser(c.UsersService())
			v1Group.GET("/files/:id", mwRequestedFile, permChecker.CanWithFile(auth.ReadAction), fileHandler.GetHandler)
			v1Group.DELETE("/files/:id", mwRequestedFile, permChecker.CanWithFile(auth.DeleteAction), fileHandler.DeleteHandler)
			v1Group.POST("/files/public/:uid", mwRequestedUser, http.OwnerOrAdminOrRoot, permChecker.CanWithUser(auth.CreateAction, auth.FilesUploadPublicResource), fileHandler.CreatePublicHandler)
			v1Group.POST("/files/private/:uid", mwRequestedUser, http.OwnerOrAdminOrRoot, permChecker.CanWithUser(auth.CreateAction, auth.FilesUploadPrivateResource), fileHandler.CreatePrivateHandler)
			v1Group.POST("/files/admin-only/:uid", mwRequestedUser, http.OwnerOrAdminOrRoot, permChecker.CanWithUser(auth.CreateAction, auth.FilesUploadPrivateResource), fileHandler.CreateAdminOnlyHandler)
			v1Group.POST("/files/profile-image", fileHandler.CreateProfileImageHandler)

			usersGroup := v1Group.Group("/users")
			{

				usersGroup.GET("/:uid", mwRequestedUser, http.OwnerOrAdminOrRoot, permChecker.CanWithUser(auth.ReadListAction, auth.FilesResource), fileHandler.GetUserFilesHandler)
			}

			storageGroup := v1Group.Group("storage")
			{
				binGroup := storageGroup.Group("/bin")
				{
					binGroup.GET("/:id", mwRequestedFile, fileHandler.GetBinary)
				}
			}
		}

		// limited routes may be accessed using temporary jwt tokens
		v1Limited := privateGroup.Group("/v1/limited", http.TmpAuthentication())
		{
			v1Limited.GET(":id", mwRequestedFile, permChecker.CanWithFile(auth.ReadAction), fileHandler.GetHandler)
			v1Limited.POST("private", fileHandler.CreatePrivateLimitedHandler)
			v1Limited.DELETE(":id", mwRequestedFile, permChecker.CanWithFile(auth.DeleteAction), fileHandler.DeleteHandler)
		}
	}

	publicGroup := apiGroup.Group("/public")
	{
		v1Group := publicGroup.Group("/v1", http.Authentication())
		{
			storageGroup := v1Group.Group("storage")
			{
				binGroup := storageGroup.Group("/bin")
				{
					binGroup.GET("/:id", mwRequestedFile, fileHandler.GetBinary)
				}
			}
		}
	}

	// Handle OPTIONS request
	r.OPTIONS("/*cors", fileHandler.OptionsHandler)
}

func GetRouter() *gin.Engine {
	if nil == r {
		initRoutes()
	}

	return r
}
