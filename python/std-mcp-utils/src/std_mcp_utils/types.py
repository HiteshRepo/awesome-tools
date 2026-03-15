from dataclasses import dataclass, field


@dataclass
class ToolDefinition:
    name: str
    description: str
    input_schema: dict = field(default_factory=dict)


@dataclass
class ToolResult:
    content: str
    is_error: bool = False
