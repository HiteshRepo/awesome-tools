from dataclasses import dataclass, field


@dataclass
class Chunk:
    content: str
    document_id: str
    chunk_index: int
    metadata: dict = field(default_factory=dict)


class TextChunker:
    """Splits text into overlapping word-boundary chunks."""

    def __init__(self, chunk_size: int = 512, chunk_overlap: int = 50) -> None:
        if chunk_overlap >= chunk_size:
            raise ValueError("chunk_overlap must be less than chunk_size")
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def chunk(self, text: str, document_id: str, metadata: dict | None = None) -> list[Chunk]:
        words = text.split()
        if not words:
            return []

        metadata = metadata or {}
        chunks: list[Chunk] = []
        step = self.chunk_size - self.chunk_overlap
        start = 0
        idx = 0

        while start < len(words):
            end = min(start + self.chunk_size, len(words))
            content = " ".join(words[start:end])
            chunks.append(Chunk(
                content=content,
                document_id=document_id,
                chunk_index=idx,
                metadata=dict(metadata),
            ))
            if end == len(words):
                break
            start += step
            idx += 1

        return chunks
