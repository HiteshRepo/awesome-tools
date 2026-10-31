# std-rag Audit — Phase C

**Date:** 2026-10-30

## Components

| Component | Decision | Notes |
|---|---|---|
| `chunker.py` (TextChunker) | **Keep** | Word-based overlapping splitter; 8 unit tests pass |
| `pipeline.py` (RAGPipeline) | **Keep** | Dense search via std-vector-store + std-embeddings; 8 tests pass |
| BM25 retriever | **Add** | Missing — implemented as `bm25_retriever.py` using `rank-bm25==0.2.2` |
| Reranker | **Add** | Missing — implemented as `reranker.py`; merges BM25 + dense, scores by recency × relevance |

## Result

All 16 tests pass after adding BM25Retriever and Reranker. `HybridRetriever` (task 10.1) to be built on top of these.
