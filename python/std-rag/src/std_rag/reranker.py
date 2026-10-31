"""Hybrid reranker: merges BM25 + dense results and scores by recency × relevance."""
from __future__ import annotations

from dataclasses import dataclass, field

from std_rag.bm25_retriever import BM25Result


@dataclass
class RankedResult:
    content: str
    document_id: str
    score: float          # final combined score
    metadata: dict = field(default_factory=dict)


class Reranker:
    """Merge BM25 and dense candidates, rerank by recency × relevance."""

    def __init__(
        self,
        recency_weight: float = 0.4,
        relevance_weight: float = 0.6,
    ) -> None:
        self._recency_weight = recency_weight
        self._relevance_weight = relevance_weight

    def rerank(
        self,
        bm25_results: list[BM25Result],
        dense_results: list,          # SearchResult from std_vector_store
        timestamps: dict[str, float] | None = None,
    ) -> list[RankedResult]:
        """
        Merge and rerank BM25 + dense results.

        Args:
            bm25_results:  BM25Result list (score = raw BM25 score)
            dense_results: SearchResult list from std_vector_store
                           (each has .document.id, .document.content,
                            .document.metadata, .score)
            timestamps:    {document_id: unix_timestamp} for recency weighting.
                           When None, all entries receive equal recency (0.5).

        Returns:
            Deduplicated list sorted by combined score descending.
        """
        # Collect per-document raw scores from each retriever
        bm25_scores: dict[str, float] = {}
        dense_scores: dict[str, float] = {}
        content_map: dict[str, str] = {}
        meta_map: dict[str, dict] = {}

        for r in bm25_results:
            bm25_scores[r.document_id] = max(bm25_scores.get(r.document_id, 0.0), r.score)
            content_map[r.document_id] = r.content
            meta_map[r.document_id] = r.metadata

        for r in dense_results:
            doc_id = r.document.id
            dense_scores[doc_id] = max(dense_scores.get(doc_id, 0.0), r.score)
            content_map[doc_id] = r.document.content
            meta_map[doc_id] = dict(r.document.metadata)

        all_ids = set(bm25_scores) | set(dense_scores)
        if not all_ids:
            return []

        # Normalise each score set independently to [0, 1]
        def _normalise(scores: dict[str, float]) -> dict[str, float]:
            if not scores:
                return {}
            max_val = max(scores.values())
            if max_val == 0:
                return {k: 0.0 for k in scores}
            return {k: v / max_val for k, v in scores.items()}

        norm_bm25 = _normalise(bm25_scores)
        norm_dense = _normalise(dense_scores)

        # Recency scores
        if timestamps:
            ts_values = [timestamps.get(doc_id, 0.0) for doc_id in all_ids]
            ts_min, ts_max = min(ts_values), max(ts_values)
            ts_range = ts_max - ts_min if ts_max != ts_min else 1.0
            recency: dict[str, float] = {
                doc_id: (timestamps.get(doc_id, 0.0) - ts_min) / ts_range
                for doc_id in all_ids
            }
        else:
            recency = {doc_id: 0.5 for doc_id in all_ids}

        # Combine: relevance = mean of available normalised scores
        results: list[RankedResult] = []
        for doc_id in all_ids:
            rel_scores = [s for s in [norm_bm25.get(doc_id), norm_dense.get(doc_id)] if s is not None]
            relevance = sum(rel_scores) / len(rel_scores) if rel_scores else 0.0
            combined = self._recency_weight * recency[doc_id] + self._relevance_weight * relevance
            results.append(RankedResult(
                content=content_map[doc_id],
                document_id=doc_id,
                score=combined,
                metadata=meta_map.get(doc_id, {}),
            ))

        results.sort(key=lambda r: r.score, reverse=True)
        return results
