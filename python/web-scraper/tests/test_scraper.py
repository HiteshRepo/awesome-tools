"""Tests for web_scraper — scrape logic + JSON-RPC server."""
from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path
from unittest.mock import patch

import httpx
import pytest
import respx

from web_scraper.scraper import ScrapeError, scrape
from web_scraper.__main__ import handle


# ── scraper.py unit tests ────────────────────────────────────────────────────

SAMPLE_HTML = """
<html><head><title>Test</title></head><body>
<nav>Skip me</nav>
<article>
  <h1>Ollama vs vLLM</h1>
  <p>Ollama bundles model management and serving in one binary.</p>
  <p>vLLM uses PagedAttention for high-throughput inference.</p>
</article>
<footer>Skip me too</footer>
</body></html>
"""


@respx.mock
def test_scrape_returns_main_content():
    respx.get("https://example.com/article").mock(
        return_value=httpx.Response(200, text=SAMPLE_HTML)
    )
    result = scrape("https://example.com/article")
    assert "Ollama" in result
    assert "vLLM" in result


@respx.mock
def test_scrape_raises_on_http_error():
    respx.get("https://example.com/404").mock(
        return_value=httpx.Response(404)
    )
    with pytest.raises(ScrapeError, match="HTTP 404"):
        scrape("https://example.com/404")


@respx.mock
def test_scrape_raises_on_network_error():
    respx.get("https://unreachable.invalid/").mock(side_effect=httpx.ConnectError("no route"))
    with pytest.raises(ScrapeError, match="Network error"):
        scrape("https://unreachable.invalid/")


# ── __main__.py JSON-RPC handler unit tests ──────────────────────────────────

def test_tools_list_returns_scrape_tool():
    resp = handle({"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}})
    assert resp["id"] == 1
    names = [t["name"] for t in resp["result"]["tools"]]
    assert "scrape" in names


def test_tools_call_unknown_tool_returns_error():
    resp = handle({"jsonrpc": "2.0", "id": 2, "method": "tools/call",
                   "params": {"name": "fly", "arguments": {}}})
    assert "error" in resp
    assert "fly" in resp["error"]["message"]


def test_tools_call_missing_url_returns_error():
    resp = handle({"jsonrpc": "2.0", "id": 3, "method": "tools/call",
                   "params": {"name": "scrape", "arguments": {}}})
    assert "error" in resp
    assert "url" in resp["error"]["message"]


def test_tools_call_scrape_error_returns_json_rpc_error():
    with patch("web_scraper.__main__.scrape", side_effect=ScrapeError("HTTP 403")):
        resp = handle({"jsonrpc": "2.0", "id": 4, "method": "tools/call",
                       "params": {"name": "scrape", "arguments": {"url": "https://x.com"}}})
    assert "error" in resp
    assert "HTTP 403" in resp["error"]["message"]


@respx.mock
def test_tools_call_success_returns_content():
    respx.get("https://example.com/article").mock(
        return_value=httpx.Response(200, text=SAMPLE_HTML)
    )
    resp = handle({"jsonrpc": "2.0", "id": 5, "method": "tools/call",
                   "params": {"name": "scrape", "arguments": {"url": "https://example.com/article"}}})
    assert "result" in resp
    assert "Ollama" in resp["result"]["content"]


def test_unknown_method_returns_error():
    resp = handle({"jsonrpc": "2.0", "id": 6, "method": "nope", "params": {}})
    assert "error" in resp
    assert resp["error"]["code"] == -32601


# ── end-to-end JSON-RPC round-trip via subprocess ────────────────────────────

def test_jsonrpc_round_trip_via_subprocess():
    """Spawn the server as a subprocess, send a tools/list request, verify response."""
    server_cmd = [sys.executable, "-m", "web_scraper"]
    request = json.dumps({"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}) + "\n"

    proc = subprocess.run(
        server_cmd,
        input=request,
        capture_output=True,
        text=True,
        timeout=10,
    )

    assert proc.returncode == 0
    resp = json.loads(proc.stdout.strip())
    assert resp["id"] == 1
    names = [t["name"] for t in resp["result"]["tools"]]
    assert "scrape" in names
