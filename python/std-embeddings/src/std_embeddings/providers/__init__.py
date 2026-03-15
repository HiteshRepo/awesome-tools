from ..base import EmbeddingConfig, EmbeddingProvider


def create_embedding_provider(config: EmbeddingConfig) -> EmbeddingProvider:
    if config.provider == "openai":
        from .openai import OpenAIEmbeddings
        return OpenAIEmbeddings(config)
    elif config.provider == "voyage":
        from .voyage import VoyageEmbeddings
        return VoyageEmbeddings(config)
    elif config.provider == "local":
        from .sentence_transformers import LocalEmbeddings
        return LocalEmbeddings(config)
    else:
        raise ValueError(f"Unknown provider: {config.provider!r}")
