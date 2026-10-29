"""Core scraping logic — kept separate so it's testable without the server."""
from __future__ import annotations

import httpx
import trafilatura


_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (compatible; agent-forge/0.1; +https://github.com/HiteshRepo/agent-forge)"
    )
}


class ScrapeError(Exception):
    pass


def scrape(url: str, timeout: float = 15.0) -> str:
    """Fetch *url* and return its main readable text content.

    Uses httpx for the HTTP request (consistent with agent-forge stack) and
    trafilatura for boilerplate removal and content extraction.

    Raises ScrapeError with a descriptive message on failure.
    """
    try:
        response = httpx.get(url, headers=_HEADERS, timeout=timeout, follow_redirects=True)
        response.raise_for_status()
    except httpx.HTTPStatusError as exc:
        raise ScrapeError(f"HTTP {exc.response.status_code} fetching {url}") from exc
    except httpx.RequestError as exc:
        raise ScrapeError(f"Network error fetching {url}: {exc}") from exc

    text = trafilatura.extract(
        response.text,
        url=url,
        include_comments=False,
        include_tables=True,
        no_fallback=False,       # use fallback chain on failure
        output_format="markdown",
    )

    if not text:
        raise ScrapeError(f"Could not extract readable content from {url}")

    return text
