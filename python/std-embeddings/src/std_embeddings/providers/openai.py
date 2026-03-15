import openai as _openai

from ..base import EmbeddingConfig, EmbeddingProvider

_DEFAULT_MODEL = "text-embedding-3-small"


class OpenAIEmbeddings(EmbeddingProvider):
    def __init__(self, config: EmbeddingConfig) -> None:
        model = config.model or _DEFAULT_MODEL
        self._model = model
        self._client = _openai.OpenAI(api_key=config.api_key)
        self._async_client = _openai.AsyncOpenAI(api_key=config.api_key)

    def embed(self, texts: list[str]) -> list[list[float]]:
        response = self._client.embeddings.create(model=self._model, input=texts)
        return [item.embedding for item in sorted(response.data, key=lambda x: x.index)]

    async def aembed(self, texts: list[str]) -> list[list[float]]:
        response = await self._async_client.embeddings.create(model=self._model, input=texts)
        return [item.embedding for item in sorted(response.data, key=lambda x: x.index)]
