"""Parent-child chunking with the grid-search-optimal 400/150 configuration.

Parents are markdown sections (## headings); children are overlapping
character chunks split on sentence boundaries where possible.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field


@dataclass
class ParentChunk:
    parent_id: str
    title: str
    content: str
    metadata: dict = field(default_factory=dict)


@dataclass
class ChildChunk:
    chunk_id: str
    parent_id: str
    content: str
    metadata: dict = field(default_factory=dict)


_SENTENCE_END = re.compile(r"(?<=[.!?])\s+")


def split_text(text: str, chunk_size: int = 400, overlap: int = 150) -> list[str]:
    """Sentence-aware sliding-window splitter (chunk_size chars, overlap chars)."""
    text = re.sub(r"\s+", " ", text).strip()
    if len(text) <= chunk_size:
        return [text] if text else []

    sentences = _SENTENCE_END.split(text)
    chunks: list[str] = []
    current = ""
    for sent in sentences:
        if not sent:
            continue
        if current and len(current) + 1 + len(sent) > chunk_size:
            chunks.append(current.strip())
            # Overlap: keep the tail of the previous chunk
            tail = current[-overlap:] if overlap > 0 else ""
            # Start overlap at a word boundary
            sp = tail.find(" ")
            current = (tail[sp + 1 :] + " " if sp != -1 else "") + sent
        else:
            current = f"{current} {sent}".strip() if current else sent
        # Hard-split pathological sentences longer than chunk_size
        while len(current) > chunk_size * 2:
            chunks.append(current[:chunk_size].strip())
            current = current[chunk_size - overlap :]
    if current.strip():
        chunks.append(current.strip())
    return chunks


def parse_markdown_document(
    doc_id: str, text: str, base_metadata: dict | None = None,
    chunk_size: int = 400, overlap: int = 150,
) -> tuple[list[ParentChunk], list[ChildChunk]]:
    """Split a markdown document into section parents and 400/150 children.

    Front-matter style `Key: value` lines at the top become metadata.
    `#` heading = chapter, `##` heading = section.
    """
    metadata = dict(base_metadata or {})
    lines = text.splitlines()

    # Parse simple front-matter (Key: value) before the first heading
    body_start = 0
    for i, line in enumerate(lines):
        s = line.strip()
        if s.startswith("#"):
            body_start = i
            break
        m = re.match(r"^([A-Za-z_][A-Za-z0-9_ ]*):\s*(.+)$", s)
        if m:
            metadata[m.group(1).strip().lower().replace(" ", "_")] = m.group(2).strip()

    chapter = metadata.get("chapter", doc_id)
    parents: list[ParentChunk] = []
    children: list[ChildChunk] = []

    section_title = "Introduction"
    section_buf: list[str] = []
    section_idx = 0

    def flush() -> None:
        nonlocal section_idx
        content = "\n".join(section_buf).strip()
        if not content:
            return
        section_idx += 1
        pid = f"{doc_id}::s{section_idx}"
        pmeta = dict(metadata)
        pmeta.update({"chapter": chapter, "section": section_title})
        parents.append(ParentChunk(pid, section_title, content, pmeta))
        for j, piece in enumerate(split_text(content, chunk_size, overlap), start=1):
            children.append(
                ChildChunk(f"{pid}::c{j}", pid, piece, dict(pmeta))
            )

    for line in lines[body_start:]:
        s = line.strip()
        if s.startswith("# ") and not s.startswith("## "):
            chapter = s[2:].strip()
            continue
        if s.startswith("## "):
            flush()
            section_title = s[3:].strip()
            section_buf = []
            continue
        section_buf.append(line)
    flush()

    return parents, children
