package handlers

import (
	"subian_go/internal/modules/rbac/contracts"
)

type RBACHandler struct {
	service contracts.RBACService
}

func NewRBACHandler(service contracts.RBACService) *RBACHandler {
	return &RBACHandler{service: service}
}
