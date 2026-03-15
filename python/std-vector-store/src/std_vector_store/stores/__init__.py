from ..base import VectorStore


def create_vector_store(backend: str = "chroma", **kwargs) -> VectorStore:
    if backend == "chroma":
        from .chroma import ChromaStore
        return ChromaStore(**kwargs)
    elif backend == "pinecone":
        from .pinecone import PineconeStore
        return PineconeStore(**kwargs)
    else:
        raise ValueError(f"Unknown backend: {backend!r}")
