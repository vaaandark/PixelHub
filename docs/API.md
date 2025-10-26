# PixelHub API 文档

## 基础信息

- **基础 URL**: `http://localhost:8080/api/v1`
- **内容类型**: `application/json`

## 认证

图床 API 不需要认证（可根据需求添加）。

---

## API 端点

### 1. 上传图片

上传图片文件到服务器，可选添加图片描述。

**请求**
```http
POST /api/v1/images/upload
Content-Type: multipart/form-data
```

**参数**
- `file` (required): 图片文件
- `description` (optional): 图片描述信息

**响应**
```json
{
  "code": 201,
  "message": "Upload successful",
  "data": {
    "image_id": "img_a1b2c3d4",
    "url": "https://cdn.your-imagehost.com/a1b2c3d4.jpg",
    "hash": "e6884675b87...",
    "description": "美丽的风景照片"
  }
}
```

**cURL 示例**
```bash
# 不带描述
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "file=@/path/to/image.jpg"

# 带描述
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "file=@/path/to/image.jpg" \
  -F "description=美丽的风景照片"
```

---

### 2. 列出所有图片

获取所有图片列表，支持分页和排序。

**请求**
```http
GET /api/v1/images?page=1&limit=20&sort=date_desc
```

**查询参数**
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 20，最大 100
- `sort` (optional): 排序方式
  - `date_desc` (默认): 最新优先
  - `date_asc`: 最旧优先

**响应**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "total": 150,
    "current_page": 1,
    "images": [
      {
        "id": "img_a1b2c3d4",
        "url": "https://cdn.your-imagehost.com/a1b2c3d4.jpg",
        "description": "美丽的风景照片",
        "upload_date": "2025-10-26T12:00:00Z",
        "tags": ["风景", "自然"]
      },
      {
        "id": "img_b5c6d7e8",
        "url": "https://cdn.your-imagehost.com/b5c6d7e8.jpg",
        "description": "城市夜景",
        "upload_date": "2025-10-25T20:30:00Z",
        "tags": ["城市", "夜景"]
      }
    ]
  }
}
```

**cURL 示例**
```bash
# 获取第一页（默认）
curl http://localhost:8080/api/v1/images

# 获取第二页，每页 10 条
curl "http://localhost:8080/api/v1/images?page=2&limit=10"

# 按上传时间升序排序
curl "http://localhost:8080/api/v1/images?sort=date_asc"
```

---

### 3. 获取图片详情

获取指定图片的完整信息，包括 URL、描述、标签等元数据。

**请求**
```http
GET /api/v1/images/{image_id}
```

**响应**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "image_id": "img_a1b2c3d4",
    "url": "https://cdn.your-imagehost.com/a1b2c3d4.jpg",
    "hash": "e6884675b87...",
    "description": "美丽的风景照片",
    "upload_date": "2025-10-26T12:00:00Z",
    "tags": ["风景", "自然", "山川"]
  }
}
```

**cURL 示例**
```bash
curl http://localhost:8080/api/v1/images/img_a1b2c3d4
```

---

### 4. 更新图片描述

更新指定图片的描述信息。

**请求**
```http
PUT /api/v1/images/{image_id}
Content-Type: application/json
```

**参数**
```json
{
  "description": "更新后的图片描述"
}
```

- `description` (required): 新的图片描述

**响应**
```json
{
  "code": 200,
  "message": "Description updated successfully"
}
```

**cURL 示例**
```bash
curl -X PUT http://localhost:8080/api/v1/images/img_a1b2c3d4 \
  -H "Content-Type: application/json" \
  -d '{"description": "更新后的图片描述"}'
```

---

### 5. 更新图片标签

为图片添加或更新标签。

**请求**
```http
PUT /api/v1/images/{image_id}/tags
Content-Type: application/json
```

**参数**
```json
{
  "tags": ["风景", "自然"],
  "mode": "set"  // "set" 或 "append"
}
```

- `tags` (required): 标签数组
- `mode` (optional): 
  - `set` (默认): 替换所有标签
  - `append`: 追加标签

**响应**
```json
{
  "code": 200,
  "message": "Tags updated successfully"
}
```

**cURL 示例**
```bash
# 替换标签
curl -X PUT http://localhost:8080/api/v1/images/img_a1b2c3d4/tags \
  -H "Content-Type: application/json" \
  -d '{"tags": ["风景", "自然"], "mode": "set"}'

# 追加标签
curl -X PUT http://localhost:8080/api/v1/images/img_a1b2c3d4/tags \
  -H "Content-Type: application/json" \
  -d '{"tags": ["美丽"], "mode": "append"}'
```

---

### 6. 删除图片

删除指定图片（软删除）。

**请求**
```http
DELETE /api/v1/images/{image_id}
```

**响应**
```json
{
  "code": 200,
  "message": "Image deleted successfully"
}
```

**cURL 示例**
```bash
curl -X DELETE http://localhost:8080/api/v1/images/img_a1b2c3d4
```

---

### 7. 列出所有标签

获取系统中所有标签的列表（按使用次数排序）。

**请求**
```http
GET /api/v1/tags?page=1&limit=50
```

**查询参数**
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 50，最大 100

**响应**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "total": 1250,
    "current_page": 1,
    "tags": [
      {"name": "风景", "count": 520},
      {"name": "猫咪", "count": 480},
      {"name": "艺术", "count": 300}
    ]
  }
}
```

**cURL 示例**
```bash
curl "http://localhost:8080/api/v1/tags?page=1&limit=50"
```

---

### 8. 精确搜索图片

使用 AND 逻辑搜索包含所有指定标签的图片。

**请求**
```http
GET /api/v1/search/exact?tags=风景,自然&page=1&limit=20
```

**查询参数**
- `tags` (required): 标签列表，用逗号分隔
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 20，最大 100

**响应**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "total": 150,
    "current_page": 1,
    "results": [
      {
        "id": "img_x1y2z3a4",
        "url": "https://cdn.your-imagehost.com/x1y2z3a4.jpg",
        "hash": "abc123...",
        "description": "美丽的山川风景",
        "upload_date": "2025-10-26T12:00:00Z",
        "tags": ["风景", "自然", "山川"]
      }
    ]
  }
}
```

**cURL 示例**
```bash
curl "http://localhost:8080/api/v1/search/exact?tags=风景,自然&page=1&limit=20"
```

---

### 9. 相关性搜索图片

使用 OR 逻辑搜索，返回包含任一标签的图片，按匹配标签数降序排列。

**请求**
```http
GET /api/v1/search/relevance?tags=风景,自然&page=1&limit=20
```

**查询参数**
- `tags` (required): 标签列表，用逗号分隔
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 20，最大 100

**响应**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "total": 520,
    "current_page": 1,
    "results": [
      {
        "id": "img_x1y2z3a4",
        "url": "https://cdn.your-imagehost.com/x1y2z3a4.jpg",
        "hash": "abc123...",
        "description": "美丽的山川风景",
        "upload_date": "2025-10-26T12:00:00Z",
        "matched_tag_count": 3,
        "tags": ["风景", "自然", "山川"]
      },
      {
        "id": "img_b5c6d7e8",
        "url": "https://cdn.your-imagehost.com/b5c6d7e8.jpg",
        "hash": "def456...",
        "description": "自然风光",
        "upload_date": "2025-10-25T10:30:00Z",
        "matched_tag_count": 2,
        "tags": ["风景", "山川", "日落"]
      }
    ]
  }
}
```

**cURL 示例**
```bash
curl "http://localhost:8080/api/v1/search/relevance?tags=风景,自然&page=1&limit=20"
```

---

## 错误响应

所有错误响应遵循统一格式：

```json
{
  "code": 400,
  "message": "Error description"
}
```

### 常见错误码

- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

---

## 数据模型

### Picture (图片)

```json
{
  "id": "string",           // 图片 ID
  "url": "string",          // 访问 URL
  "storage_key": "string",  // 存储键
  "hash": "string",         // 文件哈希
  "description": "string",  // 图片描述（可选）
  "upload_date": "string",  // 上传时间 (ISO 8601)
  "deleted": false          // 是否已删除
}
```

### Tag (标签)

```json
{
  "name": "string",  // 标签名称
  "count": 123       // 使用次数
}
```

---

## 使用限制

- 单次上传文件大小：无限制（由服务器配置决定）
- 标签数量：每张图片无限制
- 标签长度：建议不超过 50 个字符
- 搜索标签数量：建议不超过 10 个

---

## 最佳实践

1. **完善图片信息**：
   - 上传时添加描述信息，说明图片内容和用途
   - 上传成功后立即添加相关标签，便于后续检索和分类
   
2. **描述规范**：
   - 使用简洁明确的描述文字
   - 描述应包含图片的关键信息（内容、场景、用途等）
   - 建议长度：10-100 个字符
   
3. **标签规范**：
   - 使用有意义的标签，清晰描述图片特征
   - 采用统一的命名规范（如全小写、中英文统一）
   - 每张图片建议 3-10 个标签
   - 标签长度建议不超过 50 个字符
   
4. **批量操作**：
   - 如需更新大量图片，考虑分批处理
   - 使用脚本自动化标签管理
   
5. **搜索优化**：
   - **精确搜索**：用于需要同时满足多个条件的场景（AND 逻辑）
   - **相关性搜索**：用于寻找相关图片的场景（OR 逻辑）
   - 搜索标签数量建议不超过 10 个
   
6. **描述与标签配合使用**：
   - 描述用于详细说明，标签用于分类检索
   - 描述可以包含更多上下文信息
   - 标签应该是通用的、可复用的关键词

