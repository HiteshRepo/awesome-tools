"""ElasticAgent: Claude-powered agentic loop over Elastic MCP Docker server."""

import asyncio
from contextlib import AsyncExitStack
from typing import Any

import anthropic
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

from .config import ElasticAgentConfig

_MAX_ITERATIONS_MSG = "Max iterations reached without a final answer."


class ElasticAgent:
    """Run natural-language queries against Elasticsearch via Claude + Elastic MCP.

    Usage (sync):
        config = ElasticAgentConfig(es_url="...", es_api_key="...")
        with ElasticAgent(config) as agent:
            print(agent.query("List all indices"))

    Usage (async):
        async with ElasticAgent(config) as agent:
            print(await agent.aquery("List all indices"))
    """

    def __init__(self, config: ElasticAgentConfig) -> None:
        self._config = config
        self._client = anthropic.Anthropic(api_key=config.anthropic_api_key)
        self._session: ClientSession | None = None
        self._tools: list[dict[str, Any]] = []
        self._exit_stack = AsyncExitStack()
        self._loop: asyncio.AbstractEventLoop | None = None

    # ── sync context manager ──────────────────────────────────────────────────

    def __enter__(self) -> "ElasticAgent":
        self._loop = asyncio.new_event_loop()
        self._loop.run_until_complete(self._setup())
        return self

    def __exit__(self, *_: Any) -> None:
        if self._loop is not None:
            self._loop.run_until_complete(self._teardown())
            self._loop.close()
            self._loop = None

    # ── async context manager ─────────────────────────────────────────────────

    async def __aenter__(self) -> "ElasticAgent":
        await self._setup()
        return self

    async def __aexit__(self, *_: Any) -> None:
        await self._teardown()

    # ── lifecycle ─────────────────────────────────────────────────────────────

    async def _setup(self) -> None:
        server_params = StdioServerParameters(
            command="docker",
            args=[
                "run", "-i", "--rm",
                "-e", "ES_URL",
                "-e", "ES_API_KEY",
                self._config.docker_image,
                "stdio",
            ],
            env={
                "ES_URL": self._config.es_url,
                "ES_API_KEY": self._config.es_api_key,
            },
        )
        read, write = await self._exit_stack.enter_async_context(stdio_client(server_params))
        session = await self._exit_stack.enter_async_context(ClientSession(read, write))
        await session.initialize()
        self._session = session
        self._tools = await self._load_tools()

    async def _teardown(self) -> None:
        await self._exit_stack.aclose()
        self._session = None

    async def _load_tools(self) -> list[dict[str, Any]]:
        assert self._session is not None
        result = await self._session.list_tools()
        return [
            {
                "name": tool.name,
                "description": tool.description or "",
                "input_schema": tool.inputSchema,
            }
            for tool in result.tools
        ]

    # ── query ─────────────────────────────────────────────────────────────────

    def query(self, question: str) -> str:
        """Run a natural-language query synchronously. Must be used inside `with` block."""
        if self._loop is None:
            raise RuntimeError("ElasticAgent must be used as a context manager")
        return self._loop.run_until_complete(self.aquery(question))

    async def aquery(self, question: str) -> str:
        """Run a natural-language query asynchronously."""
        if self._session is None:
            raise RuntimeError("ElasticAgent must be used as a context manager")

        messages: list[dict[str, Any]] = [{"role": "user", "content": question}]

        for _ in range(self._config.max_iterations):
            response = self._client.messages.create(
                model=self._config.model,
                max_tokens=self._config.max_tokens,
                tools=self._tools,
                messages=messages,
            )

            if response.stop_reason != "tool_use":
                for block in response.content:
                    if hasattr(block, "text"):
                        return block.text  # type: ignore[no-any-return]
                return ""

            # Append assistant turn
            messages.append({"role": "assistant", "content": response.content})

            # Execute tool calls and collect results
            tool_results: list[dict[str, Any]] = []
            for block in response.content:
                if block.type == "tool_use":
                    mcp_result = await self._session.call_tool(block.name, block.input)
                    content_text = mcp_result.content[0].text if mcp_result.content else ""
                    tool_results.append({
                        "type": "tool_result",
                        "tool_use_id": block.id,
                        "content": content_text,
                    })

            messages.append({"role": "user", "content": tool_results})

        return _MAX_ITERATIONS_MSG
