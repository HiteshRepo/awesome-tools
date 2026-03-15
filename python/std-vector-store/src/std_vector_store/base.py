from abc import ABC, abstractmethod

from .document import Document, SearchResult


class VectorStore(ABC):
    @abstractmethod
    def add(self, documents: list[Document], embeddings: list[list[float]]) -> None:
        ...

    @abstractmethod
    def search(self, query_embedding: list[float], top_k: int = 5) -> list[SearchResult]:
        ...

    @abstractmethod
    def delete(self, ids: list[str]) -> None:
        ...

    @abstractmethod
    def clear(self) -> None:
        ...

    @abstractmethod
    def count(self) -> int:
        ...
