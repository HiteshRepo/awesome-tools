import inspect
from typing import Callable

from .types import ToolDefinition

_MCP_TOOL_ATTR = "_mcp_tool"


def mcp_tool(name: str, description: str, input_schema: dict | None = None):
    """Decorator that marks a method as an MCP tool."""
    def decorator(fn: Callable) -> Callable:
        setattr(fn, _MCP_TOOL_ATTR, ToolDefinition(
            name=name,
            description=description,
            input_schema=input_schema or {},
        ))
        return fn
    return decorator


def get_tools(obj: object) -> list[tuple[ToolDefinition, Callable]]:
    """Return all (ToolDefinition, bound_method) pairs on *obj*."""
    result: list[tuple[ToolDefinition, Callable]] = []
    for _, method in inspect.getmembers(obj, predicate=inspect.ismethod):
        tool_def = getattr(method, _MCP_TOOL_ATTR, None)
        if tool_def is not None:
            result.append((tool_def, method))
    return result
