import json

import chromadb

from ..base import VectorStore
from ..document import Document, SearchResult


class ChromaStore(VectorStore):
    def __init__(
        self,
        collection_name: str = "default",
        persist_directory: str = ".chroma",
    ) -> None:
        self._client = chromadb.PersistentClient(path=persist_directory)
        self._collection = self._client.get_or_create_collection(collection_name)

    def add(self, documents: list[Document], embeddings: list[list[float]]) -> None:
        if not documents:
            return
        # Chroma 1.x rejects empty metadata dicts; pass None for those
        metadatas = [d.metadata if d.metadata else None for d in documents]
        self._collection.add(
            ids=[d.id for d in documents],
            documents=[d.content for d in documents],
            embeddings=embeddings,
            metadatas=metadatas,
        )

    def search(self, query_embedding: list[float], top_k: int = 5) -> list[SearchResult]:
        results = self._collection.query(
            query_embeddings=[query_embedding],
            n_results=min(top_k, self._collection.count() or 1),
            include=["documents", "metadatas", "distances"],
        )
        output: list[SearchResult] = []
        for doc_id, content, metadata, distance in zip(
            results["ids"][0],
            results["documents"][0],
            results["metadatas"][0],
            results["distances"][0],
        ):
            doc = Document(content=content, metadata=metadata or {}, id=doc_id)
            # Chroma returns L2 distance; convert to cosine-like score (lower=better → invert)
            score = 1.0 / (1.0 + distance)
            output.append(SearchResult(document=doc, score=score))
        return output

    def delete(self, ids: list[str]) -> None:
        self._collection.delete(ids=ids)

    def clear(self) -> None:
        name = self._collection.name
        self._client.delete_collection(name)
        self._collection = self._client.get_or_create_collection(name)

    def count(self) -> int:
        return self._collection.count()
