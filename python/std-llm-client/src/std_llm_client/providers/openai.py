from typing import Any

import openai as _openai

from ..base import LLMClient, Message, LLMResponse
from ..config import LLMConfig


class OpenAIClient(LLMClient):
    def __init__(self, config: LLMConfig) -> None:
        self._config = config
        self._client = _openai.OpenAI(api_key=config.api_key)
        self._async_client = _openai.AsyncOpenAI(api_key=config.api_key)

    def _to_dicts(self, messages: list[Message]) -> list[dict]:
        return [{"role": m.role, "content": m.content} for m in messages]

    def complete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        kwargs.setdefault("max_tokens", self._config.max_tokens)
        kwargs.setdefault("temperature", self._config.temperature)
        response = self._client.chat.completions.create(
            model=self._config.model,
            messages=self._to_dicts(messages),
            **kwargs,
        )
        choice = response.choices[0]
        content = choice.message.content or ""
        usage = {
            "input_tokens": response.usage.prompt_tokens,
            "output_tokens": response.usage.completion_tokens,
        }
        return LLMResponse(content=content, model=response.model, usage=usage)

    async def acomplete(self, messages: list[Message], **kwargs: Any) -> LLMResponse:
        kwargs.setdefault("max_tokens", self._config.max_tokens)
        kwargs.setdefault("temperature", self._config.temperature)
        response = await self._async_client.chat.completions.create(
            model=self._config.model,
            messages=self._to_dicts(messages),
            **kwargs,
        )
        choice = response.choices[0]
        content = choice.message.content or ""
        usage = {
            "input_tokens": response.usage.prompt_tokens,
            "output_tokens": response.usage.completion_tokens,
        }
        return LLMResponse(content=content, model=response.model, usage=usage)
