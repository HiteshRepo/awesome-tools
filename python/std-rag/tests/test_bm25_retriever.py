"""Tests for BM25Retriever (task 1.3)."""
from __future__ import annotations

import pytest
from std_rag import BM25Retriever, Chunk


def _make_chunks(contents: list[str]) -> list[Chunk]:
    return [
        Chunk(content=c, document_id=f"doc-{i}", chunk_index=0)
        for i, c in enumerate(contents)
    ]


def test_search_returns_relevant_chunks() -> None:
    """Index 10 entries; query keyword present in 2 → those 2 appear in top results."""
    contents = [f"entry about topic {i}" for i in range(8)]
    contents += ["python BM25 retrieval algorithm", "BM25 ranking implementation python"]
    chunks = _make_chunks(contents)

    retriever = BM25Retriever()
    retriever.index(chunks)
    results = retriever.search("BM25", top_k=5)

    result_ids = {r.document_id for r in results}
    assert "doc-8" in result_ids or "doc-9" in result_ids, (
        "At least one BM25-matching chunk should appear in top results"
    )


def test_search_missing_keyword_returns_without_error() -> None:
    """Query for a keyword not in any chunk — should return results without raising."""
    chunks = _make_chunks(["hello world", "foo bar baz"])
    retriever = BM25Retriever()
    retriever.index(chunks)
    results = retriever.search("zzz_nonexistent_zzz", top_k=3)
    assert isinstance(results, list)


def test_search_empty_index_returns_empty() -> None:
    retriever = BM25Retriever()
    retriever.index([])
    assert retriever.search("anything") == []


def test_search_result_fields() -> None:
    chunks = _make_chunks(["the quick brown fox"])
    retriever = BM25Retriever()
    retriever.index(chunks)
    results = retriever.search("fox", top_k=1)
    assert len(results) == 1
    r = results[0]
    assert r.content == "the quick brown fox"
    assert r.document_id == "doc-0"
    assert isinstance(r.score, float)
