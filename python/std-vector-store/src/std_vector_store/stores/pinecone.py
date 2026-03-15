from ..base import VectorStore
from ..document import Document, SearchResult


class PineconeStore(VectorStore):
    """Pinecone-backed vector store. Requires `pinecone-client>=3.0.0`."""

    def __init__(
        self,
        api_key: str,
        index_name: str,
        namespace: str = "",
    ) -> None:
        try:
            from pinecone import Pinecone
        except ImportError as e:
            raise ImportError(
                "pinecone-client is required for PineconeStore. "
                "Install with: pip install 'std-vector-store[pinecone]'"
            ) from e
        pc = Pinecone(api_key=api_key)
        self._index = pc.Index(index_name)
        self._namespace = namespace

    def add(self, documents: list[Document], embeddings: list[list[float]]) -> None:
        vectors = [
            {"id": d.id, "values": emb, "metadata": {"content": d.content, **d.metadata}}
            for d, emb in zip(documents, embeddings)
        ]
        self._index.upsert(vectors=vectors, namespace=self._namespace)

    def search(self, query_embedding: list[float], top_k: int = 5) -> list[SearchResult]:
        response = self._index.query(
            vector=query_embedding,
            top_k=top_k,
            namespace=self._namespace,
            include_metadata=True,
        )
        results: list[SearchResult] = []
        for match in response.matches:
            meta = dict(match.metadata or {})
            content = meta.pop("content", "")
            doc = Document(content=content, metadata=meta, id=match.id)
            results.append(SearchResult(document=doc, score=match.score))
        return results

    def delete(self, ids: list[str]) -> None:
        self._index.delete(ids=ids, namespace=self._namespace)

    def clear(self) -> None:
        self._index.delete(delete_all=True, namespace=self._namespace)

    def count(self) -> int:
        stats = self._index.describe_index_stats()
        ns = stats.namespaces.get(self._namespace)
        return ns.vector_count if ns else 0
