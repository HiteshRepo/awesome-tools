from dataclasses import dataclass, field


@dataclass
class LLMConfig:
    provider: str  # "anthropic" | "openai"
    model: str
    api_key: str
    max_tokens: int = 4096
    temperature: float = 0.0
    extra: dict = field(default_factory=dict)
