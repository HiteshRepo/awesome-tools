"""Unit tests for std-vector-store (no external services)."""
import tempfile

import pytest

from std_vector_store import Document, SearchResult, create_vector_store
from std_vector_store.document import Document


def test_document_defaults():
    doc = Document(content="hello")
    assert doc.content == "hello"
    assert isinstance(doc.id, str)
    assert doc.metadata == {}


def test_search_result():
    doc = Document(content="test")
    sr = SearchResult(document=doc, score=0.9)
    assert sr.score == 0.9


def test_create_vector_store_unknown_backend():
    with pytest.raises(ValueError, match="Unknown backend"):
        create_vector_store("unknown")


def test_chroma_store_add_and_search():
    with tempfile.TemporaryDirectory() as tmpdir:
        store = create_vector_store("chroma", collection_name="test", persist_directory=tmpdir)

        docs = [
            Document(content="apple fruit", id="1"),
            Document(content="banana fruit", id="2"),
            Document(content="car vehicle", id="3"),
        ]
        # 3-dim fake embeddings
        embeddings = [
            [1.0, 0.0, 0.0],
            [0.9, 0.1, 0.0],
            [0.0, 0.0, 1.0],
        ]
        store.add(docs, embeddings)
        assert store.count() == 3

        results = store.search([1.0, 0.0, 0.0], top_k=2)
        assert len(results) == 2
        # The closest doc should be "apple fruit"
        assert results[0].document.id == "1"


def test_chroma_store_delete():
    with tempfile.TemporaryDirectory() as tmpdir:
        store = create_vector_store("chroma", collection_name="test2", persist_directory=tmpdir)
        doc = Document(content="temp", id="x")
        store.add([doc], [[0.5, 0.5, 0.0]])
        assert store.count() == 1
        store.delete(["x"])
        assert store.count() == 0


def test_chroma_store_clear():
    with tempfile.TemporaryDirectory() as tmpdir:
        store = create_vector_store("chroma", collection_name="test3", persist_directory=tmpdir)
        docs = [Document(content=f"doc{i}", id=str(i)) for i in range(3)]
        store.add(docs, [[float(i), 0.0, 0.0] for i in range(3)])
        store.clear()
        assert store.count() == 0
