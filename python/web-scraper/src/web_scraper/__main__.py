"""Raw JSON-RPC 2.0 MCP server over stdio.

Start with:  python -m web_scraper
The agent-forge MCPClient spawns this as a subprocess and communicates
via stdin/stdout JSON-RPC 2.0 — no MCP SDK required on either side.
"""
from __future__ import annotations

import json
import sys

from web_scraper.scraper import ScrapeError, scrape

_TOOLS = [
    {
        "name": "scrape",
        "description": (
            "Fetch a URL and return its main readable text content. "
            "Strips navigation, ads, and boilerplate — returns clean prose."
        ),
        "inputSchema": {
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "The URL to fetch and extract content from.",
                },
                "timeout": {
                    "type": "number",
                    "description": "Request timeout in seconds (default: 15).",
                },
            },
            "required": ["url"],
        },
    }
]


def _ok(id_: object, result: object) -> dict:
    return {"jsonrpc": "2.0", "id": id_, "result": result}


def _err(id_: object, code: int, message: str) -> dict:
    return {"jsonrpc": "2.0", "id": id_, "error": {"code": code, "message": message}}


def handle(request: dict) -> dict:
    id_ = request.get("id")
    method = request.get("method", "")
    params = request.get("params", {})

    if method == "tools/list":
        return _ok(id_, {"tools": _TOOLS})

    if method == "tools/call":
        name = params.get("name")
        args = params.get("arguments", {})

        if name != "scrape":
            return _err(id_, -32602, f"Unknown tool: {name!r}")

        url = args.get("url", "").strip()
        if not url:
            return _err(id_, -32602, "'url' argument is required")

        timeout = float(args.get("timeout", 15.0))

        try:
            content = scrape(url, timeout=timeout)
            return _ok(id_, {"content": content, "url": url})
        except ScrapeError as exc:
            return _err(id_, -32000, str(exc))

    return _err(id_, -32601, f"Method not found: {method!r}")


def main() -> None:
    for raw in sys.stdin:
        raw = raw.strip()
        if not raw:
            continue
        try:
            request = json.loads(raw)
            response = handle(request)
        except json.JSONDecodeError as exc:
            response = _err(None, -32700, f"Parse error: {exc}")
        print(json.dumps(response), flush=True)


if __name__ == "__main__":
    main()
