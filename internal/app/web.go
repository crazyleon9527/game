package internal

import (
	"net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"rk-api/internal/app/config"
	"rk-api/internal/app/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "rk-api/internal/app/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitGinEngine(r router.IRouter) *gin.Engine {
	engine := gin.Default()

	// 设置上传文件大小限制
	engine.MaxMultipartMemory = 8 << 20 // 8 MiB

	// 创建自定义的写入器  // 增加缓冲区大小为 64KB
	// bufWriter := bufio.NewWriterSize(gin.DefaultWriter, 64*1024*100)
	// // 将自定义的写入器设置为 Gin 的默认写入器
	// gin.DefaultWriter = bufWriter

	if config.Get().ServiceSettings.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// engine.NoRoute(middleware.UnityAutoFileHandler)

	// 让 public 目录可以访问，并默认访问 index.html
	// r.StaticFS("/", gin.Dir(config.Get().LandingSettings.StaticPath, true))
	// engine.Use(MultiStatic([]string{config.Get().ServiceSettings.StaticPath}))
	engine.Static("/public", filepath.Join(config.Get().ServiceSettings.StaticPath))

	// 健康监测
	engine.GET("/", func(c *gin.Context) {
		c.Status(200)
	})
	engine.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	// 路由处理静态HTML页面
	// engine.Static("/unity", "./public/unity")

	// // 路由处理静态HTML页面
	// engine.GET("/game", func(c *gin.Context) {
	// 	c.File("./public/unity/index.html") // 直接提供index.html文件
	// })

	// r.NoRoute(gin.WrapH(http.FileServer(gin.Dir(config.Get().LandingSettings.StaticPath, true))))

	if config.Get().ServiceSettings.EnablePprof {
		pprofRouter(engine) //注册 性能分析
	}

	if config.Get().ServiceSettings.EnableSwagger {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) //swagger文档

		// 动态设置 Swagger 的 host 和 schemes

	}

	applyCors(engine)
	applyRoute(engine, r)

	return engine
}

func MultiStatic(rootFolders []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, root := range rootFolders {
			filePath := root + c.Request.URL.Path
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}
		}
		c.Next()
	}
}

// pprofRouter wrapper function to initialize pprof routes
func pprofRouter(router *gin.Engine) {

	pprofGroup := router.Group("/debug/pprof")

	// 注册pprof的主路由
	pprofGroup.GET("/", gin.WrapF(pprof.Index))

	// 注册其他需要的路由
	pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
	pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
	pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
	pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
	pprofGroup.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
	pprofGroup.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
	pprofGroup.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
	pprofGroup.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
	pprofGroup.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
	pprofGroup.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
}

func applyRoute(engine *gin.Engine, r router.IRouter) {
	r.Register(engine)
}

func applyCors(r *gin.Engine) {
	// 配置跨域
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true                                                                             // 允许所有域名访问
	corsConfig.AllowCredentials = true                                                                            // 允许跨域请求携带 Cookie
	corsConfig.AddAllowHeaders("Authorization", "X-User", "Accept-Language", "X-Access-Token", "X-Refresh-Token") // 允许的请求头
	corsConfig.AddExposeHeaders("X-Access-Token", "X-Refresh-Token")                                              // 允许前端访问的响应头
	corsConfig.MaxAge = 12 * time.Hour                                                                            // 预检请求缓存时间

	// 应用 CORS 中间件
	r.Use(cors.New(corsConfig))
}
