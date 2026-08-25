from __future__ import annotations

import argparse
import sys

import uvicorn

from memlore.adapters.mcp.server import configure_mcp_logging, create_mcp_server
from memlore.bootstrap.container import build_container

_GO_DEPRECATION = (
    "note: Python memlore is legacy for governance adapters; "
    "prefer Go: `memlore {command}` (build via scripts/install-memlore.sh) "
    "or `go run ./cmd/memlore {command}` from the MemLore repo.\n"
)


def _warn_go_preferred(command: str) -> None:
    sys.stderr.write(_GO_DEPRECATION.format(command=command))


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
        _warn_go_preferred("serve")
        uvicorn.run(
            "memlore.adapters.rest.app:create_app",
            factory=True,
            host=args.host,
            port=args.port,
        )
        return
    if args.command == "mcp":
        _warn_go_preferred("mcp")
        configure_mcp_logging()
        server = create_mcp_server(build_container())
        server.run(transport="stdio")


if __name__ == "__main__":
    main()
