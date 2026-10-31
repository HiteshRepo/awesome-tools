"""JSON-RPC 2.0 stdio MCP server — exposes codebase_read tool."""
from __future__ import annotations

import json
import os
import sys
from pathlib import Path

_SKIP_DIRS = {".git", "__pycache__", ".venv", "node_modules", ".pytest_cache", "dist", "build"}
_MAX_FILE_BYTES = 100_000  # skip files larger than 100 KB

_TOOLS = [
    {
        "name": "codebase_read",
        "description": "Walk a directory and return file paths with their text content.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Directory to read."},
                "max_files": {"type": "integer", "description": "Max files to return (default 50)."},
            },
            "required": ["path"],
        },
    }
]


def _read_directory(root: str, max_files: int = 50) -> list[dict[str, str]]:
    root_path = Path(root).resolve()
    files: list[dict[str, str]] = []

    for dirpath, dirnames, filenames in os.walk(root_path):
        # Prune skipped dirs in-place so os.walk won't descend into them
        dirnames[:] = [
            d for d in dirnames
            if d not in _SKIP_DIRS and not d.startswith(".")
        ]

        for filename in filenames:
            if len(files) >= max_files:
                return files
            if filename.startswith("."):
                continue

            filepath = Path(dirpath) / filename
            try:
                if filepath.stat().st_size > _MAX_FILE_BYTES:
                    continue
                text = filepath.read_text(encoding="utf-8")
                rel = str(filepath.relative_to(root_path))
                files.append({"path": rel, "content": text})
            except (UnicodeDecodeError, OSError):
                continue  # skip binary or unreadable files

    return files


def _handle(req: dict) -> dict:
    method = req.get("method", "")

    if method == "tools/list":
        return {"result": {"tools": _TOOLS}}

    if method == "tools/call":
        tool_name = req.get("params", {}).get("name", "")
        if tool_name != "codebase_read":
            return {"error": {"code": -32601, "message": f"Unknown tool: {tool_name}"}}

        args = req.get("params", {}).get("arguments", {})
        path = args.get("path", "")
        if not path:
            return {"error": {"code": -32602, "message": "'path' argument is required"}}

        max_files = int(args.get("max_files", 50))

        if not Path(path).exists():
            return {"error": {"code": -32602, "message": f"Path not found: {path}"}}
        if not Path(path).is_dir():
            return {"error": {"code": -32602, "message": f"Not a directory: {path}"}}

        try:
            files = _read_directory(path, max_files=max_files)
            return {"result": {"files": files, "count": len(files)}}
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
