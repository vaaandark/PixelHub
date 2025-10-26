# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pixelhub ./cmd/server

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite-libs

# 从构建阶段复制二进制文件
COPY --from=builder /build/pixelhub .

# 复制前端文件
COPY --from=builder /build/web ./web

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./pixelhub"]

