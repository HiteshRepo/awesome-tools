from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any


@dataclass
class Message:
    role: str  # "user", "assistant", "system"
    content: str


@dataclass
class LLMResponse:
    content: str
    model: str
    usage: dict[str, int] = field(default_factory=dict)


class LLMClient(ABC):
    @abstractmethod
    def complete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        ...

    @abstractmethod
    async def acomplete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        ...
