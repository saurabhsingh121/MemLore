from __future__ import annotations

import argparse

import uvicorn

from memlore.adapters.mcp.server import configure_mcp_logging, create_mcp_server
from memlore.bootstrap.container import build_container


def main() -> None:
    """CLI entrypoint for local MemLore development."""

    parser = argparse.ArgumentParser(prog="memlore", description="MemLore CLI")
    sub = parser.add_subparsers(dest="command", required=True)

    serve = sub.add_parser("serve", help="Run the local REST API")
    serve.add_argument("--host", default="127.0.0.1")
    serve.add_argument("--port", type=int, default=8000)

    sub.add_parser(
        "mcp",
        help="Run the local stdio MCP server for coding-agent attachment",
    )

    args = parser.parse_args()
    if args.command == "serve":
        uvicorn.run(
            "memlore.adapters.rest.app:create_app",
            factory=True,
            host=args.host,
            port=args.port,
        )
        return
    if args.command == "mcp":
        configure_mcp_logging()
        server = create_mcp_server(build_container())
        server.run(transport="stdio")


if __name__ == "__main__":
    main()
