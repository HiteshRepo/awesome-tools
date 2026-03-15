from dataclasses import dataclass, field
from uuid import uuid4


@dataclass
class Document:
    content: str
    metadata: dict = field(default_factory=dict)
    id: str = field(default_factory=lambda: str(uuid4()))


@dataclass
class SearchResult:
    document: Document
    score: float
