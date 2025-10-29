"""
Tags related tool functions
"""

from typing import List

from mcp import types
from pydantic import Field

from mcp_server_pixelhub.common.client import make_request
from mcp_server_pixelhub.common.errors import handle_error
from mcp_server_pixelhub.tools import mcp


@mcp.tool(
    name="list_tags",
    description="List all available tags in the PixelHub system with their usage counts, sorted by usage frequency",
)
async def list_tags(
    page: int = Field(
        default=1,
        description="Page number for pagination, default is 1",
    ),
    limit: int = Field(
        default=50,
        description="Number of tags per page, default is 50, maximum is 100",
    ),
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """
    List all available tags in the PixelHub system.
    
    This tool retrieves all tags that have been used to categorize images,
    along with their usage counts. Tags are sorted by usage frequency (most used first).
    This is useful for understanding what types of images are available in the system
    and for finding relevant tags to use in image searches.
    """
    try:
        # Validate parameters
        if page < 1:
            page = 1
        if limit < 1 or limit > 100:
            limit = 50
            
        # Make request to PixelHub API
        response = make_request("GET", "/tags", {
            "page": page,
            "limit": limit
        })
        
        if not response or response.get("code") != 200:
            return handle_error("list_tags")
            
        data = response.get("data", {})
        tags = data.get("tags", [])
        total = data.get("total", 0)
        current_page = data.get("current_page", page)
        
        if not tags:
            return [types.TextContent(
                type="text", 
                text="No tags found in the system."
            )]
        
        # Format the response
        result_text = f"Found {total} tags in total (showing page {current_page}):\n\n"
        
        for i, tag in enumerate(tags, 1):
            tag_name = tag.get("name", "Unknown")
            tag_count = tag.get("count", 0)
            result_text += f"{i}. '{tag_name}' - used {tag_count} time{'s' if tag_count != 1 else ''}\n"
        
        if total > len(tags):
            remaining = total - (current_page - 1) * limit - len(tags)
            if remaining > 0:
                result_text += f"\n... and {remaining} more tags. Use page={current_page + 1} to see more."
        
        return [types.TextContent(type="text", text=result_text)]
        
    except Exception as e:
        return handle_error("list_tags", e)
