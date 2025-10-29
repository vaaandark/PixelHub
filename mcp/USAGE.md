# PixelHub MCP Server Usage Guide

## Quick Start

### 1. Prerequisites

Make sure you have:
- Python 3.12+ installed
- UV package manager installed
- PixelHub server running (default: http://localhost:8080)

### 2. Installation

```bash
cd PixelHub/mcp
uv sync
```

### 3. Running the Server

#### Stdio Mode (default, for MCP clients)
```bash
uv run mcp-server-pixelhub
```

#### SSE Mode
```bash
uv run mcp-server-pixelhub -t sse
```

#### Streamable HTTP Mode
```bash
uv run mcp-server-pixelhub -t streamable-http
```

## Configuration

### Environment Variables

- `PIXELHUB_BASE_URL`: Base URL of PixelHub server (default: http://localhost:8080)
- `MCP_SERVER_PORT`: Port for HTTP/SSE modes (default: 8000)
- `MCP_SERVER_HOST`: Host for HTTP/SSE modes (default: 127.0.0.1)

### Example
```bash
export PIXELHUB_BASE_URL=http://your-pixelhub-server:8080
export MCP_SERVER_PORT=9000
uv run mcp-server-pixelhub -t streamable-http
```

## Available Tools

### 1. list_tags

Lists all available tags in the PixelHub system with their usage counts.

**Parameters:**
- `page` (int, optional): Page number for pagination (default: 1)
- `limit` (int, optional): Number of tags per page (default: 50, max: 100)

**Example Usage:**
```
List all tags in the system
```

**Response:**
Returns a formatted list of tags with their usage counts, sorted by frequency.

### 2. search_images_by_tags

Searches for images using tag-based relevance ranking (OR logic).

**Parameters:**
- `tags` (List[str], required): List of tags to search for
- `page` (int, optional): Page number for pagination (default: 1)  
- `limit` (int, optional): Number of images per page (default: 20, max: 100)

**Example Usage:**
```
Find images related to nature and mountains
```

**Response:**
Returns images ranked by relevance (number of matching tags), with detailed information including URLs, descriptions, tags, and match counts.

## Integration with MCP Clients

### Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "pixelhub": {
      "command": "uv",
      "args": ["run", "mcp-server-pixelhub"],
      "cwd": "/path/to/PixelHub/mcp",
      "env": {
        "PIXELHUB_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

### Cursor

Add to your Cursor MCP configuration:

```json
{
  "mcpServers": {
    "pixelhub": {
      "command": "uvx",
      "args": [
        "--from", 
        "git+https://github.com/vaaandark/PixelHub#subdirectory=mcp",
        "mcp-server-pixelhub"
      ],
      "env": {
        "PIXELHUB_BASE_URL": "http://localhost:8080"
      }
    }
  }
}
```

### Remote MCP Server

For remote deployment, use URL-based configuration:

```json
{
  "mcpServers": {
    "pixelhub-remote": {
      "url": "http://your-server:8000/mcp"
    }
  }
}
```

Or with SSE transport:

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

## Example Workflows

### 1. Semantic Image Discovery

**User:** "I need some photos of mountains at sunset for my presentation"

**LLM Process:**
1. Calls `list_tags` to see available tags
2. Identifies relevant tags like "mountains", "sunset", "landscape", "nature"
3. Calls `search_images_by_tags` with those tags
4. Returns images ranked by relevance

### 2. Content Exploration

**User:** "What types of images do I have in my collection?"

**LLM Process:**
1. Calls `list_tags` to get all available tags
2. Analyzes tag patterns and frequencies
3. Provides insights about the image collection

### 3. Specific Theme Search

**User:** "Show me urban photography with good lighting"

**LLM Process:**
1. Calls `list_tags` to find relevant tags
2. Searches for tags like "urban", "city", "lighting", "photography"
3. Returns ranked results with detailed information

## Troubleshooting

### Common Issues

1. **Connection Error**: Make sure PixelHub server is running
2. **Permission Denied**: Check log directory permissions
3. **Module Not Found**: Run `uv sync` to install dependencies

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=DEBUG
uv run mcp-server-pixelhub
```

### Testing

Test the API connection:
```bash
python run_example.py
```

## API Compatibility

This MCP server is compatible with PixelHub API v1 and supports:
- `/api/v1/tags` - List tags endpoint
- `/api/v1/search/relevance` - Relevance search endpoint

Make sure your PixelHub server has these endpoints available.
