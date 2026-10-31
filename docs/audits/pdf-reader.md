# pdf-reader Audit — Phase C

**Date:** 2026-10-30

## Current State

`awesome-tools/pdf-reader/` is a **Go library** (`reader.go`, `utils.go`). It extracts text from PDFs using Go's `pdfcpu` or similar. It has no MCP/JSON-RPC interface — it is a library, not a server.

## Decision: Python reimplementation

**Rationale:** Building a JSON-RPC stdio MCP server in Python using `pypdf` is faster than wrapping the Go library in a subprocess bridge. Python's `pypdf` library is well-maintained, handles standard PDFs well, and integrates cleanly with the rest of the Python tool stack.

**Implementation:** `awesome-tools/python/pdf-reader/` — minimal JSON-RPC 2.0 stdio server exposing a `pdf_read` tool.
