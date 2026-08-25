from __future__ import annotations

import argparse

import uvicorn


def main() -> None:
    """CLI entrypoint for local MemLore development."""

    parser = argparse.ArgumentParser(prog="memlore", description="MemLore CLI")
    sub = parser.add_subparsers(dest="command", required=True)

    serve = sub.add_parser("serve", help="Run the local REST API")
    serve.add_argument("--host", default="127.0.0.1")
    serve.add_argument("--port", type=int, default=8000)

    args = parser.parse_args()
    if args.command == "serve":
        uvicorn.run(
            "memlore.adapters.rest.app:create_app",
            factory=True,
            host=args.host,
            port=args.port,
        )


if __name__ == "__main__":
    main()
