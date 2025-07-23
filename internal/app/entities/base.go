package entities

import (
	"database/sql/driver"
	"encoding/json"
	"rk-api/pkg/logger"
	"runtime"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// type Model struct {
// 	ID        uint `gorm:"primarykey"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt DeletedAt `gorm:"index"`
// }

// // // 自定义的基础模型
// type BaseModel struct {
//     ID        uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
//     CreatedAt time.Time      `json:"createdAt"`
//     UpdatedAt time.Time      `json:"updatedAt"`
//     DeletedAt gorm.DeletedAt `gorm:"index;" json:"-"`
// }

// // 自定义的基础模型
type BaseModel struct {
	ID        uint  `gorm:"primarykey" redis:"id"  json:"id"`
	CreatedAt int64 `json:"-"`
	UpdatedAt int64 `json:"-"`
	// DeletedAt gorm.DeletedAt `gorm:"index;" json:"-"`
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	// 在创建之前生成UUID作为主键
	now := time.Now().Unix()
	base.CreatedAt = now
	base.UpdatedAt = now
	return nil
}

// BeforeUpdate GORM钩子，在更新记录之前调用
func (base *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	// 更新UpdatedAt字段为当前时间
	base.UpdatedAt = time.Now().Unix()
	return nil
}

func (base *BaseModel) BeforeDelete(tx *gorm.DB) (err error) {
	// 设置DeletedAt字段为当前时间，以标记为“已删除”
	// now := time.Now()
	// base.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	return nil
}

// // // 自定义的基础模型
// type BaseModel2 struct {
// 	ID        uint           `gorm:"primarykey" json:"id"`
// 	CreatedAt int64          `json:"createdAt"`
// 	UpdatedAt int64          `json:"-"`
// 	DeletedAt gorm.DeletedAt `gorm:"index;" json:"-"`
// }

// // BeforeCreate GORM钩子，在创建记录之前调用
// func (base *BaseModel2) BeforeCreate(tx *gorm.DB) (err error) {
// 	// 在创建之前生成UUID作为主键
// 	now := time.Now().Unix()
// 	base.CreatedAt = now
// 	base.UpdatedAt = now
// 	return nil
// }

// // BeforeUpdate GORM钩子，在更新记录之前调用
// func (base *BaseModel2) BeforeUpdate(tx *gorm.DB) (err error) {
// 	// 更新UpdatedAt字段为当前时间
// 	base.UpdatedAt = time.Now().Unix()
// 	return nil
// }

// func (base *BaseModel2) BeforeDelete(tx *gorm.DB) (err error) {
// 	// 设置DeletedAt字段为当前时间，以标记为“已删除”
// 	now := time.Now()
// 	base.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
// 	return nil
// }

type KeyValues map[string]interface{}

func (p KeyValues) Value() (driver.Value, error) {
	// b, err := json.Marshal(p)
	return json.Marshal(p)
}

// Scan 实现方法
func (p *KeyValues) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &p)
}

func (p KeyValues) HasKey(key string) bool {
	_, ok := p[key]
	return ok
}

type IDReq struct {
	ID uint `json:"id"`
}

// ALTER TABLE your_table_name AUTO_INCREMENT = 80000; 这行命令将会把自增ID的起始值设置为80000，下一个插入的新记录的ID就会从80000开始。

func AddPrecise(base float64, val float64) float64 {
	decimalBase := decimal.NewFromFloat(base)
	decimalAdd := decimal.NewFromFloat(val)
	decimalResult := decimalBase.Add(decimalAdd).Round(3)
	result, _ := decimalResult.Float64() // Handle this error in production code

	if result < 0 {
		// Print the name of the function that called AddPrecise
		if pc, _, _, ok := runtime.Caller(1); ok {
			f := runtime.FuncForPC(pc)
			logger.ZError("AddPrecise less than zero", zap.String("call", f.Name()), zap.Float64("base", base), zap.Float64("val", val))
		}
	}

	// if result <= 0 {
	// 	result = 0.0000001 // Set to a minimum non-zero value to ensure DB update
	// }
	return result
}
