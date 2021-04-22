package policy

import (
	"github.com/Confialink/wallet-files/internal/srvdiscovery"
	"context"
	"log"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-permissions/rpc/permissions"
	"github.com/Confialink/wallet-pkg-acl"
	"github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-files/internal/service"
)

type Permission string

const (
	ViewUserProfiles    = Permission("view_user_profiles")
	ViewAdminProfiles   = Permission("view_admin_profiles")
	ModifyUserProfiles  = Permission("modify_user_profiles")
	ModifyAdminProfiles = Permission("modify_admin_profiles")
)

type Policy func(interface{}, *users.User) bool

type PermissionsService struct {
	usersService *service.Users
	logger       log15.Logger
}

func NewPermissionsService(usersService *service.Users, logger log15.Logger) *PermissionsService {
	return &PermissionsService{usersService, logger}
}

// CanClientReadFile checks if client can read a file
func (p *PermissionsService) CanClientReadFile(file interface{}, user *users.User) bool {
	f := file.(*database.FileModel)
	if f.IsAdminOnly == true {
		return false
	} else if f.IsPrivate == true && f.UserId != user.UID {
		return false
	}

	return true
}

// CanClientDeleteFile checks if client can delete a file
func (p *PermissionsService) CanClientDeleteFile(file interface{}, user *users.User) bool {
	f := file.(*database.FileModel)
	if f.UserId == user.UID && !f.IsAdminOnly {
		return true
	}

	return false
}

func (p *PermissionsService) CanAdminReadFile(file interface{}, user *users.User) bool {
	f := file.(*database.FileModel)
	if f.UserId == user.UID {
		return true
	}

	fileOwner, err := p.usersService.GetByUID(f.UserId)
	if err != nil {
		p.logger.New("method", "CanAdminReadFile", "err", err)
		return false
	}

	return p.CanAdminReadFiles(fileOwner, user)
}

func (p *PermissionsService) CanAdminReadFiles(filesOwner interface{}, user *users.User) bool {
	owner := filesOwner.(*users.User)
	if owner.UID == user.UID {
		return true
	}

	ownerRole := acl.RolesHelper.FromName(owner.RoleName)
	if ownerRole > acl.RolesHelper.FromName(user.RoleName) {
		return false
	}

	actionKey := ViewUserProfiles
	if ownerRole != acl.Client {
		actionKey = ViewAdminProfiles
	}

	return p.CheckPermission(actionKey, user)
}

func (p *PermissionsService) CanAdminDeleteFile(file interface{}, user *users.User) bool {
	f := file.(*database.FileModel)
	if f.UserId == user.UID {
		return true
	}

	fileOwner, err := p.usersService.GetByUID(f.UserId)
	if err != nil {
		p.logger.New("method", "CanAdminEditFile", "err", err)
		return false
	}

	return p.CanAdminUploadFiles(fileOwner, user)
}

func (p *PermissionsService) CanAdminUploadFiles(filesOwner interface{}, user *users.User) bool {
	owner := filesOwner.(*users.User)
	if owner.UID == user.UID {
		return true
	}

	ownerRole := acl.RolesHelper.FromName(owner.RoleName)
	if ownerRole > acl.RolesHelper.FromName(user.RoleName) {
		return false
	}

	actionKey := ModifyUserProfiles
	if ownerRole != acl.Client {
		actionKey = ModifyAdminProfiles
	}

	return p.CheckPermission(actionKey, user)
}

// CheckPermission calls permission service in order to check if user granted permission
func (p *PermissionsService) CheckPermission(permissionValue interface{}, user *users.User) bool {
	perm := permissionValue.(Permission)
	result, err := p.Check(user.UID, string(perm))
	if err != nil {
		log.Printf("permission policy failed to check permission: %s", err.Error())
		return false
	}
	return result
}

//Check checks if specified user is granted permission to perform some action
func (p *PermissionsService) Check(userId, actionKey string) (bool, error) {
	request := &permissions.PermissionReq{UserId: userId, ActionKey: actionKey}

	checker, err := p.checker()
	if nil != err {
		return false, err
	}

	response, err := checker.Check(context.Background(), request)
	if nil != err {
		return false, err
	}
	return response.IsAllowed, nil
}

func (p *PermissionsService) checker() (permissions.PermissionChecker, error) {
	permissionsUrl, err := srvdiscovery.ResolveRPC(srvdiscovery.ServiceNamePermissions)
	if nil != err {
		return nil, err
	}
	checker := permissions.NewPermissionCheckerProtobufClient(permissionsUrl.String(), http.DefaultClient)
	return checker, nil
}
