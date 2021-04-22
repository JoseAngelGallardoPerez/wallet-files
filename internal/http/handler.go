package http

import (
	"bytes"
	"net/http"
	"regexp"
	"strconv"

	list_params "github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-files/internal/auth"
	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-files/internal/errcodes"
	"github.com/Confialink/wallet-files/internal/service"
	errors "github.com/Confialink/wallet-pkg-errors"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
)

// Handler
type Handler struct {
	repo                *database.Repository
	authService         auth.ServiceInterface
	storageService      *service.StorageService
	userService         *service.Users
	logger              log15.Logger
}

// NewHandler creates new handler instance
func NewHandler(
	repo *database.Repository,
	authService auth.ServiceInterface,
	storageService *service.StorageService,
	userService *service.Users,
	logger log15.Logger,
) *Handler {
	return &Handler{
		repo,
		authService,
		storageService,
		userService,
		logger,
	}
}

// HealthCheckHandler is ping test
func (h Handler) HealthCheckHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

// OptionsHandler handle options request
func (h Handler) OptionsHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *Handler) GetBinary(c *gin.Context) {
	logger := h.logger.New("action", "GetBinary")
	id := h.getIdParam(c)
	currentUser := h.getCurrentUser(c)

	file := h.getRequestedFile(c)
	if file == nil {
		logger.Error("file not found", "id", id)
		errcodes.AddError(c, errcodes.FileNotFound)
		return
	}

	if (file.IsPrivate && nil == currentUser) ||
		((file.IsPrivate || file.IsAdminOnly) && !h.authService.Can(
			currentUser,
			auth.ReadAction,
			auth.FilesResource,
			file,
		)) {
		logger.Error("forbidden", "id", id,
			"file is private", file.IsPrivate, "file is admin only", file.IsAdminOnly, "current user", currentUser)
		errcodes.AddError(c, errcodes.Forbidden)
		return
	}

	b := h.storageService.Download(file)
	r := bytes.NewReader(b)

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="` + file.Filename + `"`,
	}

	c.DataFromReader(http.StatusOK, file.Size, file.ContentType, r, extraHeaders)
}

// GetHandler returns file by id
func (h *Handler) GetHandler(c *gin.Context) {
	file := h.getRequestedFile(c)
	if file == nil {
		logger := h.logger.New("action", "GetHandler")
		logger.Error("not found", "id", h.getIdParam(c))
		errcodes.AddError(c, errcodes.FileNotFound)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(file))
}

// CreatePublicHandler creates new public file
func (h *Handler) CreatePublicHandler(c *gin.Context) {
	uid := c.Params.ByName("uid")
	file, header, err := c.Request.FormFile("file")

	if nil != err {
		privateError := errors.PrivateError{Message: "can't read file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	res, tErr := h.storageService.Upload(file, header, uid, false, false, nil)

	if nil != tErr {
		errors.AddErrors(c, tErr)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(res))

}

// CreatePrivateHandler creates new private file
func (h *Handler) CreatePrivateHandler(c *gin.Context) {
	uid := c.Params.ByName("uid")
	file, header, err := c.Request.FormFile("file")

	if nil != err {
		privateError := errors.PrivateError{Message: "can't read file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	res, tErr := h.storageService.Upload(file, header, uid, false, true, nil)

	if nil != tErr {
		errors.AddErrors(c, tErr)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(res))

}

// CreatePrivateLimitedHandler creates new private file for the new user
func (h *Handler) CreatePrivateLimitedHandler(c *gin.Context) {
	currentUser := h.mustGetCurrentUser(c)

	file, header, err := c.Request.FormFile("file")
	if nil != err {
		privateError := errors.PrivateError{Message: "can't read file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	res, tErr := h.storageService.Upload(file, header, currentUser.UID, false, true, nil)
	if nil != tErr {
		errors.AddErrors(c, tErr)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(res))
}

// CreateAdminOnlyHandler creates new private file visible for admin only
func (h *Handler) CreateAdminOnlyHandler(c *gin.Context) {
	uid := c.Params.ByName("uid")
	file, header, err := c.Request.FormFile("file")

	if nil != err {
		privateError := errors.PrivateError{Message: "can't read file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	res, tErr := h.storageService.Upload(file, header, uid, true, true, nil)

	if nil != tErr {
		errors.AddErrors(c, tErr)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(res))

}

// DeleteHandler deletes an existing file
func (h *Handler) DeleteHandler(c *gin.Context) {
	logger := h.logger.New("action", "DeleteHandler")
	file := h.getRequestedFile(c)
	if file == nil {
		logger.Error("not found", "id", h.getIdParam(c))
		errcodes.AddError(c, errcodes.FileNotFound)
		return
	}

	err := h.storageService.Delete(file)

	if nil != err {
		privateError := errors.PrivateError{Message: "can't delete a file"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("id", h.getIdParam(c))
		errors.AddErrors(c, &privateError)
		return
	}

	c.Status(http.StatusOK)
}

// GetUserFilesHandler returns list of files
func (h *Handler) GetUserFilesHandler(c *gin.Context) {
	uid := c.Params.ByName("uid")
	currentUser := h.mustGetCurrentUser(c)

	params := h.getListParamsByRoleName(currentUser.RoleName, c.Request.URL.RawQuery)
	params.AddFilter("user_id", []string{uid})
	files, err := h.repo.GetList(params)
	if nil != err {
		privateError := errors.PrivateError{Message: "can't retrieve files"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	list, err := NewResponseList(files)
	if nil != err {
		privateError := errors.PrivateError{Message: "can't create response list"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, NewResponse().SetData(list))
}

func (h *Handler) CreateProfileImageHandler(c *gin.Context) {
	currentUser := h.mustGetCurrentUser(c)

	file, header, err := c.Request.FormFile("file")

	if nil != err {
		privateError := errors.PrivateError{Message: "can't read file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	res, tErr := h.storageService.Upload(
		file,
		header,
		currentUser.UID,
		false,
		false,
		regexp.MustCompile(`(image)/([a-z-.]{2,})`),
	)
	if nil != tErr {
		errors.AddErrors(c, tErr)
		return
	}

	err = h.userService.UpdateProfileImageID(currentUser.UID, res.ID)
	if err != nil {
		err2 := h.storageService.Delete(res)
		if err2 != nil {
			privateError := errors.PrivateError{Message: "can't delete recently uploaded image"}
			privateError.AddLogPair("error", err2.Error())
			errors.AddErrors(c, &privateError)
			return
		}

		privateError := errors.PrivateError{Message: "can't update profile image id"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	if currentUser.ProfileImageID != 0 {
		currentImage, err := h.repo.GetByID(currentUser.ProfileImageID)
		if nil != err {
			privateError := errors.PrivateError{Message: "can't find current profile image"}
			privateError.AddLogPair("error", err.Error())
			errors.AddErrors(c, &privateError)
			return
		}

		err = h.storageService.Delete(currentImage)
		if nil != err {
			privateError := errors.PrivateError{Message: "can't delete current profile image"}
			privateError.AddLogPair("error", err.Error())
			errors.AddErrors(c, &privateError)
			return
		}
	}

	c.JSON(http.StatusOK, NewResponse().SetData(res))
}

// NotFoundHandler returns 404 NotFound
func (h *Handler) NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "file": "Page not found"})
}

// getIdParam returns id or nil
func (h *Handler) getIdParam(c *gin.Context) uint64 {
	id := c.Params.ByName("id")

	// convert string to uint
	id64, err := strconv.ParseUint(id, 10, 32)

	if err != nil {
		c.JSON(http.StatusBadRequest, NewResponseWithError(
			"id param must be an integer",
			err.Error(),
			http.StatusBadRequest,
			nil,
		))
		return 0
	}

	return uint64(id64)
}

// getCurrentUser returns current user or nil
func (h *Handler) getCurrentUser(c *gin.Context) *userpb.User {
	user, exist := c.Get("_user")
	if !exist {
		return nil
	}
	return user.(*userpb.User)
}

// mustGetCurrentUser returns current user or throw error
func (h *Handler) mustGetCurrentUser(c *gin.Context) *userpb.User {
	user := h.getCurrentUser(c)
	if nil == user {
		panic("user must be set")
	}
	return user
}

// getRequestedFile returns requested file
func (h *Handler) getRequestedFile(c *gin.Context) *database.FileModel {
	file, exist := c.Get("_requested_file")
	if !exist {
		return nil
	}
	return file.(*database.FileModel)
}

func (h *Handler) getListParamsByRoleName(roleName string, query string) *list_params.ListParams {
	params := getListParams(query)
	if roleName != auth.RoleRoot && roleName != auth.RoleAdmin {
		params.AddFilter("is_admin_only", []string{"false"})
	}
	return params
}
