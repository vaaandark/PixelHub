"""
PixelHub MCP Server.

This server provides MCP tools to interact with PixelHub image hosting service.
It enables semantic image discovery through tag management and relevance-based search.
"""

import argparse

from mcp_server_pixelhub.common.logs import LOG
from mcp_server_pixelhub.tools import mcp, tags, images


def main():
    parser = argparse.ArgumentParser(description="Run the PixelHub MCP Server")
    parser.add_argument(
        "--transport",
        "-t",
        choices=["sse", "stdio", "streamable-http"],
        default="stdio",
        help="Transport protocol to use (sse, stdio, or streamable-http)",
    )

    args = parser.parse_args()
    LOG.info(
        f"Including tool types: {tags.__name__}, {images.__name__}"
    )
    LOG.info(f"Starting PixelHub MCP Server with {args.transport} transport")

    mcp.run(transport=args.transport)


if __name__ == "__main__":
    main()
