package mock

import (
	// "rk-api/internal/app/entities"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// var req entities.IDReq

// ActivitySet 注入Activity
var ActivitySet = wire.NewSet(wire.Struct(new(Activity), "*"))

// Activity 示例程序
type Activity struct {
}

// @Summary 获取活动列表信息
// @Description 获取活动列表信息
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.Activity "活动信息"
// @Router /api/activity/get-activity-list [post]
func (c *Activity) GetActivityList(ctx *gin.Context) {

}

// @Summary 获取banner列表信息
// @Description 获取banner列表信息
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.Banner "banner信息"
// @Router /api/activity/get-banner-list [post]
func (c *Activity) GetBannerList(ctx *gin.Context) {

}

// @Summary 获取logo图片信息
// @Description 获取logo图片信息
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param logoType query int false "logo类型 (1.平台主界面左上角 2.平台跳转游戏加载页 3.游戏加载页)" default(1)
// @Success 200 {array} entities.Logo "logo信息"
// @Router /api/activity/get-logo-list [post]
func (c *Activity) GetLogoList(ctx *gin.Context) {

}
