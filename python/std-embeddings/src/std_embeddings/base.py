from abc import ABC, abstractmethod
from dataclasses import dataclass, field


@dataclass
class EmbeddingConfig:
    provider: str  # "openai" | "voyage" | "local"
    model: str
    api_key: str = ""


class EmbeddingProvider(ABC):
    @abstractmethod
    def embed(self, texts: list[str]) -> list[list[float]]:
        ...

    def embed_one(self, text: str) -> list[float]:
        return self.embed([text])[0]

    @abstractmethod
    async def aembed(self, texts: list[str]) -> list[list[float]]:
        ...

    async def aembed_one(self, text: str) -> list[float]:
        results = await self.aembed([text])
        return results[0]
