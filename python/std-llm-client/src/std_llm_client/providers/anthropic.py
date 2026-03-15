from typing import Any

import anthropic

from ..base import LLMClient, Message, LLMResponse
from ..config import LLMConfig


class AnthropicClient(LLMClient):
    def __init__(self, config: LLMConfig) -> None:
        self._config = config
        self._client = anthropic.Anthropic(api_key=config.api_key)
        self._async_client = anthropic.AsyncAnthropic(api_key=config.api_key)

    def _split_messages(self, messages: list[Message]) -> tuple[str | None, list[dict]]:
        system = None
        chat: list[dict] = []
        for m in messages:
            if m.role == "system":
                system = m.content
            else:
                chat.append({"role": m.role, "content": m.content})
        return system, chat

    def complete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        system, chat = self._split_messages(messages)
        kwargs.setdefault("max_tokens", self._config.max_tokens)
        extra: dict[str, Any] = {}
        if system:
            extra["system"] = system
        response = self._client.messages.create(
            model=self._config.model,
            messages=chat,
            **extra,
            **kwargs,
        )
        content = response.content[0].text if response.content else ""
        usage = {
            "input_tokens": response.usage.input_tokens,
            "output_tokens": response.usage.output_tokens,
        }
        return LLMResponse(content=content, model=response.model, usage=usage)

    async def acomplete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        system, chat = self._split_messages(messages)
        kwargs.setdefault("max_tokens", self._config.max_tokens)
        extra: dict[str, Any] = {}
        if system:
            extra["system"] = system
        response = await self._async_client.messages.create(
            model=self._config.model,
            messages=chat,
            **extra,
            **kwargs,
        )
        content = response.content[0].text if response.content else ""
        usage = {
            "input_tokens": response.usage.input_tokens,
            "output_tokens": response.usage.output_tokens,
        }
        return LLMResponse(content=content, model=response.model, usage=usage)
