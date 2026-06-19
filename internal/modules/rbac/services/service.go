package services

import (
	"subian_go/internal/modules/rbac/contracts"

	userContracts "subian_go/internal/modules/users/contracts"
)

type rbacService struct {
	rbacRepo contracts.RBACRepository
	userRepo userContracts.Repository
}

func NewRBACService(rbacRepo contracts.RBACRepository, userRepo userContracts.Repository) contracts.RBACService {
	return &rbacService{rbacRepo: rbacRepo, userRepo: userRepo}
}
