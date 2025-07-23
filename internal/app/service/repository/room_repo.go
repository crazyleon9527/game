package repository

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

var RoomRepositorySet = wire.NewSet(wire.Struct(new(RoomRepository), "*"))

type RoomRepository struct {
	DB *gorm.DB
}

// func (r *RoomRepository) CreateGameRoom(gameRoom *entities.GameRoom) error {
// 	return r.DB.Create(gameRoom).Error
// }

// // GetExpiredRooms 获取已过期但未关闭的房间列表
// func (r *RoomRepository) GetExpiredRooms() ([]entities.GameRoom, error) {
// 	var expiredRooms []entities.GameRoom
// 	err := r.DB.Where("expired_at < ? AND status <> ?", time.Now(), "closed").Find(&expiredRooms).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return expiredRooms, nil
// }

// // UpdateRoomStatus 更新房间的状态
// func (r *RoomRepository) UpdateRoomStatus(roomID uint, newStatus string) error {
// 	// 更新房间的状态为新的状态
// 	err := r.DB.Model(&entities.GameRoom{}).Where("id = ?", roomID).Update("status", newStatus).Error
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
