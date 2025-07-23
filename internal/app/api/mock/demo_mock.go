package mock

import (
	"github.com/google/wire"
)

// DemoSet 注入Demo
var DemoSet = wire.NewSet(wire.Struct(new(Demo), "*"))

// Demo 示例程序
type Demo struct {
}

// // Query 查询数据
// // @Tags 菜单管理
// // @Summary 查询数据
// // @Security ApiKeyAuth
// // @Param current query int true "分页索引" default(1)
// // @Param pageSize query int true "分页大小" default(10)
// // @Param queryValue query string false "查询值"
// // @Param status query int false "状态(1:启用 2:禁用)"
// // @Param showStatus query int false "显示状态(1:显示 2:隐藏)"
// // @Param parentID query string false "父级ID"
// // @Success 200 {object} schema.ListResult{list=[]schema.Demo} "查询结果"
// // @Failure 401 {object} schema.ErrorResult "{error:{code:0,message:未授权}}"
// // @Failure 500 {object} schema.ErrorResult "{error:{code:0,message:服务器错误}}"
// // @Router /api/v1/Demos [get]
// func (a *Demo) Query(c *gin.Context) {
// }

// // QueryTree 查询菜单树
// // @Tags 菜单管理
// // @Summary 查询菜单树
// // @Security ApiKeyAuth
// // @Param status query int false "状态(1:启用 2:禁用)"
// // @Param parentID query string false "父级ID"
// // @Success 200 {object} schema.ListResult{list=[]schema.DemoTree} "查询结果"
// // @Failure 401 {object} schema.ErrorResult "{error:{code:0,message:未授权}}"
// // @Failure 500 {object} schema.ErrorResult "{error:{code:0,message:服务器错误}}"
// // @Router /api/v1/Demos.tree [get]
// func (a *Demo) QueryTree(c *gin.Context) {
// }

// // Get 查询指定数据
// // @Tags 菜单管理
// // @Summary 查询指定数据
// // @Security ApiKeyAuth
// // @Param id path string true "唯一标识"
// // @Success 200 {object} schema.Demo
// // @Failure 401 {object} schema.ErrorResult "{error:{code:0,message:未授权}}"
// // @Failure 404 {object} schema.ErrorResult "{error:{code:0,message:资源不存在}}"
// // @Failure 500 {object} schema.ErrorResult "{error:{code:0,message:服务器错误}}"
// // @Router /api/v1/Demos/{id} [get]
// func (a *Demo) Get(c *gin.Context) {
// }

// // Create 创建数据
// // @Tags 菜单管理
// // @Summary 创建数据
// // @Security ApiKeyAuth
// // @Param body body schema.Demo true "创建数据"
// // @Success 200 {object} schema.IDResult
// // @Failure 400 {object} schema.ErrorResult "{error:{code:0,message:无效的请求参数}}"
// // @Failure 401 {object} schema.ErrorResult "{error:{code:0,message:未授权}}"
// // @Failure 500 {object} schema.ErrorResult "{error:{code:0,message:服务器错误}}"
// // @Router /api/v1/Demos [post]
// func (a *Demo) Create(c *gin.Context) {
// }

// // Update 更新数据
// // @Tags 菜单管理
// // @Summary 更新数据
// // @Security ApiKeyAuth
// // @Param id path string true "唯一标识"
// // @Param body body schema.Demo true "更新数据"
// // @Success 200 {object} schema.StatusResult "{status:OK}"
// // @Failure 400 {object} schema.ErrorResult "{error:{code:0,message:无效的请求参数}}"
// // @Failure 401 {object} schema.ErrorResult "{error:{code:0,message:未授权}}"
// // @Failure 500 {object} schema.ErrorResult "{error:{code:0,message:服务器错误}}"
// // @Router /api/v1/Demos/{id} [put]
// func (a *Demo) Update(c *gin.Context) {
// }
