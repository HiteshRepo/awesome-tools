"""Integration test for HybridRetriever (task 10.1)."""
from __future__ import annotations

import time
from std_rag import HybridRetriever, Chunk


def _chunks(entries: list[str]) -> list[Chunk]:
    return [Chunk(content=e, document_id=f"doc-{i}", chunk_index=0) for i, e in enumerate(entries)]


def test_hybrid_bm25_only_returns_relevant_entry_first() -> None:
    """Without dense retriever, BM25 + reranker should surface the topically relevant chunk."""
    entries = [f"generic entry number {i}" for i in range(9)]
    entries.append("python machine learning neural network")

    chunks = _chunks(entries)
    timestamps = {f"doc-{i}": float(i) for i in range(10)}  # doc-9 is newest

    retriever = HybridRetriever(embedder=None, vector_store=None)
    retriever.index(chunks, timestamps=timestamps)

    results = retriever.search("neural network python", top_k=3)
    assert len(results) >= 1
    top_ids = [r.document_id for r in results]
    assert "doc-9" in top_ids, f"Expected doc-9 in top results, got: {top_ids}"


def test_hybrid_no_duplicates() -> None:
    """Results should not contain the same document_id twice."""
    entries = ["alpha beta gamma", "delta epsilon"]
    retriever = HybridRetriever()
    retriever.index(_chunks(entries))
    results = retriever.search("alpha", top_k=5)
    ids = [r.document_id for r in results]
    assert len(ids) == len(set(ids)), "Duplicate document_ids in results"


def test_hybrid_empty_returns_empty() -> None:
    retriever = HybridRetriever()
    retriever.index([])
    assert retriever.search("anything") == []
