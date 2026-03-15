"""Unit tests for std-llm-client (no API calls)."""
from unittest.mock import MagicMock, AsyncMock, patch

import pytest

from std_llm_client import LLMConfig, LLMResponse, Message, create_llm_client
from std_llm_client.base import LLMClient


def test_llm_config_defaults():
    cfg = LLMConfig(provider="anthropic", model="claude-sonnet-4-6", api_key="k")
    assert cfg.max_tokens == 4096
    assert cfg.temperature == 0.0


def test_message():
    m = Message(role="user", content="hello")
    assert m.role == "user"
    assert m.content == "hello"


def test_llm_response():
    r = LLMResponse(content="hi", model="m", usage={"input_tokens": 1})
    assert r.content == "hi"


def test_create_llm_client_unknown_provider():
    cfg = LLMConfig(provider="unknown", model="x", api_key="k")
    with pytest.raises(ValueError, match="Unknown provider"):
        create_llm_client(cfg)


@patch("std_llm_client.providers.anthropic.anthropic.Anthropic")
@patch("std_llm_client.providers.anthropic.anthropic.AsyncAnthropic")
def test_anthropic_client_complete(mock_async_cls, mock_sync_cls):
    mock_response = MagicMock()
    mock_response.content = [MagicMock(text="Answer")]
    mock_response.model = "claude-sonnet-4-6"
    mock_response.usage.input_tokens = 10
    mock_response.usage.output_tokens = 5
    mock_sync_cls.return_value.messages.create.return_value = mock_response

    cfg = LLMConfig(provider="anthropic", model="claude-sonnet-4-6", api_key="test")
    client = create_llm_client(cfg)
    result = client.complete([Message(role="user", content="hi")])

    assert result.content == "Answer"
    assert result.model == "claude-sonnet-4-6"
    assert result.usage["input_tokens"] == 10


@patch("std_llm_client.providers.openai._openai.OpenAI")
@patch("std_llm_client.providers.openai._openai.AsyncOpenAI")
def test_openai_client_complete(mock_async_cls, mock_sync_cls):
    mock_choice = MagicMock()
    mock_choice.message.content = "OpenAI answer"
    mock_response = MagicMock()
    mock_response.choices = [mock_choice]
    mock_response.model = "gpt-4o"
    mock_response.usage.prompt_tokens = 8
    mock_response.usage.completion_tokens = 4
    mock_sync_cls.return_value.chat.completions.create.return_value = mock_response

    cfg = LLMConfig(provider="openai", model="gpt-4o", api_key="test")
    client = create_llm_client(cfg)
    result = client.complete([Message(role="user", content="hi")])

    assert result.content == "OpenAI answer"
    assert result.usage["input_tokens"] == 8


@pytest.mark.asyncio
@patch("std_llm_client.providers.anthropic.anthropic.Anthropic")
@patch("std_llm_client.providers.anthropic.anthropic.AsyncAnthropic")
async def test_anthropic_client_acomplete(mock_async_cls, mock_sync_cls):
    mock_response = MagicMock()
    mock_response.content = [MagicMock(text="Async answer")]
    mock_response.model = "claude-sonnet-4-6"
    mock_response.usage.input_tokens = 10
    mock_response.usage.output_tokens = 5
    mock_async_cls.return_value.messages.create = AsyncMock(return_value=mock_response)

    cfg = LLMConfig(provider="anthropic", model="claude-sonnet-4-6", api_key="test")
    client = create_llm_client(cfg)
    result = await client.acomplete([Message(role="user", content="hi")])

    assert result.content == "Async answer"
