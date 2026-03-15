import asyncio
import json
from typing import Any

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp import types as mcp_types

from .tool import get_tools
from .types import ToolResult


class BaseMCPServer:
    """Base class for stdio-based MCP servers.

    Subclasses decorate methods with ``@mcp_tool`` to expose them as tools.
    """

    def __init__(self, name: str, version: str = "0.1.0") -> None:
        self._name = name
        self._version = version
        self._server = Server(name)
        self._tool_map: dict[str, Any] = {}
        self._register_handlers()

    def _register_handlers(self) -> None:
        tool_pairs = get_tools(self)
        for tool_def, method in tool_pairs:
            self._tool_map[tool_def.name] = (tool_def, method)

        server = self._server

        @server.list_tools()
        async def list_tools() -> list[mcp_types.Tool]:
            return [
                mcp_types.Tool(
                    name=td.name,
                    description=td.description,
                    inputSchema=td.input_schema,
                )
                for td, _ in self._tool_map.values()
            ]

        @server.call_tool()
        async def call_tool(name: str, arguments: dict) -> list[mcp_types.TextContent]:
            if name not in self._tool_map:
                return [mcp_types.TextContent(type="text", text=f"Unknown tool: {name}")]
            _, method = self._tool_map[name]
            try:
                result = method(**arguments)
                if asyncio.iscoroutine(result):
                    result = await result
                if isinstance(result, ToolResult):
                    return [mcp_types.TextContent(type="text", text=result.content)]
                return [mcp_types.TextContent(type="text", text=str(result))]
            except Exception as exc:
                return [mcp_types.TextContent(type="text", text=f"Error: {exc}")]

    async def run_stdio(self) -> None:
        async with stdio_server() as (read_stream, write_stream):
            await self._server.run(
                read_stream,
                write_stream,
                self._server.create_initialization_options(),
            )

    def run(self) -> None:
        asyncio.run(self.run_stdio())
