#!/usr/bin/env python3
"""
Example script to test PixelHub MCP Server functionality
"""

import asyncio
import json
from mcp_server_pixelhub.common.client import make_request


async def test_list_tags():
    """Test listing tags"""
    print("Testing list_tags...")
    try:
        response = make_request("GET", "/tags", {"page": 1, "limit": 10})
        print(f"Response: {json.dumps(response, indent=2)}")
    except Exception as e:
        print(f"Error: {e}")


async def test_search_images():
    """Test searching images by tags"""
    print("\nTesting search_images_by_tags...")
    try:
        response = make_request("GET", "/search/relevance", {
            "tags": "nature,landscape",
            "page": 1,
            "limit": 5
        })
        print(f"Response: {json.dumps(response, indent=2)}")
    except Exception as e:
        print(f"Error: {e}")


async def main():
    """Main test function"""
    print("PixelHub MCP Server Test")
    print("=" * 40)
    print("Make sure PixelHub server is running on http://localhost:8080")
    print()
    
    await test_list_tags()
    await test_search_images()


if __name__ == "__main__":
    asyncio.run(main())
