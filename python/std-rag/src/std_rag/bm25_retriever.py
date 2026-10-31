"""BM25 sparse retriever for journal entry chunks."""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import TYPE_CHECKING

from rank_bm25 import BM25Okapi

if TYPE_CHECKING:
    from std_rag.chunker import Chunk


@dataclass
class BM25Result:
    content: str
    document_id: str
    score: float
    metadata: dict = field(default_factory=dict)


class BM25Retriever:
    """Index a list of Chunks and retrieve the top-k by BM25 score."""

    def __init__(self) -> None:
        self._chunks: list[Chunk] = []
        self._bm25: BM25Okapi | None = None

    def index(self, chunks: list[Chunk]) -> None:
        """Build a BM25 index over the provided chunks."""
        self._chunks = list(chunks)
        if not chunks:
            self._bm25 = None
            return
        tokenised = [c.content.lower().split() for c in chunks]
        self._bm25 = BM25Okapi(tokenised)

    def search(self, query: str, top_k: int = 5) -> list[BM25Result]:
        """Return up to top_k chunks ranked by BM25 score."""
        if self._bm25 is None or not self._chunks:
            return []

        query_tokens = query.lower().split()
        scores = self._bm25.get_scores(query_tokens)

        ranked = sorted(
            zip(scores, self._chunks),
            key=lambda x: x[0],
            reverse=True,
        )

        return [
            BM25Result(
                content=chunk.content,
                document_id=chunk.document_id,
                score=float(score),
                metadata=dict(chunk.metadata),
            )
            for score, chunk in ranked[:top_k]
            if score > 0  # skip zero-score results
        ] or [
            # If all scores are 0 return top_k anyway (no matching terms)
            BM25Result(
                content=chunk.content,
                document_id=chunk.document_id,
                score=float(score),
                metadata=dict(chunk.metadata),
            )
            for score, chunk in ranked[:top_k]
        ]
