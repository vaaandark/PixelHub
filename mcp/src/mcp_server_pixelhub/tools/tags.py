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
    description="STEP 1: List all available tags in the PixelHub system. ALWAYS call this tool FIRST before searching for images to understand what tags are available and find the most relevant ones for your search.",
)
async def list_tags(
    page: int = Field(
        default=1,
        description="Page number for pagination, default is 1",
    ),
    limit: int = Field(
        default=1000,
        description="Number of tags per page, default is 1000, maximum is 1000",
    ),
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """
    WORKFLOW STEP 1: List all available tags in the PixelHub system.
    
    IMPORTANT: This tool should ALWAYS be called FIRST when a user asks for images.
    
    This tool retrieves all tags that have been used to categorize images,
    along with their usage counts. Tags are sorted by usage frequency (most used first).
    
    WORKFLOW:
    1. Call this tool first to see all available tags
    2. Analyze the user's request and identify relevant tags from the list
    3. Then call search_images_by_tags with the selected tags
    
    This ensures you only use tags that actually exist in the system and helps
    you find the most relevant images for the user's request.
    """
    try:
        # Validate parameters
        if page < 1:
            page = 1
        if limit < 1 or limit > 1000:
            limit = 1000
            
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
        
        result_text += f"\n\nNEXT STEP: Now use search_images_by_tags with relevant tags from this list to find images."
        
        return [types.TextContent(type="text", text=result_text)]
        
    except Exception as e:
        return handle_error("list_tags", e)
