package handlers

import "github.com/razponn/Resilient-Scatter-Gather/internal/clients"

type Handlers struct {
	users clients.UserService
	perms clients.PermissionsService
	vm    clients.VectorMemory
}

func New(users clients.UserService, perms clients.PermissionsService, vm clients.VectorMemory) *Handlers {
	return &Handlers{
		users: users,
		perms: perms,
		vm:    vm,
	}
}
