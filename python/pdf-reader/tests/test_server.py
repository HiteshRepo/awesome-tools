"""Tests for pdf-reader MCP server (task 1.5)."""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

import pytest

_SERVER = [sys.executable, "-m", "pdf_reader.server"]
_TEST_PDF = Path(__file__).parent.parent.parent.parent.parent / "pdf-reader" / "test-data" / "Hitesh-Pattanayak-Resume-2024.pdf"


def _call(method: str, params: dict | None = None, req_id: int = 1) -> dict:
    req = {"jsonrpc": "2.0", "id": req_id, "method": method}
    if params:
        req["params"] = params
    proc = subprocess.run(
        _SERVER,
        input=json.dumps(req) + "\n",
        capture_output=True,
        text=True,
        timeout=10,
    )
    return json.loads(proc.stdout.strip())


def test_tools_list_returns_pdf_read() -> None:
    resp = _call("tools/list")
    assert "result" in resp
    tools = resp["result"]["tools"]
    names = [t["name"] for t in tools]
    assert "pdf_read" in names


def test_unknown_method_returns_error() -> None:
    resp = _call("tools/nonexistent")
    assert "error" in resp


def test_missing_path_returns_error() -> None:
    resp = _call("tools/call", {"name": "pdf_read", "arguments": {}})
    assert "error" in resp


def test_nonexistent_file_returns_error() -> None:
    resp = _call("tools/call", {"name": "pdf_read", "arguments": {"path": "/no/such/file.pdf"}})
    assert "error" in resp


@pytest.mark.skipif(not _TEST_PDF.exists(), reason="Test PDF not found")
def test_pdf_read_extracts_text() -> None:
    resp = _call("tools/call", {"name": "pdf_read", "arguments": {"path": str(_TEST_PDF)}})
    assert "result" in resp, f"Expected result, got: {resp}"
    content = resp["result"]["content"]
    assert isinstance(content, list) and len(content) > 0
    text = content[0]["text"]
    assert len(text) > 10, "Expected non-trivial text from PDF"
