"""Tests for Reranker (task 1.4)."""
from __future__ import annotations

import time
from dataclasses import dataclass, field
from unittest.mock import MagicMock

import pytest
from std_rag import Reranker, BM25Result


def _bm25(doc_id: str, score: float, content: str = "text") -> BM25Result:
    return BM25Result(content=content, document_id=doc_id, score=score)


def _dense(doc_id: str, score: float, content: str = "text"):
    """Create a mock dense SearchResult."""
    r = MagicMock()
    r.document.id = doc_id
    r.document.content = content
    r.document.metadata = {}
    r.score = score
    return r


def test_newer_entry_ranks_higher_with_equal_relevance() -> None:
    """Two docs, same BM25 score, A is newer → A should rank first."""
    now = time.time()
    bm25 = [_bm25("old", 1.0), _bm25("new", 1.0)]
    dense = []
    timestamps = {"old": now - 3600, "new": now}

    reranker = Reranker(recency_weight=0.4, relevance_weight=0.6)
    results = reranker.rerank(bm25, dense, timestamps=timestamps)

    assert len(results) == 2
    assert results[0].document_id == "new", "Newer doc should rank first"


def test_deduplication_same_doc_in_both_retrievers() -> None:
    """Same doc_id in BM25 and dense → only one entry in output."""
    bm25 = [_bm25("doc-1", 0.9, "shared content")]
    dense = [_dense("doc-1", 0.8, "shared content")]

    reranker = Reranker()
    results = reranker.rerank(bm25, dense)

    doc_ids = [r.document_id for r in results]
    assert doc_ids.count("doc-1") == 1, "Duplicate doc_id must be deduplicated"


def test_empty_inputs_return_empty() -> None:
    reranker = Reranker()
    assert reranker.rerank([], []) == []


def test_no_timestamps_uses_equal_recency() -> None:
    """Without timestamps, all entries get equal recency (0.5); relevance drives ranking."""
    bm25 = [_bm25("high", 2.0), _bm25("low", 0.5)]
    reranker = Reranker(recency_weight=0.4, relevance_weight=0.6)
    results = reranker.rerank(bm25, [], timestamps=None)
    assert results[0].document_id == "high"
