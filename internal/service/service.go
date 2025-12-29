package service

import (
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

type Service interface {
	Users() UserSrv
	Auth() AuthSrv
	WG() WGSrv
}

type service struct {
	store store.Factory
}

func NewService(store store.Factory) Service {
	return &service{store: store}
}

func (s *service) Users() UserSrv {
	return newUsers(s)
}

func (s *service) Auth() AuthSrv {
	return newAuth(s)
}

func (s *service) WG() WGSrv {
	return newWG(s)
}
