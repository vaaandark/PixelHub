# PixelHub 🖼️

PixelHub 是一个现代化的图床应用，支持图片上传、管理和标签系统。它提供强大的标签功能，让你可以通过标签快速检索和组织图片。

## ✨ 特性

- 🚀 **快速上传**: 支持拖拽上传、点击上传多种方式，可添加图片描述
- 📝 **图片描述**: 为每张图片添加详细描述信息，便于管理和搜索
- 🏷️ **标签管理**: 为图片添加标签，支持多标签组合搜索
- 🔍 **智能搜索**: 
  - 精确搜索（AND 逻辑）：只返回包含所有指定标签的图片
  - 相关性搜索（OR 逻辑）：按匹配标签数量排序
- 📦 **灵活存储**: 抽象的存储接口，支持多种对象存储服务
- 🔌 **可扩展**: 支持通过 MCP 等协议与外部服务集成（可独立部署）
- 🎨 **现代 UI**: 美观的用户界面，响应式设计

## 🏗️ 技术栈

- **后端**: Go 1.21+, Gin Web Framework
- **前端**: 原生 JavaScript, 现代 CSS
- **数据库**: SQLite 3
- **存储**: 腾讯云 COS（可扩展其他存储）

## 📦 安装

### 前置要求

- Go 1.21 或更高版本
- SQLite 3

### 克隆项目

```bash
git clone https://github.com/vaaandark/PixelHub.git
cd PixelHub
```

### 安装依赖

```bash
go mod download
```

### 配置

复制配置文件模板并编辑：

```bash
cp config.example.toml config.toml
```

编辑 `config.toml`，填入你的配置信息：

```toml
[server]
host = "0.0.0.0"
port = "8080"

[database]
path = "./data/pixelhub.db"

[storage]
provider = "tencent-cos"

[storage.tencent_cos]
secret_id = "your-secret-id"        # 填入腾讯云 SecretId
secret_key = "your-secret-key"      # 填入腾讯云 SecretKey
bucket_url = "https://your-bucket-1234567890.cos.ap-guangzhou.myqcloud.com"  # 填入你的存储桶 URL
cdn_url = "https://cdn.your-imagehost.com"  # 可选：CDN 加速域名
```

### 运行

```bash
go run cmd/server/main.go
```

服务器将在 `http://localhost:8080` 启动。

## 🚀 使用指南

### Web 界面

访问 `http://localhost:8080` 即可使用图床的 Web 界面。

**功能说明**：

1. **上传图片**: 点击或拖拽图片到上传区域，可选择添加描述
2. **管理图片**: 点击图片查看详情，可编辑描述、标签或删除图片
3. **搜索图片**: 在搜索框输入标签（用逗号分隔），选择搜索模式
4. **浏览图片**: 查看所有已上传的图片，支持排序和分页
5. **浏览标签**: 查看热门标签，点击标签快速搜索

### API 使用

#### 图床后端 API

基础 URL: `http://localhost:8080/api/v1`

**上传图片**:
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "file=@/path/to/image.jpg" \
  -F "description=美丽的风景照片"
```

**列出所有图片**:
```bash
# 获取第一页（默认最新优先）
curl http://localhost:8080/api/v1/images

# 获取第二页，每页 10 条
curl "http://localhost:8080/api/v1/images?page=2&limit=10"

# 按上传时间升序排序
curl "http://localhost:8080/api/v1/images?sort=date_asc"
```

**获取图片详情**:
```bash
curl http://localhost:8080/api/v1/images/{image_id}
```

**更新图片描述**:
```bash
curl -X PUT http://localhost:8080/api/v1/images/{image_id} \
  -H "Content-Type: application/json" \
  -d '{"description": "更新后的描述"}'
```

**更新图片标签**:
```bash
curl -X PUT http://localhost:8080/api/v1/images/{image_id}/tags \
  -H "Content-Type: application/json" \
  -d '{"tags": ["风景", "自然"], "mode": "set"}'
```

**搜索图片**:
```bash
# 精确搜索（AND 逻辑）
curl "http://localhost:8080/api/v1/search/exact?tags=风景,自然&page=1&limit=20"

# 相关性搜索（OR 逻辑）
curl "http://localhost:8080/api/v1/search/relevance?tags=风景,自然&page=1&limit=20"
```

**列出所有标签**:
```bash
curl "http://localhost:8080/api/v1/tags?page=1&limit=50"
```

**删除图片**:
```bash
curl -X DELETE http://localhost:8080/api/v1/images/{image_id}
```

完整的 API 文档请参考 [docs/API.md](docs/API.md)。

## 🔧 开发

### 项目结构

```
PixelHub/
├── cmd/
│   └── server/          # 主程序入口
│       └── main.go
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库层
│   ├── handlers/        # API 处理器
│   └── storage/         # 存储接口和实现
├── web/                 # 前端文件
│   ├── index.html
│   └── static/
│       ├── css/
│       └── js/
├── data/                # 数据目录（自动创建）
├── config.toml          # 配置文件（需手动创建）
├── config.example.toml  # 配置模板
├── go.mod
└── README.md
```

### 扩展存储提供商

PixelHub 的存储层采用接口设计，可以轻松扩展支持其他对象存储服务：

1. 在 `internal/storage/` 创建新的 provider 文件
2. 实现 `Provider` 接口：
   ```go
   type Provider interface {
       Upload(filename string, content io.Reader, contentType string) (storageKey string, url string, err error)
       Delete(storageKey string) error
       GetURL(storageKey string) string
   }
   ```
3. 在 `storage.go` 的 `NewProvider` 函数中注册新的 provider

## 📝 API 文档

详细的 API 文档请参考：[.cursor/rules/api.mdc](.cursor/rules/api.mdc)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web Framework
- [SQLite](https://www.sqlite.org/) - 数据库
- [Tencent Cloud COS](https://cloud.tencent.com/product/cos) - 对象存储

---

Made with ❤️ by [vaaandark](https://github.com/vaaandark)

