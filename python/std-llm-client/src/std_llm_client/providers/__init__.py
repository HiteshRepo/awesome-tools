from ..config import LLMConfig
from ..base import LLMClient


def create_llm_client(config: LLMConfig) -> LLMClient:
    if config.provider == "anthropic":
        from .anthropic import AnthropicClient
        return AnthropicClient(config)
    elif config.provider == "openai":
        from .openai import OpenAIClient
        return OpenAIClient(config)
    else:
        raise ValueError(f"Unknown provider: {config.provider!r}")
