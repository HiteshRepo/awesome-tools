"""HybridRetriever: BM25 + dense search → reranker → single ranked list."""
from __future__ import annotations

from dataclasses import dataclass, field

from std_rag.bm25_retriever import BM25Retriever
from std_rag.reranker import Reranker, RankedResult
from std_rag.chunker import Chunk


@dataclass
class HybridDocument:
    content: str
    document_id: str
    score: float
    metadata: dict = field(default_factory=dict)


class HybridRetriever:
    """Combines BM25 sparse retrieval and dense vector search, reranked by recency × relevance."""

    def __init__(
        self,
        embedder=None,       # EmbeddingProvider from std-embeddings (optional)
        vector_store=None,   # VectorStore from std-vector-store (optional)
        recency_weight: float = 0.4,
        relevance_weight: float = 0.6,
    ) -> None:
        self._bm25 = BM25Retriever()
        self._reranker = Reranker(recency_weight=recency_weight, relevance_weight=relevance_weight)
        self._embedder = embedder
        self._store = vector_store
        self._chunks: list[Chunk] = []

    def index(self, chunks: list[Chunk], timestamps: dict[str, float] | None = None) -> None:
        """Index chunks in both BM25 and (if configured) the dense vector store."""
        self._chunks = list(chunks)
        self._timestamps = timestamps or {}
        self._bm25.index(chunks)

        if self._embedder is not None and self._store is not None:
            texts = [c.content for c in chunks]
            embeddings = self._embedder.embed(texts)
            from std_vector_store import Document
            docs = [
                Document(
                    content=c.content,
                    metadata={**c.metadata, "document_id": c.document_id},
                    id=c.document_id,
                )
                for c in chunks
            ]
            self._store.add(docs, embeddings)

    def search(self, query: str, top_k: int = 5) -> list[HybridDocument]:
        """Retrieve top_k results via BM25 + dense, then rerank."""
        bm25_results = self._bm25.search(query, top_k=top_k * 2)

        dense_results: list = []
        if self._embedder is not None and self._store is not None:
            query_emb = self._embedder.embed_one(query)
            dense_results = self._store.search(query_emb, top_k=top_k * 2)

        ranked: list[RankedResult] = self._reranker.rerank(
            bm25_results,
            dense_results,
            timestamps=self._timestamps or None,
        )

        return [
            HybridDocument(
                content=r.content,
                document_id=r.document_id,
                score=r.score,
                metadata=r.metadata,
            )
            for r in ranked[:top_k]
        ]
