# 使用官方 Go 镜像作为基础镜像
FROM golang:1.22 AS builder
# 设置 GOPROXY 使用国内镜像
# ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录
WORKDIR /app

# 将 go.mod 和 go.sum 复制到工作目录
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将源代码和公共文件夹复制到工作目录
COPY . .

# 使用静态编译构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X main.VERSION=v8.1.0" -o rk-api ./cmd/rk-api

# 使用轻量级的镜像运行应用程序
FROM alpine:latest

# 安装基本的CA证书，时区数据
# RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata ca-certificates

# 设置工作目录
WORKDIR /app
# # 將本地的 configs 目錄內容複製到容器的 /app/configs
# COPY configs/ /app/configs/
# # 將本地的 public 目錄內容複製到容器的 /app/public
# COPY public/ /app/public/

# 从builder阶段复制文件
COPY --from=builder /app/rk-api .
COPY --from=builder /app/public ./public
COPY configs/*.yaml ./configs/

# 设置执行权限
RUN chmod +x /app/rk-api

# 暴露端口
EXPOSE 3001

# 设置入口点
ENTRYPOINT ["/app/rk-api"]
CMD ["web", "-c", "/app/configs/config-linux-dev.yaml"]