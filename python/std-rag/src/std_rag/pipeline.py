from dataclasses import dataclass, field
from uuid import uuid4

from std_llm_client import LLMClient
from std_embeddings import EmbeddingProvider
from std_vector_store import VectorStore, Document

from .chunker import TextChunker, Chunk
from .prompt import build_rag_prompt


@dataclass
class RAGResponse:
    answer: str
    sources: list[str] = field(default_factory=list)
    usage: dict = field(default_factory=dict)


class RAGPipeline:
    def __init__(
        self,
        llm: LLMClient,
        embedder: EmbeddingProvider,
        store: VectorStore,
        chunker: TextChunker | None = None,
    ) -> None:
        self._llm = llm
        self._embedder = embedder
        self._store = store
        self._chunker = chunker or TextChunker()

    def ingest(self, documents: list[Document]) -> int:
        all_chunks: list[Chunk] = []
        for doc in documents:
            chunks = self._chunker.chunk(
                text=doc.content,
                document_id=doc.id,
                metadata=doc.metadata,
            )
            all_chunks.extend(chunks)

        if not all_chunks:
            return 0

        texts = [c.content for c in all_chunks]
        embeddings = self._embedder.embed(texts)

        chunk_docs = [
            Document(
                content=c.content,
                metadata={**c.metadata, "document_id": c.document_id, "chunk_index": c.chunk_index},
                id=str(uuid4()),
            )
            for c in all_chunks
        ]
        self._store.add(chunk_docs, embeddings)
        return len(all_chunks)

    def query(self, question: str, top_k: int = 5) -> RAGResponse:
        query_emb = self._embedder.embed_one(question)
        results = self._store.search(query_emb, top_k=top_k)

        chunks = [
            Chunk(
                content=r.document.content,
                document_id=r.document.metadata.get("document_id", r.document.id),
                chunk_index=r.document.metadata.get("chunk_index", 0),
                metadata=r.document.metadata,
            )
            for r in results
        ]
        messages = build_rag_prompt(question, chunks)
        response = self._llm.complete(messages)

        sources = list({r.document.metadata.get("document_id", "") for r in results})
        return RAGResponse(answer=response.content, sources=sources, usage=response.usage)

    async def aquery(self, question: str, top_k: int = 5) -> RAGResponse:
        query_emb = await self._embedder.aembed_one(question)
        results = self._store.search(query_emb, top_k=top_k)

        chunks = [
            Chunk(
                content=r.document.content,
                document_id=r.document.metadata.get("document_id", r.document.id),
                chunk_index=r.document.metadata.get("chunk_index", 0),
                metadata=r.document.metadata,
            )
            for r in results
        ]
        messages = build_rag_prompt(question, chunks)
        response = await self._llm.acomplete(messages)

        sources = list({r.document.metadata.get("document_id", "") for r in results})
        return RAGResponse(answer=response.content, sources=sources, usage=response.usage)
