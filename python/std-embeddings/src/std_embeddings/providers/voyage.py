import voyageai

from ..base import EmbeddingConfig, EmbeddingProvider

_DEFAULT_MODEL = "voyage-3"


class VoyageEmbeddings(EmbeddingProvider):
    def __init__(self, config: EmbeddingConfig) -> None:
        self._model = config.model or _DEFAULT_MODEL
        self._client = voyageai.Client(api_key=config.api_key)

    def embed(self, texts: list[str]) -> list[list[float]]:
        result = self._client.embed(texts, model=self._model)
        return result.embeddings

    async def aembed(self, texts: list[str]) -> list[list[float]]:
        # voyageai does not provide a native async client; run sync in executor
        import asyncio
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self.embed, texts)
