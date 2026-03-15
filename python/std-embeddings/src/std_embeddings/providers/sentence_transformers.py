from ..base import EmbeddingConfig, EmbeddingProvider

_DEFAULT_MODEL = "all-MiniLM-L6-v2"


class LocalEmbeddings(EmbeddingProvider):
    """Local embeddings via sentence-transformers (optional extra)."""

    def __init__(self, config: EmbeddingConfig) -> None:
        try:
            from sentence_transformers import SentenceTransformer
        except ImportError as e:
            raise ImportError(
                "sentence-transformers is required for LocalEmbeddings. "
                "Install with: pip install 'std-embeddings[local]'"
            ) from e
        model = config.model or _DEFAULT_MODEL
        self._model = SentenceTransformer(model)

    def embed(self, texts: list[str]) -> list[list[float]]:
        return self._model.encode(texts, convert_to_numpy=True).tolist()

    async def aembed(self, texts: list[str]) -> list[list[float]]:
        import asyncio
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self.embed, texts)
