# 快速开始指南

本指南将帮助你在 5 分钟内启动 PixelHub。

## 前置要求

- Go 1.21+ 已安装
- 有腾讯云 COS 账号和存储桶（或使用其他对象存储）

## 快速安装

### 方法 1: 使用初始化脚本（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/vaaandark/PixelHub.git
cd PixelHub

# 2. 运行初始化脚本
chmod +x scripts/init.sh
./scripts/init.sh

# 3. 编辑配置文件
vim config.toml
# 填入你的腾讯云 COS 配置

# 4. 启动服务
./bin/pixelhub
```

### 方法 2: 手动安装

```bash
# 1. 克隆项目
git clone https://github.com/vaaandark/PixelHub.git
cd PixelHub

# 2. 下载依赖
go mod download

# 3. 创建配置文件
cp config.example.toml config.toml
vim config.toml  # 编辑配置

# 4. 编译并运行
go run cmd/server/main.go
```

### 方法 3: 使用 Docker

```bash
# 1. 克隆项目
git clone https://github.com/vaaandark/PixelHub.git
cd PixelHub

# 2. 创建配置文件
cp config.example.toml config.toml
vim config.toml  # 编辑配置

# 3. 使用 Docker Compose 启动
docker-compose up -d
```

## 配置说明

编辑 `config.toml` 文件：

```toml
[server]
host = "0.0.0.0"
port = "8080"

[database]
path = "./data/pixelhub.db"

[storage]
provider = "tencent-cos"

[storage.tencent_cos]
secret_id = "AKIDxxxxxxxxxxxxxxxx"      # ⚠️ 填入你的 SecretId
secret_key = "xxxxxxxxxxxxxxxxxxxxxxx"  # ⚠️ 填入你的 SecretKey
bucket_url = "https://your-bucket-1234567890.cos.ap-guangzhou.myqcloud.com"  # ⚠️ 填入存储桶 URL
cdn_url = "https://cdn.your-domain.com"  # 可选：CDN 加速域名
```

### 获取腾讯云 COS 配置

1. 登录 [腾讯云控制台](https://console.cloud.tencent.com/)
2. 进入 [对象存储 COS](https://console.cloud.tencent.com/cos)
3. 创建存储桶（如果还没有）
4. 获取以下信息：
   - **SecretId 和 SecretKey**: 在"访问管理" → "API密钥管理"
   - **Bucket URL**: 在存储桶列表中查看
   - **CDN URL**: 如果配置了 CDN，填入 CDN 域名

## 验证安装

### 1. 访问 Web 界面

打开浏览器访问：`http://localhost:8080`

你应该看到 PixelHub 的主界面。

### 2. 测试上传

```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "file=@/path/to/test-image.jpg"
```

成功响应示例：
```json
{
  "code": 201,
  "message": "Upload successful",
  "data": {
    "image_id": "img_abc123456789",
    "url": "https://your-bucket.cos.ap-guangzhou.myqcloud.com/img_abc123456789.jpg",
    "hash": "e6884675b87..."
  }
}
```

### 3. 测试标签功能

```bash
# 添加标签
curl -X PUT http://localhost:8080/api/v1/images/img_abc123456789/tags \
  -H "Content-Type: application/json" \
  -d '{"tags": ["测试", "风景"], "mode": "set"}'

# 搜索图片
curl "http://localhost:8080/api/v1/search/exact?tags=测试,风景"
```

## 常见问题

### Q: 端口被占用怎么办？

修改 `config.toml` 中的端口号：
```toml
[server]
port = "8090"  # 改为其他端口
```

### Q: 上传失败，提示存储错误？

1. 检查腾讯云 COS 配置是否正确
2. 确认 SecretId 和 SecretKey 有效
3. 确认存储桶 URL 正确
4. 检查存储桶权限设置

### Q: 数据库文件在哪里？

默认位置：`./data/pixelhub.db`

可在 `config.toml` 中修改：
```toml
[database]
path = "/your/custom/path/pixelhub.db"
```

### Q: 如何备份数据？

```bash
# 备份数据库
cp data/pixelhub.db data/pixelhub.db.backup

# 图片已存储在对象存储中，无需额外备份
```

### Q: 如何开启开发模式（热重载）？

```bash
# 安装 Air
go install github.com/cosmtrek/air@latest

# 运行
air
```

## 下一步

- 📖 阅读 [完整文档](README.md)
- 📚 查看 [API 文档](docs/API.md)
- 🏗️ 了解 [系统架构](docs/ARCHITECTURE.md)
- 🤝 参与 [贡献](CONTRIBUTING.md)

## 获取帮助

- 📝 [提交 Issue](https://github.com/vaaandark/PixelHub/issues)
- 💬 [参与讨论](https://github.com/vaaandark/PixelHub/discussions)

---

祝你使用愉快！如果遇到问题，请随时提 Issue。

