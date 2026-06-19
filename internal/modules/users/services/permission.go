package services

import (
	rbacMiddlewares "subian_go/internal/modules/rbac/middlewares"
	rbacModels "subian_go/internal/modules/rbac/models"
	he "subian_go/internal/shared/httputil"
)

func (s *service) canCreateUser(actor he.AuthContext) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersCreate); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersManage); err != nil || has {
		return has, err
	}
	return rbacMiddlewares.HasAnyRole(s.rbacRepo, actor.UserID, "admin", "superadmin", "hrd")
}

func (s *service) canUpdateUser(actor he.AuthContext, targetUserID int64) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if actor.UserID == targetUserID {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersUpdate); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersManage); err != nil || has {
		return has, err
	}
	return rbacMiddlewares.HasAnyRole(s.rbacRepo, actor.UserID, "admin", "superadmin", "hrd")
}

func (s *service) canDeleteUser(actor he.AuthContext) (bool, error) {
	return actor.IsSuperadmin, nil
}

func (s *service) canReadUser(actor he.AuthContext, targetUserID int64) (bool, error) {
	if actor.IsSuperadmin {
		return true, nil
	}
	if actor.UserID == targetUserID {
		return true, nil
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersRead); err != nil || has {
		return has, err
	}
	if has, err := rbacMiddlewares.HasPermission(s.rbacRepo, actor.UserID, rbacModels.PermUsersManage); err != nil || has {
		return has, err
	}
	return false, nil
}
