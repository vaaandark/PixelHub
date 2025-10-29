# PixelHub MCP Server

## Version
v0.1.0

## Overview

PixelHub MCP Server is a Model Context Protocol server that provides MCP clients (such as Claude Desktop) with the ability to interact with the PixelHub image hosting service. It enables natural language-based image discovery through tag management and semantic search capabilities.

## Category
Image Hosting & Management

## Features

- List all available tags with usage counts
- Search images by tags with relevance ranking
- Semantic image discovery through LLM-powered tag matching

## Available Tools

- `list_tags`: List all available tags in the system with their usage counts
- `search_images_by_tags`: Search images by tags using relevance ranking (OR logic with match count sorting)

## Usage Guide

### Prerequisites
- Python 3.12+
- UV (recommended) or pip
- Running PixelHub server

**Install UV (Linux/macOS):**
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Install UV (Windows):**
```bash
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
```

### Installation
Clone the repository:
```bash
git clone <repository-url>
cd PixelHub/mcp
```

### Usage
Start the server:

#### UV
```bash
cd mcp
uv run mcp-server-pixelhub

# Start with streamable-http mode (default is stdio)
uv run mcp-server-pixelhub -t streamable-http

# Start with SSE mode
uv run mcp-server-pixelhub -t sse
```

#### Pip
```bash
cd mcp
pip install -e .
mcp-server-pixelhub
```

Use a client to interact with the server:
```
Cursor | Claude Desktop | Cline | ...
```

## Configuration

### Environment Variables

The following environment variables are available for configuring the MCP server:

| Environment Variable | Description | Default Value |
|----------|------|--------|
| `PIXELHUB_BASE_URL` | PixelHub server base URL | `http://localhost:8080` |
| `MCP_SERVER_PORT` | MCP server listening port | `8000` |

For example, set these environment variables before starting the server:

```bash
export PIXELHUB_BASE_URL=http://localhost:8080
export MCP_SERVER_PORT=8000
```

### Run with uvx (Local Installation)
```json
{
    "mcpServers": {
        "mcp-server-pixelhub": {
            "command": "uvx",
            "args": [
            "--from",
            "git+https://github.com/vaaandark/PixelHub#subdirectory=mcp",
            "mcp-server-pixelhub"
          ],
            "env": {
                "PIXELHUB_BASE_URL": "http://localhost:8080",
                "MCP_SERVER_PORT": "8000"
            }
        }
    }
}
```

### Remote MCP Server Configuration

If you want to run the MCP server as a remote service (e.g., on a server), you can use the HTTP transport modes:

#### Option 1: Streamable HTTP
```json
{
    "mcpServers": {
        "pixelhub-remote": {
            "url": "http://your-server:8000/mcp"
        }
    }
}
```

#### Option 2: SSE (Server-Sent Events)
```json
{
    "mcpServers": {
        "pixelhub-remote": {
            "url": "http://your-server:8000/sse",
            "transport": "sse"
        }
    }
}
```

#### Running Remote Server

On your server, start the MCP server with HTTP transport:

```bash
# For streamable-http mode
export PIXELHUB_BASE_URL=http://your-pixelhub-server:8080
export MCP_SERVER_HOST=0.0.0.0
export MCP_SERVER_PORT=8000
uv run mcp-server-pixelhub -t streamable-http

# For SSE mode
uv run mcp-server-pixelhub -t sse
```

## Examples

### Use Cases

1. **Semantic Image Discovery**: 
   - LLM analyzes user's request and finds relevant tags
   - Search images using those tags with relevance ranking
   - Perfect for finding images that match a specific mood, style, or theme

2. **Tag Exploration**:
   - List all available tags to understand the image collection
   - Discover new categories and themes in your image library

3. **Content Creation**:
   - Find images for blog posts, presentations, or creative projects
   - Search by multiple related tags to find the perfect match

### Example Workflow

1. User: "I need some nature photos with mountains"
2. MCP calls `list_tags` to see available tags
3. LLM identifies relevant tags like "nature", "mountains", "landscape"
4. MCP calls `search_images_by_tags` with those tags
5. Returns images ranked by relevance (number of matching tags)

## License
This project follows the same license as the main PixelHub project.
