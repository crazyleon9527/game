package service

import (
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

var RoomServiceSet = wire.NewSet(
	ProvideRoomService,
)

type RoomService struct {
	Repo    *repository.RoomRepository
	UserSrv *UserService
}

func ProvideRoomService(repo *repository.RoomRepository,
	userSrv *UserService,
) *RoomService {
	return &RoomService{
		Repo:    repo,
		UserSrv: userSrv,
	}
}

var RoomStatusWaiting = "waiting"
var RoomStatusRunning = "running"
var RoomStatusClosed = "closed"
