"""
Image search related tool functions
"""

from typing import List

from mcp import types
from pydantic import Field

from mcp_server_pixelhub.common.client import make_request
from mcp_server_pixelhub.common.errors import handle_error
from mcp_server_pixelhub.tools import mcp


@mcp.tool(
    name="search_images_by_tags",
    description="Search images by tags using relevance ranking (OR logic). Images are ranked by the number of matching tags in descending order.",
)
async def search_images_by_tags(
    tags: List[str] = Field(
        description="List of tags to search for. Images matching any of these tags will be returned, ranked by relevance (number of matching tags)",
    ),
    page: int = Field(
        default=1,
        description="Page number for pagination, default is 1",
    ),
    limit: int = Field(
        default=20,
        description="Number of images per page, default is 20, maximum is 100",
    ),
) -> list[types.TextContent | types.ImageContent | types.EmbeddedResource]:
    """
    Search for images using tag-based relevance ranking.
    
    This tool searches for images that match any of the provided tags using OR logic.
    Results are ranked by relevance - images with more matching tags appear first.
    This is perfect for semantic image discovery where you want to find images
    related to multiple concepts or themes.
    
    For example, searching for ["nature", "mountains", "sunset"] will return:
    1. Images tagged with all three tags first
    2. Images tagged with any two of the tags next  
    3. Images tagged with only one of the tags last
    """
    try:
        # Validate parameters
        if not tags:
            return [types.TextContent(
                type="text", 
                text="Error: At least one tag must be provided for search."
            )]
            
        if page < 1:
            page = 1
        if limit < 1 or limit > 100:
            limit = 20
            
        # Convert tags list to comma-separated string for API
        tags_param = ",".join(tags)
        
        # Make request to PixelHub API
        response = make_request("GET", "/search/relevance", {
            "tags": tags_param,
            "page": page,
            "limit": limit
        })
        
        if not response or response.get("code") != 200:
            return handle_error("search_images_by_tags")
            
        data = response.get("data", {})
        results = data.get("results", [])
        total = data.get("total", 0)
        current_page = data.get("current_page", page)
        
        if not results:
            return [types.TextContent(
                type="text", 
                text=f"No images found matching the tags: {', '.join(tags)}"
            )]
        
        # Format the response
        result_text = f"Found {total} images matching tags [{', '.join(tags)}] (showing page {current_page}):\n\n"
        
        for i, image in enumerate(results, 1):
            image_id = image.get("id", "Unknown")
            image_url = image.get("url", "")
            description = image.get("description", "No description")
            image_tags = image.get("tags", [])
            matched_count = image.get("matched_tag_count", 0)
            upload_date = image.get("upload_date", "Unknown")
            
            result_text += f"{i}. Image ID: {image_id}\n"
            result_text += f"   URL: {image_url}\n"
            result_text += f"   Description: {description}\n"
            result_text += f"   Tags: {', '.join(image_tags) if image_tags else 'No tags'}\n"
            result_text += f"   Matched tags: {matched_count}/{len(tags)}\n"
            result_text += f"   Upload date: {upload_date}\n\n"
        
        if total > len(results):
            remaining = total - (current_page - 1) * limit - len(results)
            if remaining > 0:
                result_text += f"... and {remaining} more images. Use page={current_page + 1} to see more."
        
        return [types.TextContent(type="text", text=result_text)]
        
    except Exception as e:
        return handle_error("search_images_by_tags", e)
