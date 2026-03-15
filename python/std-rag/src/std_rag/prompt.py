from std_llm_client import Message

from .chunker import Chunk

_SYSTEM = (
    "You are a helpful assistant. Answer the user's question using ONLY the provided context. "
    "If the context does not contain enough information to answer, say so."
)


def build_rag_prompt(question: str, chunks: list[Chunk]) -> list[Message]:
    context_parts = [f"[{i + 1}] {chunk.content}" for i, chunk in enumerate(chunks)]
    context_block = "\n\n".join(context_parts)
    user_content = f"Context:\n{context_block}\n\nQuestion: {question}"
    return [
        Message(role="system", content=_SYSTEM),
        Message(role="user", content=user_content),
    ]
