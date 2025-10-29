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
    description="STEP 2: Search images by tags using relevance ranking (OR logic). ONLY call this tool AFTER calling list_tags first to get available tags. Use tags from the list_tags result.",
)
async def search_images_by_tags(
    tags: List[str] = Field(
        description="List of tags to search for. IMPORTANT: Only use tags that exist in the system (from list_tags result). Images matching any of these tags will be returned, ranked by relevance (number of matching tags)",
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
    WORKFLOW STEP 2: Search for images using tag-based relevance ranking.
    
    PREREQUISITE: You MUST call list_tags first to get available tags before using this tool.
    
    This tool searches for images that match any of the provided tags using OR logic.
    Results are ranked by relevance - images with more matching tags appear first.
    
    IMPORTANT WORKFLOW:
    1. First call list_tags to see all available tags
    2. Analyze user's request and select relevant tags from the list_tags result
    3. Then call this tool with the selected tags
    
    ONLY use tags that actually exist in the system (from list_tags output).
    
    For example, if user wants "purple iPhone photos" and list_tags shows:
    - "purple gradient background", "iPhone", "eye-level shot"
    Then search with: ["purple gradient background", "iPhone", "eye-level shot"]
    
    Results are ranked by relevance:
    1. Images tagged with all selected tags first
    2. Images tagged with multiple selected tags next  
    3. Images tagged with fewer selected tags last
    """
    try:
        # Validate parameters
        if not tags:
            return [types.TextContent(
                type="text", 
                text="Error: At least one tag must be provided for search. Please call list_tags first to see available tags, then use those tags for search."
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
                text=f"No images found matching the tags: {', '.join(tags)}. Try using different tags from the list_tags result, or call list_tags again to see all available options."
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
