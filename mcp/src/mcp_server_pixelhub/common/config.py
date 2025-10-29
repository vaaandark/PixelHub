import os

deploy_config = {
    "is_local": True,
    "port": int(os.getenv("MCP_SERVER_PORT", "8000")),
}

log_config = {
    "level": "INFO",
    "file": os.path.expanduser("~/.local/log/mcp/pixelhub.log"),
    "max_size": 1024000,
    "backup_count": 10,
}

pixelhub_config = {
    "base_url": os.getenv("PIXELHUB_BASE_URL", "http://localhost:8080"),
    "api_version": "v1",
}
