"""JSON-RPC 2.0 stdio MCP server — exposes pdf_read tool."""
from __future__ import annotations

import json
import sys
from pathlib import Path


_TOOLS = [
    {
        "name": "pdf_read",
        "description": "Extract text from a PDF file.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Absolute or relative path to the PDF file."},
            },
            "required": ["path"],
        },
    }
]


def _handle(req: dict) -> dict:
    method = req.get("method", "")

    if method == "tools/list":
        return {"result": {"tools": _TOOLS}}

    if method == "tools/call":
        tool_name = req.get("params", {}).get("name", "")
        if tool_name != "pdf_read":
            return {"error": {"code": -32601, "message": f"Unknown tool: {tool_name}"}}

        path_str: str = req.get("params", {}).get("arguments", {}).get("path", "")
        if not path_str:
            return {"error": {"code": -32602, "message": "'path' argument is required"}}

        try:
            from pypdf import PdfReader  # lazy import — keeps startup fast if pypdf is missing
            reader = PdfReader(path_str)
            pages = [page.extract_text() or "" for page in reader.pages]
            text = "\n".join(pages)
            return {"result": {"content": [{"type": "text", "text": text}]}}
        except FileNotFoundError:
            return {"error": {"code": -32602, "message": f"File not found: {path_str}"}}
        except Exception as exc:  # noqa: BLE001
            return {"error": {"code": -32603, "message": str(exc)}}

    return {"error": {"code": -32601, "message": f"Method not found: {method}"}}


def main() -> None:
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
        except json.JSONDecodeError:
            continue
        resp = {"jsonrpc": "2.0", "id": req.get("id"), **_handle(req)}
        sys.stdout.write(json.dumps(resp) + "\n")
        sys.stdout.flush()


if __name__ == "__main__":
    main()
