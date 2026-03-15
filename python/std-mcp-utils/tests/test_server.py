"""Unit tests for std-mcp-utils."""
import pytest

from std_mcp_utils import BaseMCPServer, ToolDefinition, ToolResult, mcp_tool, get_tools


class SampleServer(BaseMCPServer):
    def __init__(self):
        super().__init__(name="test-server", version="0.1.0")

    @mcp_tool(
        name="add",
        description="Add two numbers",
        input_schema={
            "type": "object",
            "properties": {"a": {"type": "number"}, "b": {"type": "number"}},
            "required": ["a", "b"],
        },
    )
    def add(self, a: float, b: float) -> ToolResult:
        return ToolResult(content=str(a + b))

    @mcp_tool(name="ping", description="Ping the server")
    def ping(self) -> ToolResult:
        return ToolResult(content="pong")


def test_tool_definition():
    td = ToolDefinition(name="x", description="desc")
    assert td.name == "x"
    assert td.input_schema == {}


def test_tool_result():
    tr = ToolResult(content="ok")
    assert not tr.is_error


def test_mcp_tool_decorator():
    server = SampleServer()
    tools = get_tools(server)
    names = {td.name for td, _ in tools}
    assert "add" in names
    assert "ping" in names


def test_add_tool_executes():
    server = SampleServer()
    tools = {td.name: method for td, method in get_tools(server)}
    result = tools["add"](a=3, b=4)
    assert result.content == "7"


def test_ping_tool_executes():
    server = SampleServer()
    tools = {td.name: method for td, method in get_tools(server)}
    result = tools["ping"]()
    assert result.content == "pong"


def test_tool_map_populated():
    server = SampleServer()
    assert "add" in server._tool_map
    assert "ping" in server._tool_map
