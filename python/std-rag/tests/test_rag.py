"""Unit tests for std-rag (no external services)."""
import tempfile
from unittest.mock import MagicMock

import pytest

from std_rag import TextChunker, Chunk, build_rag_prompt, RAGPipeline, RAGResponse
from std_llm_client import Message, LLMResponse
from std_vector_store import Document


# ---- TextChunker ----

def test_chunker_empty():
    chunker = TextChunker()
    assert chunker.chunk("", document_id="d1") == []


def test_chunker_single_chunk():
    chunker = TextChunker(chunk_size=10, chunk_overlap=2)
    chunks = chunker.chunk("hello world", document_id="d1")
    assert len(chunks) == 1
    assert chunks[0].content == "hello world"
    assert chunks[0].chunk_index == 0


def test_chunker_multiple_chunks():
    words = ["w"] * 20
    text = " ".join(words)
    chunker = TextChunker(chunk_size=10, chunk_overlap=2)
    chunks = chunker.chunk(text, document_id="d1")
    assert len(chunks) > 1
    # Each chunk must have at most chunk_size words
    for c in chunks:
        assert len(c.content.split()) <= 10


def test_chunker_invalid_overlap():
    with pytest.raises(ValueError):
        TextChunker(chunk_size=5, chunk_overlap=5)


# ---- build_rag_prompt ----

def test_build_rag_prompt():
    chunks = [
        Chunk(content="Paris is the capital of France.", document_id="doc1", chunk_index=0),
        Chunk(content="France is in Europe.", document_id="doc1", chunk_index=1),
    ]
    messages = build_rag_prompt("What is the capital of France?", chunks)
    assert len(messages) == 2
    assert messages[0].role == "system"
    assert messages[1].role == "user"
    assert "[1]" in messages[1].content
    assert "[2]" in messages[1].content
    assert "capital of France" in messages[1].content


# ---- RAGPipeline ----

def _make_pipeline(store=None):
    llm = MagicMock()
    llm.complete.return_value = LLMResponse(
        content="Paris", model="mock", usage={"input_tokens": 5, "output_tokens": 1}
    )
    embedder = MagicMock()
    embedder.embed.return_value = [[0.1, 0.2, 0.3]]
    embedder.embed_one.return_value = [0.1, 0.2, 0.3]

    if store is None:
        from std_vector_store import create_vector_store
        import tempfile
        tmpdir = tempfile.mkdtemp()
        store = create_vector_store("chroma", collection_name="rag_test", persist_directory=tmpdir)

    return RAGPipeline(llm=llm, embedder=embedder, store=store), llm, embedder


def test_ingest_returns_chunk_count():
    pipeline, _, embedder = _make_pipeline()
    docs = [Document(content="Hello world. " * 5, id="d1")]
    count = pipeline.ingest(docs)
    assert count >= 1
    embedder.embed.assert_called()


def test_query_returns_rag_response():
    pipeline, llm, embedder = _make_pipeline()
    docs = [Document(content="Paris is the capital of France.", id="d1")]
    pipeline.ingest(docs)

    result = pipeline.query("What is the capital?")
    assert isinstance(result, RAGResponse)
    assert result.answer == "Paris"
    llm.complete.assert_called_once()


@pytest.mark.asyncio
async def test_aquery_returns_rag_response():
    import asyncio
    from unittest.mock import AsyncMock

    llm = MagicMock()
    llm.acomplete = AsyncMock(return_value=LLMResponse(
        content="Async Paris", model="mock", usage={}
    ))
    embedder = MagicMock()
    embedder.embed.return_value = [[0.1, 0.2, 0.3]]
    embedder.embed_one.return_value = [0.1, 0.2, 0.3]
    embedder.aembed_one = AsyncMock(return_value=[0.1, 0.2, 0.3])

    import tempfile
    from std_vector_store import create_vector_store
    tmpdir = tempfile.mkdtemp()
    store = create_vector_store("chroma", collection_name="async_test", persist_directory=tmpdir)

    pipeline = RAGPipeline(llm=llm, embedder=embedder, store=store)
    docs = [Document(content="Paris is in France.", id="d1")]
    pipeline.ingest(docs)

    result = await pipeline.aquery("Where is Paris?")
    assert result.answer == "Async Paris"
