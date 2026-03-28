"""ElasticAgentConfig dataclass."""

import os
from dataclasses import dataclass, field


@dataclass
class ElasticAgentConfig:
    es_url: str
    es_api_key: str
    model: str = "claude-sonnet-4-6"
    max_tokens: int = 4096
    max_iterations: int = 10
    anthropic_api_key: str = field(default_factory=lambda: os.environ["ANTHROPIC_API_KEY"])
    docker_image: str = "docker.elastic.co/mcp/elasticsearch"
