package auth

import (
	"github.com/Confialink/wallet-files/internal/policy"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	goAcl "github.com/kildevaeld/go-acl"
)

const (
	FilesResource                = "private_files"
	FilesUploadPrivateResource   = "private_files_upload_private"
	FilesUploadPublicResource    = "private_files_upload_public"
	FilesUploadAdminOnlyResource = "private_files_upload_admin_only"

	CreateAction   = "create"
	UpdateAction   = "update"
	ReadAction     = "read"
	ReadListAction = "read_list"
	DeleteAction   = "delete"

	RoleRoot      = "root"
	RoleAdmin     = "admin"
	RoleClient    = "client"
)

// ServiceInterface
type ServiceInterface interface {
	Can(user *userpb.User, action string, resourceName string, resource interface{}) bool
}

type PermissionMap map[string]map[string]map[string]policy.Policy

// Service
type Service struct {
	Acl                *goAcl.ACL
	permissions        PermissionMap
	permissionsService *policy.PermissionsService
}

func NewService(acl *goAcl.ACL, permissionsService *policy.PermissionsService) *Service {
	auth := Service{Acl: acl, permissionsService: permissionsService}
	auth.permissions = PermissionMap{
		RoleClient: {
			FilesResource: {
				UpdateAction:   allowFunc,
				ReadAction:     auth.permissionsService.CanClientReadFile,
				ReadListAction: allowFunc,
				DeleteAction:   auth.permissionsService.CanClientDeleteFile,
			},
			FilesUploadPublicResource: {
				CreateAction: allowFunc,
			},
			FilesUploadPrivateResource: {
				CreateAction: allowFunc,
			},
		},
		RoleAdmin: {
			FilesResource: {
				ReadAction:     auth.permissionsService.CanAdminReadFile,
				ReadListAction: auth.permissionsService.CanAdminReadFiles,
				DeleteAction:   auth.permissionsService.CanAdminDeleteFile,
			},
			FilesUploadPublicResource: {
				CreateAction: auth.permissionsService.CanAdminUploadFiles,
			},
			FilesUploadPrivateResource: {
				CreateAction: auth.permissionsService.CanAdminUploadFiles,
			},
			FilesUploadAdminOnlyResource: {
				CreateAction: auth.permissionsService.CanAdminUploadFiles,
			},
		},
	}
	return &auth
}

// Can checks action is allowed
func (auth *Service) Can(user *userpb.User, action string, resourceName string, resource interface{}) bool {
	if user.RoleName == RoleRoot {
		return true
	}

	function := auth.getPermissionFunc(user.RoleName, action, resourceName)
	return function(resource, user)
}

// allowFunc always allows access
func allowFunc(_ interface{}, _ *userpb.User) bool {
	return true
}

// blockFunc always block access
func blockFunc(_ interface{}, _ *userpb.User) bool {
	return false
}

// getPermissionFunc returns function by role, action and resourceName.
// Returns blockFunc if proposed func not found
func (auth *Service) getPermissionFunc(role string, action string, resourceName string) policy.Policy {
	if rolePermission, ok := auth.permissions[role]; ok {
		if resourcePermission, ok := rolePermission[resourceName]; ok {
			if actionPermission, ok := resourcePermission[action]; ok {
				return actionPermission
			}
		}
	}
	return blockFunc
}
