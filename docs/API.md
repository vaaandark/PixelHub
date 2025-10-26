# PixelHub API 文档

## 基础信息

- **基础 URL**: `http://localhost:8080/api/v1`
- **MCP 基础 URL**: `http://localhost:8080/mcp/v1`
- **内容类型**: `application/json`

## 认证

### 图床 API
图床 API 不需要认证（可根据需求添加）。

### MCP API
MCP API 需要在请求头中提供 API Key：
```
Authorization: Bearer {your-api-key}
```

---

## 图床 API 端点

### 1. 上传图片

上传图片文件到服务器。

**请求**
```http
POST /api/v1/images/upload
Content-Type: multipart/form-data
```

**参数**
- `file` (required): 图片文件

**响应**
```json
{
  "code": 201,
  "message": "Upload successful",
  "data": {
    "image_id": "img_a1b2c3d4",
    "url": "https://cdn.your-imagehost.com/a1b2c3d4.jpg",
    "hash": "e6884675b87..."
  }
}
```

**cURL 示例**
```bash
curl -X POST http://localhost:8080/api/v1/images/upload \
  -F "file=@/path/to/image.jpg"
```

---

### 2. 获取图片详情

获取指定图片的元数据和标签。

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

### 3. 更新图片标签

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

### 4. 删除图片

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

### 5. 列出所有标签

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

### 6. 精确搜索图片

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

## MCP API 端点

### 1. 列出标签

获取标签列表，供 AI 了解可用的标签。

**请求**
```http
GET /mcp/v1/tags?page=1&limit=100
Authorization: Bearer {your-api-key}
```

**查询参数**
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 100，最大 1000

**响应**
```json
{
  "tags": ["风景", "猫咪", "艺术", "食物", "抽象"],
  "total": 1250,
  "has_more": true
}
```

**cURL 示例**
```bash
curl http://localhost:8080/mcp/v1/tags \
  -H "Authorization: Bearer your-secret-api-key"
```

---

### 2. 相关性搜索

使用 OR 逻辑搜索，返回按匹配标签数降序排列的图片。

**请求**
```http
GET /mcp/v1/search/relevance?tags=cat,cute&page=1&limit=20
Authorization: Bearer {your-api-key}
```

**查询参数**
- `tags` (required): 标签列表，用逗号分隔
- `page` (optional): 页码，默认 1
- `limit` (optional): 每页数量，默认 20，最大 100

**响应**
```json
{
  "total": 520,
  "results": [
    {
      "image_id": "img_x1y2z3a4",
      "url": "https://cdn.your-imagehost.com/x1y2z3a4.jpg",
      "matched_tag_count": 3,
      "tags": ["cat", "cute", "pet"]
    },
    {
      "image_id": "img_b5c6d7e8",
      "url": "https://cdn.your-imagehost.com/b5c6d7e8.jpg",
      "matched_tag_count": 2,
      "tags": ["cat", "animal"]
    }
  ]
}
```

**cURL 示例**
```bash
curl "http://localhost:8080/mcp/v1/search/relevance?tags=cat,cute&limit=10" \
  -H "Authorization: Bearer your-secret-api-key"
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
- `401 Unauthorized`: 未授权（MCP API）
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

1. **图片上传后添加标签**：上传成功后立即为图片添加标签，便于后续检索
2. **使用有意义的标签**：标签应该清晰描述图片内容
3. **标签规范化**：使用统一的标签命名规范（如全小写、中英文统一等）
4. **批量操作**：如需更新大量图片，考虑分批处理
5. **搜索优化**：
   - 精确搜索：用于需要同时满足多个条件的场景
   - 相关性搜索：用于寻找相关图片的场景

