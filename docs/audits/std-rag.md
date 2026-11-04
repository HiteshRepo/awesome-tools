# std-rag Audit — Phase C

**Date:** 2026-10-30

## Components

| Component | Decision | Notes |
|---|---|---|
| `chunker.py` (TextChunker) | **Keep** | Word-based overlapping splitter; 8 unit tests pass |
| `pipeline.py` (RAGPipeline) | **Keep** | Dense search via std-vector-store + std-embeddings; 8 tests pass |
| BM25 retriever | **Add** | Missing — implemented as `bm25_retriever.py` using `rank-bm25==0.2.2` |
| Reranker | **Add** | Missing — implemented as `reranker.py`; merges BM25 + dense, scores by recency × relevance |

| `hybrid_retriever.py` (HybridRetriever) | **Add** | Built on BM25 + dense; deduplicates, applies reranker scoring |

## Result

19/19 tests pass (verified 2026-11-03): BM25Retriever, HybridRetriever, Reranker, TextChunker, RAGPipeline all confirmed. **M31 complete.**
