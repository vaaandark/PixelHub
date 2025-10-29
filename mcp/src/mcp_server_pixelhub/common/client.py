import httpx
from typing import Dict, Any, Optional

from mcp_server_pixelhub.common.config import pixelhub_config
from mcp_server_pixelhub.common.logs import LOG

_http_client = None


def get_pixelhub_client() -> httpx.Client:
    """Get HTTP client for PixelHub API"""
    global _http_client
    
    if _http_client is None:
        _http_client = httpx.Client(
            base_url=f"{pixelhub_config['base_url']}/api/{pixelhub_config['api_version']}",
            timeout=30.0,
            headers={
                "Content-Type": "application/json",
                "User-Agent": "PixelHub-MCP-Server/0.1.0"
            }
        )
    
    return _http_client


async def get_async_pixelhub_client() -> httpx.AsyncClient:
    """Get async HTTP client for PixelHub API"""
    return httpx.AsyncClient(
        base_url=f"{pixelhub_config['base_url']}/api/{pixelhub_config['api_version']}",
        timeout=30.0,
        headers={
            "Content-Type": "application/json",
            "User-Agent": "PixelHub-MCP-Server/0.1.0"
        }
    )


def make_request(method: str, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    """Make HTTP request to PixelHub API
    
    Args:
        method: HTTP method (GET, POST, etc.)
        endpoint: API endpoint path
        params: Query parameters or request body
        
    Returns:
        Response data as dictionary
        
    Raises:
        Exception: If request fails
    """
    try:
        client = get_pixelhub_client()
        
        if method.upper() == "GET":
            response = client.get(endpoint, params=params)
        elif method.upper() == "POST":
            response = client.post(endpoint, json=params)
        else:
            raise ValueError(f"Unsupported HTTP method: {method}")
            
        response.raise_for_status()
        return response.json()
        
    except httpx.HTTPError as e:
        LOG.error(f"HTTP error when calling {endpoint}: {e}")
        raise Exception(f"HTTP error: {e}")
    except Exception as e:
        LOG.error(f"Error when calling {endpoint}: {e}")
        raise e
