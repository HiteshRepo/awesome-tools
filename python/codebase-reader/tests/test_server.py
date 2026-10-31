"""Tests for codebase-reader MCP server (task 1.6)."""
from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path

_SERVER = [sys.executable, "-m", "codebase_reader.server"]
_SELF_DIR = str(Path(__file__).parent.parent)  # the codebase-reader package dir itself


_SRC_DIR = str(Path(__file__).parent.parent / "src")


def _call(method: str, params: dict | None = None, req_id: int = 1) -> dict:
    req = {"jsonrpc": "2.0", "id": req_id, "method": method}
    if params:
        req["params"] = params
    env = {**os.environ, "PYTHONPATH": _SRC_DIR}
    proc = subprocess.run(
        _SERVER,
        input=json.dumps(req) + "\n",
        capture_output=True,
        text=True,
        timeout=15,
        env=env,
    )
    return json.loads(proc.stdout.strip())


def test_tools_list_returns_codebase_read() -> None:
    resp = _call("tools/list")
    assert "result" in resp
    names = [t["name"] for t in resp["result"]["tools"]]
    assert "codebase_read" in names


def test_codebase_read_on_own_directory() -> None:
    """Reading the codebase-reader directory itself should return at least one file."""
    resp = _call("tools/call", {"name": "codebase_read", "arguments": {"path": _SELF_DIR}})
    assert "result" in resp, f"Expected result, got: {resp}"
    files = resp["result"]["files"]
    assert len(files) >= 1
    paths = [f["path"] for f in files]
    assert any("server.py" in p for p in paths)


def test_nonexistent_path_returns_error() -> None:
    resp = _call("tools/call", {"name": "codebase_read", "arguments": {"path": "/no/such/dir"}})
    assert "error" in resp


def test_unknown_method_returns_error() -> None:
    resp = _call("other/method")
    assert "error" in resp


def test_max_files_respected() -> None:
    resp = _call("tools/call", {"name": "codebase_read", "arguments": {"path": _SELF_DIR, "max_files": 1}})
    assert "result" in resp
    assert resp["result"]["count"] <= 1
