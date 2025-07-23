package main

import (
	"context"
	"os"
	app "rk-api/internal/app"

	"log"

	"github.com/urfave/cli/v2"
)

// VERSION 版本号，可以通过编译的方式指定版本号：go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "1.0.0"

// Swagger 文档规则请参考：https://github.com/swaggo/swag#declarative-comments-format
// @title rk-api
// @version 1.0.0
// @description rk-api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @schemes http https
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host
//	@BasePath	/

func main() {
	// ctx := logger.NewTagContext(context.Background(), "__main__")

	ctx := context.Background()
	app := cli.NewApp()
	app.Name = "rk-api"
	app.Version = VERSION
	app.Usage = "turbos landing system"
	app.Commands = []*cli.Command{
		newWebCmd(ctx),
	}
	err := app.Run(os.Args)
	if err != nil {
		// logger.WithContext(ctx).Errorf(err.Error())

		log.Fatalf("Application error: %v", err)
	}
}

func newWebCmd(ctx context.Context) *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Run http server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Usage:   "App configuration file(.yaml),like: configs/config.yaml",
				Value:   "configs/config-linux.yaml",
			},
		},
		Action: func(c *cli.Context) error {
			return app.Run(ctx,
				app.SetConfigFile(c.String("conf")),
				app.SetVersion(VERSION))
		},
	}
}

// import (
// 	// "rk-api/utils"

// 	"fmt"
// 	"os"

// 	"github.com/urfave/cli/v2"
// )

// func main() {
// 	// 创建一个命令行应用
// 	app := cli.NewApp()
// 	app.Name = "turbos-langding"
// 	app.Usage = "turbos luodiye"
// 	// 添加命令行启动参数
// 	app.Flags = []cli.Flag{
// 		&cli.StringFlag{
// 			Name:    "conf",
// 			Aliases: []string{"c"},
// 			Usage:   "App configuration file(.yaml),like: configs/config.yaml",
// 			Value:   "configs/config-linux.yaml",
// 		},
// 	}

// 	// 定义命令行操作
// 	app.Action = func(c *cli.Context) (err error) {
// 		// 获取命令行启动参数
// 		conf := c.String("conf")
// 		startWithConf(conf)

// 		return
// 	}

// 	// 运行命令行应用
// 	err := app.Run(os.Args)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

// // func startWithConf(conf string) {

// // 	config.MustLoad(conf)
// // 	logger := logger.GetLogger()
// // 	defer logger.Sync()
// // 	zap.L().Info("load", zap.Any("config", config.Get()), zap.String("file", conf))

// // 	r := gin.Default()
// // 	// gin.SetMode(gin.ReleaseMode)

// // 	// 让 public 目录可以访问，并默认访问 index.html
// // 	// r.StaticFS("/", gin.Dir(config.Get().LandingSettings.StaticPath, true))

// // 	r.Use(MultiStatic([]string{config.Get().LandingSettings.StaticPath, config.Get().LandingSettings.DomainRootPath}))

// // 	// r.NoRoute(gin.WrapH(http.FileServer(gin.Dir(config.Get().LandingSettings.StaticPath, true))))

// // 	applyCors(r)
// // 	applyRoute(r)

// // 	server := &http.Server{
// // 		Addr:         config.Get().LandingSettings.ListenAddress,
// // 		Handler:      r,
// // 		ReadTimeout:  5 * time.Minute,
// // 		WriteTimeout: 5 * time.Minute,
// // 	}
// // 	err := server.ListenAndServe()
// // 	if err != nil {
// // 	}
// // }

// // func MultiStatic(rootFolders []string) gin.HandlerFunc {
// // 	return func(c *gin.Context) {
// // 		for _, root := range rootFolders {
// // 			filePath := root + c.Request.URL.Path
// // 			if _, err := os.Stat(filePath); err == nil {
// // 				c.File(filePath)
// // 				return
// // 			}
// // 		}

// // 		c.Next()
// // 	}
// // }

// // func applyRoute(r *gin.Engine) {
// // 	router.Register(r)
// // }

// // func applyCors(r *gin.Engine) {
// // 	// // 配置跨域
// // 	corsConfig := cors.DefaultConfig()
// // 	corsConfig.AllowAllOrigins = true
// // 	// corsConfig.AllowCredentials = true
// // 	// corsConfig.AllowOrigins = []string{"http://127.0.0.1:5502/"}
// // 	corsConfig.AddAllowHeaders("Authorization", "X-User")
// // 	r.Use(cors.New(corsConfig))
// // }
