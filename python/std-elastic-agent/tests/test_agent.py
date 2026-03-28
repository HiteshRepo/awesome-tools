"""Unit tests for ElasticAgent using MagicMock / AsyncMock."""

import asyncio
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from std_elastic_agent import ElasticAgent, ElasticAgentConfig


# ── fixtures ──────────────────────────────────────────────────────────────────


@pytest.fixture
def config() -> ElasticAgentConfig:
    return ElasticAgentConfig(
        es_url="https://test.es.io:9200",
        es_api_key="test-es-key",
        anthropic_api_key="test-anthropic-key",
        max_iterations=3,
    )


def _make_mcp_tool(name: str = "search", description: str = "Search Elasticsearch") -> MagicMock:
    tool = MagicMock()
    tool.name = name
    tool.description = description
    tool.inputSchema = {"type": "object", "properties": {}}
    return tool


def _make_text_response(text: str) -> MagicMock:
    block = MagicMock()
    block.type = "text"
    block.text = text
    response = MagicMock()
    response.stop_reason = "end_turn"
    response.content = [block]
    return response


def _make_tool_use_response(tool_name: str, tool_id: str, tool_input: dict) -> MagicMock:
    block = MagicMock()
    block.type = "tool_use"
    block.name = tool_name
    block.id = tool_id
    block.input = tool_input
    response = MagicMock()
    response.stop_reason = "tool_use"
    response.content = [block]
    return response


def _make_agent_with_session(config: ElasticAgentConfig) -> tuple[ElasticAgent, AsyncMock]:
    """Return agent with mocked session + tools pre-loaded."""
    agent = ElasticAgent(config)
    mock_session = AsyncMock()
    agent._session = mock_session
    agent._tools = [{"name": "search", "description": "Search", "input_schema": {}}]
    return agent, mock_session


# ── config tests ──────────────────────────────────────────────────────────────


def test_config_defaults() -> None:
    cfg = ElasticAgentConfig(es_url="https://es.io", es_api_key="key", anthropic_api_key="ak")
    assert cfg.model == "claude-sonnet-4-6"
    assert cfg.max_tokens == 4096
    assert cfg.max_iterations == 10
    assert cfg.docker_image == "docker.elastic.co/mcp/elasticsearch"


def test_config_anthropic_key_from_env(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("ANTHROPIC_API_KEY", "env-key")
    cfg = ElasticAgentConfig(es_url="https://es.io", es_api_key="key")
    assert cfg.anthropic_api_key == "env-key"


# ── _load_tools tests ─────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_load_tools_converts_mcp_format(config: ElasticAgentConfig) -> None:
    agent = ElasticAgent(config)
    mock_session = AsyncMock()
    mock_session.list_tools.return_value = MagicMock(
        tools=[_make_mcp_tool("indices", "List indices")]
    )
    agent._session = mock_session

    tools = await agent._load_tools()

    assert len(tools) == 1
    assert tools[0]["name"] == "indices"
    assert tools[0]["description"] == "List indices"
    assert tools[0]["input_schema"] == {"type": "object", "properties": {}}


@pytest.mark.asyncio
async def test_load_tools_empty_description_defaults_to_empty_string(
    config: ElasticAgentConfig,
) -> None:
    agent = ElasticAgent(config)
    tool = _make_mcp_tool()
    tool.description = None
    mock_session = AsyncMock()
    mock_session.list_tools.return_value = MagicMock(tools=[tool])
    agent._session = mock_session

    tools = await agent._load_tools()
    assert tools[0]["description"] == ""


# ── aquery tests ──────────────────────────────────────────────────────────────


@pytest.mark.asyncio
async def test_aquery_returns_text_on_end_turn(config: ElasticAgentConfig) -> None:
    agent, _ = _make_agent_with_session(config)
    agent._client = MagicMock()
    agent._client.messages.create.return_value = _make_text_response("Found 3 indices.")

    result = await agent.aquery("List all indices")

    assert result == "Found 3 indices."
    agent._client.messages.create.assert_called_once()


@pytest.mark.asyncio
async def test_aquery_calls_tool_then_returns_text(config: ElasticAgentConfig) -> None:
    agent, mock_session = _make_agent_with_session(config)

    # First call: tool_use; second call: end_turn
    tool_response = _make_tool_use_response("search", "tu_001", {"query": "list"})
    final_response = _make_text_response("Here are your indices: .ds-logs")

    agent._client = MagicMock()
    agent._client.messages.create.side_effect = [tool_response, final_response]

    # Mock MCP tool result
    mcp_content = MagicMock()
    mcp_content.text = '{"indices": [".ds-logs"]}'
    mock_session.call_tool.return_value = MagicMock(content=[mcp_content])

    result = await agent.aquery("List all indices")

    assert result == "Here are your indices: .ds-logs"
    mock_session.call_tool.assert_called_once_with("search", {"query": "list"})
    assert agent._client.messages.create.call_count == 2


@pytest.mark.asyncio
async def test_aquery_tool_result_appended_to_messages(config: ElasticAgentConfig) -> None:
    agent, mock_session = _make_agent_with_session(config)

    tool_response = _make_tool_use_response("search", "tu_abc", {})
    final_response = _make_text_response("Done.")

    agent._client = MagicMock()
    agent._client.messages.create.side_effect = [tool_response, final_response]

    mcp_content = MagicMock()
    mcp_content.text = "result-text"
    mock_session.call_tool.return_value = MagicMock(content=[mcp_content])

    await agent.aquery("query")

    # Second Claude call should include tool result in messages
    second_call_messages = agent._client.messages.create.call_args_list[1][1]["messages"]
    tool_result_turn = second_call_messages[-1]
    assert tool_result_turn["role"] == "user"
    assert tool_result_turn["content"][0]["type"] == "tool_result"
    assert tool_result_turn["content"][0]["tool_use_id"] == "tu_abc"
    assert tool_result_turn["content"][0]["content"] == "result-text"


@pytest.mark.asyncio
async def test_aquery_max_iterations_returns_fallback(config: ElasticAgentConfig) -> None:
    agent, mock_session = _make_agent_with_session(config)

    # Always return tool_use so we never reach end_turn
    tool_response = _make_tool_use_response("search", "tu_x", {})
    agent._client = MagicMock()
    agent._client.messages.create.return_value = tool_response

    mcp_content = MagicMock()
    mcp_content.text = "data"
    mock_session.call_tool.return_value = MagicMock(content=[mcp_content])

    result = await agent.aquery("loop forever")

    assert "Max iterations" in result
    assert agent._client.messages.create.call_count == config.max_iterations


@pytest.mark.asyncio
async def test_aquery_empty_mcp_content_uses_empty_string(config: ElasticAgentConfig) -> None:
    agent, mock_session = _make_agent_with_session(config)

    tool_response = _make_tool_use_response("search", "tu_1", {})
    final_response = _make_text_response("ok")
    agent._client = MagicMock()
    agent._client.messages.create.side_effect = [tool_response, final_response]
    mock_session.call_tool.return_value = MagicMock(content=[])

    result = await agent.aquery("q")
    assert result == "ok"
    # Verify empty content handled gracefully
    second_messages = agent._client.messages.create.call_args_list[1][1]["messages"]
    assert second_messages[-1]["content"][0]["content"] == ""


@pytest.mark.asyncio
async def test_aquery_raises_without_context_manager(config: ElasticAgentConfig) -> None:
    agent = ElasticAgent(config)  # no __enter__
    with pytest.raises(RuntimeError, match="context manager"):
        await agent.aquery("test")


# ── sync query tests ──────────────────────────────────────────────────────────


def test_query_raises_without_context_manager(config: ElasticAgentConfig) -> None:
    agent = ElasticAgent(config)
    with pytest.raises(RuntimeError, match="context manager"):
        agent.query("test")


def test_query_delegates_to_aquery(config: ElasticAgentConfig) -> None:
    agent, _ = _make_agent_with_session(config)
    agent._client = MagicMock()
    agent._client.messages.create.return_value = _make_text_response("answer")

    # Provide a real event loop for sync path
    loop = asyncio.new_event_loop()
    agent._loop = loop
    try:
        result = agent.query("what indices exist?")
    finally:
        loop.close()
        agent._loop = None

    assert result == "answer"


# ── context manager lifecycle tests ───────────────────────────────────────────


def test_context_manager_calls_setup_and_teardown(config: ElasticAgentConfig) -> None:
    with (
        patch.object(ElasticAgent, "_setup", new_callable=AsyncMock) as mock_setup,
        patch.object(ElasticAgent, "_teardown", new_callable=AsyncMock) as mock_teardown,
    ):
        with ElasticAgent(config):
            mock_setup.assert_called_once()
            mock_teardown.assert_not_called()
        mock_teardown.assert_called_once()


@pytest.mark.asyncio
async def test_async_context_manager_calls_setup_and_teardown(config: ElasticAgentConfig) -> None:
    with (
        patch.object(ElasticAgent, "_setup", new_callable=AsyncMock) as mock_setup,
        patch.object(ElasticAgent, "_teardown", new_callable=AsyncMock) as mock_teardown,
    ):
        async with ElasticAgent(config):
            mock_setup.assert_called_once()
            mock_teardown.assert_not_called()
        mock_teardown.assert_called_once()
