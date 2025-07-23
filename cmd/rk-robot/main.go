package main

import (
	"fmt"
	"net/http"
	"rk-api/pkg/logger"
	"rk-api/test"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// 假设有一个日志的存储结构，这里我们简单地使用一个字符串切片
var logs []string
var logsMutex sync.Mutex  // 用于保证日志并发访问的安全
var activeRobotsCount int // 活跃机器人数目

var robotIndex int // robot 添加序号
// 最大日志条数
const maxLogEntries = 40

const APIURL = "https://api-dev.cheetahs.asia/api/"

// appendLog 添加一条新的日志记录，同时确保日志记录不超过40条
func appendLog(newLog string) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	// 添加新的日志记录
	logs = append(logs, newLog)

	// 保留最新的40条日志记录
	if len(logs) > maxLogEntries {
		// 移除最旧的记录
		start := len(logs) - maxLogEntries
		logs = logs[start:]
	}
}

func InitLogger() (func(), error) {
	logger := logger.GetLogger()
	defer logger.Sync()

	return func() {}, nil
}

func startRobot(username, password string) {
	robot := test.NewApiRobot(username, password)
	robot.SetLogger(appendLog)
	robot.SetAPIUrl(APIURL)
	go robot.Start()

	activeRobotsCount++ //数量加1
}

func main() {
	router := gin.Default()

	// 静态资源服务，假设前端HTML文件放在 "./static" 目录下
	router.Static("/public", "./public")

	InitLogger()

	accounts, passwords, err := test.ReadCredentials()
	if err != nil {
		fmt.Println("Error reading credentials:", err)
	}

	// 注册机器人的API
	router.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		// 实际操作：添加新机器人到数据库或其他存储系统

		// 记录操作到日志
		appendLog(fmt.Sprintf("Registered new robot: %s", username))

		startRobot(username, password)

		// 响应前端
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Robot registered."})
	})

	// 增加机器人数量的API
	router.POST("/add", func(c *gin.Context) {
		quantityStr := c.PostForm("quantity")
		quantity, _ := strconv.Atoi(quantityStr)
		for i := 0; i < quantity; i++ {
			startRobot(accounts[robotIndex], passwords[robotIndex])
			robotIndex++
		}

		// 实际操作：根据数量增加机器人到数据库或其他存储系统
		// 记录操作到日志
		logs = append(logs, fmt.Sprintf("Added %d robots", quantity))
		// 响应前端
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": fmt.Sprintf("%d robots added.", quantity)})
	})

	// 获取日志的API
	router.GET("/logs", func(c *gin.Context) {
		logsMutex.Lock()
		defer logsMutex.Unlock()
		c.JSON(http.StatusOK, gin.H{"status": "success", "logs": logs})
	})

	// 添加处理 /activeRobotsCount 端点的路由
	router.GET("/activeRobotsCount", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"count": activeRobotsCount,
		})
	})

	// 启动Gin服务
	router.Run(":8080")

	// 异步启动Gin服务
	// go func() {
	// 	if err := router.Run(":8080"); err != nil {
	// 		fmt.Println("Gin server failed to run:", err)
	// 	}
	// }()

	// ui, _ := lorca.New("", "", 800, 600)
	// defer ui.Close()

	// <-ui.Done()

}

func startWebView() {
	// // Webview 启动设置
	// debug := true
	// w := webview.New(debug)
	// defer w.Destroy()
	// w.SetTitle("机器人管理系统")
	// w.SetSize(800, 600, webview.HintNone)

	// // 给Gin web server一秒钟的时间启动
	// time.Sleep(time.Second)

	// // 加载本地服务器的URL
	// w.Navigate("http://localhost:8080/public/robot.html")

	// // 启动Webview界面
	// w.Run()

}
