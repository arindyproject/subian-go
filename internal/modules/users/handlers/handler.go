package handlers

import (
	userContracts "subian_go/internal/modules/users/contracts"
)

type Handler struct {
	service userContracts.Service
}

func NewHandler(service userContracts.Service) *Handler {
	return &Handler{service: service}
}
