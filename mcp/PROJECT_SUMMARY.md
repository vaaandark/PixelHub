# PixelHub MCP Server - Project Summary

## 项目概述

成功为 PixelHub 图床服务实现了 MCP (Model Context Protocol) 服务器，使 LLM 能够通过自然语言与图片库进行语义化交互。

## 实现的功能

### 1. 核心 MCP 工具

- **`list_tags`**: 列出所有可用标签及其使用次数
- **`search_images_by_tags`**: 基于标签的相关性搜索（OR 逻辑，按匹配标签数降序排列）

### 2. 传输协议支持

- **stdio**: 标准输入输出（默认，适用于大多数 MCP 客户端）
- **sse**: Server-Sent Events
- **streamable-http**: HTTP 流式传输

### 3. 语义化搜索能力

通过 LLM 的理解能力，用户可以用自然语言描述需求，MCP 服务器会：
1. 获取所有可用标签
2. LLM 分析并选择最相关的标签
3. 执行相关性搜索，返回按匹配度排序的图片

## 项目结构

```
mcp/
├── pyproject.toml              # 项目配置和依赖
├── uv.lock                     # 依赖锁定文件
├── README.md                   # 项目说明文档
├── USAGE.md                    # 详细使用指南
├── example_config.json         # MCP 客户端配置示例
├── run_example.py             # 测试脚本
└── src/mcp_server_pixelhub/
    ├── __init__.py
    ├── main.py                # 主程序入口
    ├── common/                # 公共模块
    │   ├── __init__.py
    │   ├── client.py          # HTTP 客户端
    │   ├── config.py          # 配置管理
    │   ├── errors.py          # 错误处理
    │   └── logs.py            # 日志管理
    └── tools/                 # MCP 工具实现
        ├── __init__.py        # FastMCP 初始化
        ├── images.py          # 图片搜索工具
        └── tags.py            # 标签列表工具
```

## 技术特点

### 1. 架构设计
- 模块化设计，易于扩展
- 统一的错误处理机制
- 完善的日志记录
- 配置化的 HTTP 客户端

### 2. API 集成
- 与 PixelHub API v1 完全兼容
- 支持分页查询
- 自动参数验证和错误处理

### 3. MCP 协议实现
- 基于 FastMCP 框架
- 支持多种传输协议
- 完整的工具描述和参数定义

## 使用场景

### 1. 语义化图片发现
```
用户: "我需要一些山景日落的照片"
→ LLM 识别相关标签: ["mountains", "sunset", "landscape"]
→ MCP 搜索并返回按相关性排序的图片
```

### 2. 内容探索
```
用户: "我的图片库里都有什么类型的照片？"
→ MCP 列出所有标签和使用频率
→ LLM 分析并总结图片库的内容分布
```

### 3. 精确主题搜索
```
用户: "找一些城市夜景的照片，要有好的光线效果"
→ LLM 选择标签: ["urban", "night", "city", "lighting"]
→ MCP 返回匹配度最高的图片
```

## 配置和部署

### 环境变量
- `PIXELHUB_BASE_URL`: PixelHub 服务器地址（默认: http://localhost:8080）
- `MCP_SERVER_PORT`: MCP 服务器端口（默认: 8000）

### 客户端集成
支持主流 MCP 客户端：
- Claude Desktop
- Cursor
- Cline
- 其他兼容 MCP 协议的工具

## 优势

1. **语义化搜索**: 利用 LLM 的理解能力实现自然语言图片搜索
2. **相关性排序**: 按标签匹配数量排序，确保最相关的结果优先显示
3. **易于扩展**: 模块化设计，可轻松添加新的搜索功能
4. **协议兼容**: 支持多种 MCP 传输协议，适配不同客户端
5. **完善的错误处理**: 统一的错误处理和日志记录机制

## 测试验证

- ✅ MCP 服务器正常启动
- ✅ 支持所有三种传输协议
- ✅ 工具参数验证正确
- ✅ 错误处理机制完善
- ✅ 日志记录功能正常

## 总结

成功实现了一个功能完整、架构清晰的 PixelHub MCP 服务器，为图片库提供了强大的语义化搜索能力。通过 LLM 的自然语言理解和 MCP 的标准化接口，用户可以用自然语言轻松发现和获取所需的图片资源。
