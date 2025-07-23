package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// # 自定义参数
// go run proxy.go --target=http://192.168.1.100:8080 --port=9090

func main() {
	// 定义命令行参数
	targetURL := flag.String("target", "http://172.19.19.22:18082", "目标服务器地址（格式：http://ip:port）")
	proxyPort := flag.String("port", "8082", "代理服务器监听的端口")
	flag.Parse()

	// 解析目标地址
	target, err := url.Parse(*targetURL)
	if err != nil {
		log.Fatal("无效的目标地址：", err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 强制覆盖Host头（确保目标服务器能正确接收）
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}

	// 在代理响应中添加CORS头
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("Access-Control-Allow-Origin", "*")
		resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		return nil
	}

	// 自定义处理器（处理OPTIONS请求）
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 处理OPTIONS预检请求
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "1728000") // 20天缓存
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 其他请求转发到目标服务器
		proxy.ServeHTTP(w, r)
	})

	// 启动代理服务器（端口前添加冒号，格式如 :8081）
	listenAddr := ":" + *proxyPort
	log.Printf("代理服务器已启动，监听端口 %s，目标地址 %s\n", listenAddr, *targetURL)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal("代理服务器错误：", err)
	}
}
