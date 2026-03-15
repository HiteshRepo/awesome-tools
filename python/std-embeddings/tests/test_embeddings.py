"""Unit tests for std-embeddings (no API calls)."""
from unittest.mock import MagicMock, patch, AsyncMock

import pytest

from std_embeddings import EmbeddingConfig, create_embedding_provider
from std_embeddings.base import EmbeddingProvider


def test_embedding_config():
    cfg = EmbeddingConfig(provider="openai", model="text-embedding-3-small", api_key="k")
    assert cfg.provider == "openai"
    assert cfg.api_key == "k"


def test_create_embedding_provider_unknown():
    cfg = EmbeddingConfig(provider="bad", model="x")
    with pytest.raises(ValueError, match="Unknown provider"):
        create_embedding_provider(cfg)


@patch("std_embeddings.providers.openai._openai.OpenAI")
@patch("std_embeddings.providers.openai._openai.AsyncOpenAI")
def test_openai_embed(mock_async_cls, mock_sync_cls):
    mock_item = MagicMock()
    mock_item.embedding = [0.1, 0.2, 0.3]
    mock_item.index = 0
    mock_response = MagicMock()
    mock_response.data = [mock_item]
    mock_sync_cls.return_value.embeddings.create.return_value = mock_response

    cfg = EmbeddingConfig(provider="openai", model="text-embedding-3-small", api_key="test")
    provider = create_embedding_provider(cfg)
    result = provider.embed(["hello"])

    assert result == [[0.1, 0.2, 0.3]]


@patch("std_embeddings.providers.openai._openai.OpenAI")
@patch("std_embeddings.providers.openai._openai.AsyncOpenAI")
def test_openai_embed_one(mock_async_cls, mock_sync_cls):
    mock_item = MagicMock()
    mock_item.embedding = [0.4, 0.5]
    mock_item.index = 0
    mock_response = MagicMock()
    mock_response.data = [mock_item]
    mock_sync_cls.return_value.embeddings.create.return_value = mock_response

    cfg = EmbeddingConfig(provider="openai", model="text-embedding-3-small", api_key="test")
    provider = create_embedding_provider(cfg)
    result = provider.embed_one("world")

    assert result == [0.4, 0.5]


@patch("std_embeddings.providers.voyage.voyageai.Client")
def test_voyage_embed(mock_cls):
    mock_result = MagicMock()
    mock_result.embeddings = [[0.1, 0.2]]
    mock_cls.return_value.embed.return_value = mock_result

    cfg = EmbeddingConfig(provider="voyage", model="voyage-3", api_key="test")
    provider = create_embedding_provider(cfg)
    result = provider.embed(["hi"])

    assert result == [[0.1, 0.2]]
