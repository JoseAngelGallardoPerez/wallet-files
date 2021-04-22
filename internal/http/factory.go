package http

import "github.com/Confialink/wallet-files/internal/di"

var c = di.Container
var Factory *factory

type factory struct{}

func init() {
	Factory = &factory{}
}

func (*factory) FilesHandlerFactory() *Handler {
	return NewHandler(
		c.Repository(),
		c.AuthService(),
		c.StorageService(),
		c.UsersService(),
		c.ServiceLogger(),
	)
}
