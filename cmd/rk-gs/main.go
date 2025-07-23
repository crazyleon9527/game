package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"rk-api/internal/app/middleware"

	"github.com/gin-gonic/gin"
)

func main() {

	port := flag.String("port", "8080", "指定服务器端口")
	wwwRootParam := flag.String("wwwroot", "./ddll", "指定 wwwroot 路径") // 默认值 ddll
	flag.Parse()

	// 创建一个Gin路由
	r := gin.Default()

	// 使用自定义中间件
	r.NoRoute(middleware.UnityAutoFileHandler(*wwwRootParam))

	// 使用自定义中间件

	// 设置一个上传文件的路由
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	// 启动服务，使用通过 flag 获取的端口
	err := r.Run(":" + *port)
	if err != nil {
		fmt.Println("启动服务器失败:", err)
		os.Exit(1)
	}
}
