"""
Incremental ingestion module for RAG pipeline.
Implements change data capture for efficient document updates.

Features:
- Document-level change detection (SHA-256)
- Page-level change detection
- Chunk versioning
- Orphaned chunk cleanup
- Rollback support
"""

import hashlib
import json
from typing import List, Dict, Optional, Tuple, Set
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
import asyncio
import asyncpg


@dataclass
class DocumentVersion:
    """Represents a document version with change tracking."""

    taxonomy_id: int
    content_hash: str
    version: int
    ingested_at: datetime
    page_hashes: Dict[int, str]  # page_number -> hash
    metadata: Dict = field(default_factory=dict)


@dataclass
class ChangeDetection:
    """Result of change detection analysis."""

    has_changes: bool
    is_new_document: bool
    changed_pages: List[int]
    deleted_pages: List[int]
    unchanged_pages: List[int]
    old_version: Optional[int]
    new_version: int


class DocumentTracker:
    """
    Tracks document versions and detects changes.

    Usage:
        tracker = DocumentTracker(dsn)
        await tracker.connect()
        changes = await tracker.detect_changes(pdf_path, taxonomy_id)
        if changes.has_changes:
            await ingest_incremental(changes)
    """

    def __init__(self, dsn: str):
        """
        Initialize document tracker.

        Args:
            dsn: Database connection string
        """
        self.dsn = dsn
        self.pool: Optional[asyncpg.Pool] = None

    async def connect(self):
        """Create database connection pool."""
        self.pool = await asyncpg.create_pool(
            self.dsn,
            min_size=2,
            max_size=10,
            command_timeout=60,
        )

    async def close(self):
        """Close connection pool."""
        if self.pool:
            await self.pool.close()

    async def get_current_version(self, taxonomy_id: int) -> Optional[DocumentVersion]:
        """Get current document version for a taxonomy ID."""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                SELECT
                    taxonomy_id,
                    content_hash,
                    version,
                    ingested_at,
                    page_hashes,
                    metadata
                FROM document_versions
                WHERE taxonomy_id = $1
                ORDER BY version DESC
                LIMIT 1
                """,
                taxonomy_id,
            )

            if not row:
                return None

            return DocumentVersion(
                taxonomy_id=row["taxonomy_id"],
                content_hash=row["content_hash"],
                version=row["version"],
                ingested_at=row["ingested_at"],
                page_hashes=row["page_hashes"],
                metadata=row["metadata"] or {},
            )

    async def detect_changes(
        self, file_path: str, taxonomy_id: int
    ) -> ChangeDetection:
        """
        Detect changes between file and database version.

        Args:
            file_path: Path to document file
            taxonomy_id: Taxonomy ID to check

        Returns:
            ChangeDetection result
        """
        # Calculate current file hashes
        doc_hash = self._calculate_file_hash(file_path)
        page_hashes = self._calculate_page_hashes(file_path)

        # Get stored version
        current_version = await self.get_current_version(taxonomy_id)

        if current_version is None:
            # New document
            return ChangeDetection(
                has_changes=True,
                is_new_document=True,
                changed_pages=list(page_hashes.keys()),
                deleted_pages=[],
                unchanged_pages=[],
                old_version=None,
                new_version=1,
            )

        # Compare document hash (quick check)
        if doc_hash == current_version.content_hash:
            # No changes
            return ChangeDetection(
                has_changes=False,
                is_new_document=False,
                changed_pages=[],
                deleted_pages=[],
                unchanged_pages=list(page_hashes.keys()),
                old_version=current_version.version,
                new_version=current_version.version,
            )

        # Document changed - compare page hashes
        old_pages = current_version.page_hashes
        new_pages = page_hashes

        changed_pages = []
        unchanged_pages = []

        for page_num, new_hash in new_pages.items():
            old_hash = old_pages.get(page_num)
            if old_hash != new_hash:
                changed_pages.append(page_num)
            else:
                unchanged_pages.append(page_num)

        deleted_pages = [
            page_num for page_num in old_pages.keys() if page_num not in new_pages
        ]

        return ChangeDetection(
            has_changes=True,
            is_new_document=False,
            changed_pages=changed_pages,
            deleted_pages=deleted_pages,
            unchanged_pages=unchanged_pages,
            old_version=current_version.version,
            new_version=current_version.version + 1,
        )

    async def update_version(
        self,
        taxonomy_id: int,
        content_hash: str,
        page_hashes: Dict[int, str],
        metadata: Dict = None,
    ) -> DocumentVersion:
        """
        Update or create document version.

        Args:
            taxonomy_id: Taxonomy ID
            content_hash: Document-level hash
            page_hashes: Page-level hashes
            metadata: Optional metadata

        Returns:
            New DocumentVersion
        """
        async with self.pool.acquire() as conn:
            # Get current version
            current = await self.get_current_version(taxonomy_id)
            new_version = 1 if current is None else current.version + 1

            # Upsert version
            await conn.execute(
                """
                INSERT INTO document_versions (
                    taxonomy_id,
                    content_hash,
                    version,
                    ingested_at,
                    page_hashes,
                    metadata
                ) VALUES ($1, $2, $3, NOW(), $4, $5)
                ON CONFLICT (taxonomy_id) DO UPDATE SET
                    content_hash = EXCLUDED.content_hash,
                    version = EXCLUDED.version,
                    ingested_at = EXCLUDED.ingested_at,
                    page_hashes = EXCLUDED.page_hashes,
                    metadata = EXCLUDED.metadata
                """,
                taxonomy_id,
                content_hash,
                new_version,
                json.dumps(page_hashes),
                json.dumps(metadata or {}),
            )

            return DocumentVersion(
                taxonomy_id=taxonomy_id,
                content_hash=content_hash,
                version=new_version,
                ingested_at=datetime.now(),
                page_hashes=page_hashes,
                metadata=metadata or {},
            )

    async def mark_chunks_superseded(
        self, taxonomy_id: int, page_numbers: List[int], new_version: int
    ):
        """
        Mark chunks from old pages as superseded.

        Args:
            taxonomy_id: Taxonomy ID
            page_numbers: Page numbers to mark
            new_version: New version number
        """
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                # Mark parent chunks as superseded
                await conn.execute(
                    """
                    UPDATE parent_chunks
                    SET
                        is_current = FALSE,
                        superseded_version = $1,
                        updated_at = NOW()
                    WHERE taxonomy_id = $2
                    AND page_number = ANY($3)
                    AND is_current = TRUE
                    """,
                    new_version,
                    taxonomy_id,
                    page_numbers,
                )

                # Mark child chunks as superseded
                await conn.execute(
                    """
                    UPDATE child_chunks
                    SET
                        is_current = FALSE,
                        superseded_version = $1,
                        updated_at = NOW()
                    WHERE taxonomy_id = $2
                    AND page_number = ANY($3)
                    AND is_current = TRUE
                    """,
                    new_version,
                    taxonomy_id,
                    page_numbers,
                )

    async def cleanup_orphaned_chunks(
        self, taxonomy_id: int, deleted_pages: List[int]
    ) -> int:
        """
        Delete chunks from deleted pages.

        Args:
            taxonomy_id: Taxonomy ID
            deleted_pages: Page numbers that were deleted

        Returns:
            Number of chunks deleted
        """
        if not deleted_pages:
            return 0

        async with self.pool.acquire() as conn:
            # Delete child chunks
            child_result = await conn.execute(
                """
                DELETE FROM child_chunks
                WHERE taxonomy_id = $1
                AND page_number = ANY($2)
                AND is_current = FALSE
                """,
                taxonomy_id,
                deleted_pages,
            )

            # Delete parent chunks
            parent_result = await conn.execute(
                """
                DELETE FROM parent_chunks
                WHERE taxonomy_id = $1
                AND page_number = ANY($2)
                AND is_current = FALSE
                """,
                taxonomy_id,
                deleted_pages,
            )

            # Parse row counts from command tag
            child_count = int(child_result.split()[-1]) if child_result else 0
            parent_count = int(parent_result.split()[-1]) if parent_result else 0

            return child_count + parent_count

    async def rollback_to_version(self, taxonomy_id: int, target_version: int) -> bool:
        """
        Rollback to a previous version.

        Args:
            taxonomy_id: Taxonomy ID
            target_version: Version to rollback to

        Returns:
            True if successful
        """
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                # Get target version
                target = await conn.fetchrow(
                    """
                    SELECT * FROM document_versions
                    WHERE taxonomy_id = $1 AND version = $2
                    """,
                    taxonomy_id,
                    target_version,
                )

                if not target:
                    raise ValueError(f"Version {target_version} not found")

                # Mark current chunks as not current
                await conn.execute(
                    """
                    UPDATE parent_chunks
                    SET is_current = FALSE
                    WHERE taxonomy_id = $1 AND is_current = TRUE
                    """,
                    taxonomy_id,
                )

                await conn.execute(
                    """
                    UPDATE child_chunks
                    SET is_current = FALSE
                    WHERE taxonomy_id = $1 AND is_current = TRUE
                    """,
                    taxonomy_id,
                )

                # Restore target version chunks
                await conn.execute(
                    """
                    UPDATE parent_chunks
                    SET is_current = TRUE
                    WHERE taxonomy_id = $1
                    AND version_ingested = (
                        SELECT ingested_at FROM document_versions
                        WHERE taxonomy_id = $1 AND version = $2
                    )
                    """,
                    taxonomy_id,
                    target_version,
                )

                await conn.execute(
                    """
                    UPDATE child_chunks
                    SET is_current = TRUE
                    WHERE taxonomy_id = $1
                    AND version_ingested = (
                        SELECT ingested_at FROM document_versions
                        WHERE taxonomy_id = $1 AND version = $2
                    )
                    """,
                    taxonomy_id,
                    target_version,
                )

                return True

    def _calculate_file_hash(self, file_path: str) -> str:
        """Calculate SHA-256 hash of entire file."""
        hasher = hashlib.sha256()
        with open(file_path, "rb") as f:
            buf = f.read(65536)
            while len(buf) > 0:
                hasher.update(buf)
                buf = f.read(65536)
        return hasher.hexdigest()

    def _calculate_page_hashes(self, file_path: str) -> Dict[int, str]:
        """
        Calculate per-page hashes.

        For PDF files, uses PyMuPDF.
        For other formats, approximates pages.
        """
        from .parser import ParserFactory

        try:
            parser = ParserFactory.get_parser(file_path)
            elements = parser.parse(file_path)

            # Group by page and hash
            page_hashes = {}
            page_content = {}

            for elem in elements:
                if elem.page_number not in page_content:
                    page_content[elem.page_number] = []
                page_content[elem.page_number].append(elem.text)

            for page_num, content in page_content.items():
                combined = "\n".join(content)
                page_hashes[page_num] = hashlib.sha256(
                    combined.encode("utf-8")
                ).hexdigest()

            return page_hashes

        except Exception as e:
            # Fallback: treat entire file as one page
            with open(file_path, "rb") as f:
                content = f.read()
            return {1: hashlib.sha256(content).hexdigest()}


class IncrementalIngester:
    """
    High-level incremental ingestion orchestrator.

    Usage:
        ingester = IncrementalIngester(dsn)
        await ingester.ingest_file(pdf_path, grade=7, subject="Science")
    """

    def __init__(self, dsn: str):
        """
        Initialize incremental ingester.

        Args:
            dsn: Database connection string
        """
        self.dsn = dsn
        self.tracker = DocumentTracker(dsn)

    async def connect(self):
        """Initialize connections."""
        await self.tracker.connect()

    async def close(self):
        """Close connections."""
        await self.tracker.close()

    async def ingest_file(
        self,
        file_path: str,
        grade: int = 7,
        subject: str = "Science",
        taxonomy_id: int = 0,
        force_full: bool = False,
    ) -> Dict:
        """
        Ingest file with incremental support.

        Args:
            file_path: Path to file
            grade: Grade level
            subject: Subject name
            taxonomy_id: Taxonomy ID
            force_full: Force full re-ingestion

        Returns:
            Ingestion statistics
        """
        stats = {
            "file": file_path,
            "taxonomy_id": taxonomy_id,
            "grade": grade,
            "subject": subject,
            "is_incremental": False,
            "changed_pages": [],
            "deleted_pages": [],
            "unchanged_pages": [],
            "parents_inserted": 0,
            "children_inserted": 0,
            "chunks_deleted": 0,
        }

        # Detect changes
        changes = await self.tracker.detect_changes(file_path, taxonomy_id)

        if not changes.has_changes and not force_full:
            stats["message"] = "No changes detected"
            return stats

        stats["is_incremental"] = not changes.is_new_document
        stats["changed_pages"] = changes.changed_pages
        stats["deleted_pages"] = changes.deleted_pages
        stats["unchanged_pages"] = changes.unchanged_pages

        # Parse document
        from .parser import ParserFactory

        parser = ParserFactory.get_parser(file_path)
        elements = parser.parse(file_path, grade=grade, subject=subject, taxonomy_id=taxonomy_id)

        # Filter elements based on changes
        if not force_full and not changes.is_new_document:
            # Only process changed pages
            elements = [
                elem
                for elem in elements
                if elem.page_number in changes.changed_pages
                or elem.page_number in changes.unchanged_pages
            ]

        # Chunk elements
        from .chunker import create_parent_child_chunks

        parents, children = create_parent_child_chunks(elements)

        # Apply deduplication
        from .dedup import DeduplicationDetector

        detector = DeduplicationDetector()
        filtered_parents = []
        filtered_children = []

        for parent in parents:
            chunk = Chunk(content=parent.content, chunk_id=parent.parent_id, taxonomy_id=parent.taxonomy_id)
            is_dup, _ = detector.is_duplicate(chunk)
            if not is_dup:
                filtered_parents.append(parent)
                detector.add_to_index(chunk)

        for child in children:
            chunk = Chunk(content=child.content, chunk_id=child.child_id, taxonomy_id=child.taxonomy_id)
            is_dup, _ = detector.is_duplicate(chunk)
            if not is_dup:
                filtered_children.append(child)
                detector.add_to_index(chunk)

        # Mark old chunks as superseded
        if not changes.is_new_document:
            await self.tracker.mark_chunks_superseded(
                taxonomy_id, changes.changed_pages + changes.deleted_pages, changes.new_version
            )

        # Insert new chunks
        from .writer import AlloyDBWriter

        async with self.pool.acquire() as conn:
            writer = AlloyDBWriter(conn)
            success, failure = await writer.insert_batch(filtered_parents, filtered_children)
            stats["parents_inserted"] = success
            stats["children_inserted"] = len(filtered_children) - failure

        # Cleanup orphaned chunks
        deleted_count = await self.tracker.cleanup_orphaned_chunks(
            taxonomy_id, changes.deleted_pages
        )
        stats["chunks_deleted"] = deleted_count

        # Update version tracking
        doc_hash = self.tracker._calculate_file_hash(file_path)
        page_hashes = self.tracker._calculate_page_hashes(file_path)
        await self.tracker.update_version(taxonomy_id, doc_hash, page_hashes)

        return stats


# Import for Chunk type
from .dedup import Chunk
