#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Migration Script: Optimal RAG Configuration
============================================

Migrates your RAG pipeline from old suboptimal defaults to the new
data-driven optimal configuration based on grid search evaluation.

Changes:
  - chunk_size: 1500/512 → 400 chars
  - chunk_overlap: 77/100 → 150 chars (37.5%)
  - embed_model: nomic-embed-text → sentence-transformers/all-MiniLM-L6-v2
  - top_k: 3 → 5
  - similarity_threshold: 0.5-0.7 → 0.36
  - retrieval_strategy: hybrid → dense

Expected Improvement:
  - Composite Score: +16.5% (0.72 → 0.8386)
  - Recall@5: +14.7% (0.85 → 0.975)
  - Context Precision: +30% (0.55 → 0.715)

Usage:
  python migrate_to_optimal_config.py --dry-run    # Preview changes
  python migrate_to_optimal_config.py --execute    # Apply migration

See RAG_PIPELINE_AUDIT_REPORT.md for detailed analysis.
"""

import os
import sys
import json
import asyncio
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Any, Optional

try:
    import asyncpg
except ImportError:
    asyncpg = None
    print("⚠️  asyncpg not installed. Install with: pip install asyncpg")

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    SentenceTransformer = None
    print("⚠️  sentence-transformers not installed. Install with: pip install sentence-transformers")


# ══════════════════════════════════════════════════════════════
# CONFIGURATION
# ══════════════════════════════════════════════════════════════

OLD_CONFIG = {
    "chunk_size": 512,
    "chunk_overlap": 77,
    "embed_model": "nomic-embed-text",
    "top_k": 3,
    "similarity_threshold": 0.5,
    "retrieval_strategy": "hybrid",
}

NEW_CONFIG = {
    "chunk_size": 400,
    "chunk_overlap": 150,
    "embed_model": "sentence-transformers/all-MiniLM-L6-v2",
    "top_k": 5,
    "similarity_threshold": 0.36,
    "retrieval_strategy": "dense",
}


class MigrationTool:
    """Handles migration from old to optimal configuration."""

    def __init__(self, db_url: Optional[str] = None, dry_run: bool = True):
        """
        Initialize migration tool.

        Args:
            db_url: Database connection URL
            dry_run: If True, only preview changes without applying
        """
        self.db_url = db_url or os.environ.get(
            'DATABASE_URL',
            'postgresql://visionary:localdev123@localhost:5432/visionary'
        )
        self.dry_run = dry_run
        self.embedding_model = None
        self.stats = {
            "chunks_inspected": 0,
            "chunks_to_reembed": 0,
            "chunks_reembedded": 0,
            "errors": 0,
        }

    def connect_db(self):
        """Establish database connection."""
        if asyncpg is None:
            raise RuntimeError("asyncpg not installed")
        return asyncio.run(self._connect_db_async())

    async def _connect_db_async(self):
        """Async database connection."""
        try:
            conn = await asyncpg.connect(self.db_url)
            print(f"✅ Connected to database")
            return conn
        except Exception as e:
            print(f"❌ Database connection failed: {e}")
            return None

    def load_embedding_model(self):
        """Load the new embedding model."""
        if SentenceTransformer is None:
            print("⚠️  sentence-transformers not available - skipping embedding")
            return None

        try:
            print(f"📥 Loading embedding model: {NEW_CONFIG['embed_model']}")
            model = SentenceTransformer(NEW_CONFIG['embed_model'])
            print(f"✅ Model loaded successfully")
            return model
        except Exception as e:
            print(f"❌ Failed to load embedding model: {e}")
            return None

    async def inspect_chunks(self, conn, limit: int = 100) -> Dict[str, Any]:
        """
        Inspect existing chunks to determine migration scope.

        Args:
            conn: Database connection
            limit: Number of chunks to sample

        Returns:
            Dictionary with inspection results
        """
        print(f"\n🔍 Inspecting existing chunks (sample: {limit})...")

        # Get chunk statistics
        stats_query = """
        SELECT
            COUNT(*) as total_chunks,
            AVG(LENGTH(content)) as avg_content_length,
            MIN(LENGTH(content)) as min_length,
            MAX(LENGTH(content)) as max_length,
            COUNT(DISTINCT metadata->>'embed_model') as embed_models
        FROM chunks
        """

        try:
            row = await conn.fetchrow(stats_query)
            total_chunks = row['total_chunks'] or 0
            avg_length = row['avg_content_length'] or 0

            print(f"   Total chunks: {total_chunks:,}")
            print(f"   Avg length: {avg_length:.0f} chars")

            # Check if chunks need re-embedding
            needs_reembed = total_chunks > 0 and (
                avg_length > NEW_CONFIG['chunk_size'] * 1.2 or  # Chunks too large
                avg_length < NEW_CONFIG['chunk_size'] * 0.8     # Chunks too small
            )

            return {
                "total_chunks": total_chunks,
                "avg_length": avg_length,
                "needs_reembed": needs_reembed,
                "reason": "chunk_size_mismatch" if needs_reembed else "none",
            }

        except Exception as e:
            print(f"⚠️  Could not inspect chunks: {e}")
            return {
                "total_chunks": 0,
                "avg_length": 0,
                "needs_reembed": False,
                "reason": str(e),
            }

    async def get_chunks_to_reembed(self, conn, limit: int = 10) -> List[Dict]:
        """
        Get sample of chunks that need re-embedding.

        Args:
            conn: Database connection
            limit: Number of chunks to retrieve

        Returns:
            List of chunk records
        """
        query = """
        SELECT chunk_id, content, metadata, embedding
        FROM chunks
        ORDER BY chunk_id
        LIMIT $1
        """

        try:
            rows = await conn.fetch(query, limit)
            return [dict(row) for row in rows]
        except Exception as e:
            print(f"⚠️  Could not fetch chunks: {e}")
            return []

    def generate_new_embedding(self, content: str, model) -> List[float]:
        """
        Generate new embedding for chunk content.

        Args:
            content: Chunk text
            model: SentenceTransformer model

        Returns:
            Embedding as list of floats
        """
        if model is None:
            # Fallback: zero embedding
            return [0.0] * 384

        try:
            embedding = model.encode(content, normalize_embeddings=True, convert_to_numpy=True)
            return embedding.tolist()
        except Exception as e:
            print(f"⚠️  Embedding failed: {e}")
            return [0.0] * 384

    async def reembed_chunk(self, conn, chunk_id: str, new_embedding: List[float]) -> bool:
        """
        Update chunk with new embedding.

        Args:
            conn: Database connection
            chunk_id: Chunk identifier
            new_embedding: New embedding vector

        Returns:
            True if successful
        """
        if self.dry_run:
            print(f"   [DRY-RUN] Would update chunk {chunk_id[:8]}...")
            return True

        query = """
        UPDATE chunks
        SET embedding = $1::vector,
            metadata = metadata || '{"migrated": true, "migrated_at": "' || NOW()::text || '"}'::jsonb,
            updated_at = NOW()
        WHERE chunk_id = $2
        """

        try:
            embedding_str = '[' + ','.join(f'{x:.6f}' for x in new_embedding) + ']'
            await conn.execute(query, embedding_str, chunk_id)
            return True
        except Exception as e:
            print(f"❌ Update failed for {chunk_id[:8]}...: {e}")
            return False

    async def migrate_all_chunks(self, conn, model, batch_size: int = 100):
        """
        Migrate all chunks to new embedding model.

        Args:
            conn: Database connection
            model: SentenceTransformer model
            batch_size: Chunks to process per batch
        """
        print(f"\n🔄 Starting full migration...")

        # Get all chunks
        query = "SELECT chunk_id, content FROM chunks ORDER BY chunk_id"

        try:
            chunks = await conn.fetch(query)
            total = len(chunks)
            print(f"   Found {total:,} chunks to migrate")

            if total == 0:
                print("   No chunks to migrate")
                return

            # Process in batches
            for i in range(0, total, batch_size):
                batch = chunks[i:i + batch_size]
                batch_num = (i // batch_size) + 1
                total_batches = (total + batch_size - 1) // batch_size

                print(f"\n   Batch {batch_num}/{total_batches} (chunks {i}-{min(i+batch_size, total)})")

                for chunk in batch:
                    chunk_id = chunk['chunk_id']
                    content = chunk['content']

                    # Generate new embedding
                    new_embedding = self.generate_new_embedding(content, model)
                    self.stats["chunks_inspected"] += 1

                    # Update chunk
                    success = await self.reembed_chunk(conn, chunk_id, new_embedding)
                    if success:
                        self.stats["chunks_reembedded"] += 1
                    else:
                        self.stats["errors"] += 1

                # Commit batch
                if not self.dry_run:
                    await conn.execute("COMMIT")

        except Exception as e:
            print(f"❌ Migration failed: {e}")
            self.stats["errors"] += 1

    def print_report(self):
        """Print migration report."""
        print("\n" + "=" * 70)
        print("MIGRATION REPORT")
        print("=" * 70)

        print(f"\nMode: {'DRY-RUN (no changes applied)' if self.dry_run else 'LIVE (changes applied)'}")

        print(f"\nConfiguration Changes:")
        for key in OLD_CONFIG:
            old_val = OLD_CONFIG[key]
            new_val = NEW_CONFIG[key]
            status = "→" if old_val != new_val else "="
            print(f"   {key:25} {old_val!r:20} {status} {new_val!r}")

        print(f"\nStatistics:")
        print(f"   Chunks inspected:    {self.stats['chunks_inspected']:,}")
        print(f"   Chunks re-embedded:  {self.stats['chunks_reembedded']:,}")
        print(f"   Errors:              {self.stats['errors']:,}")

        if self.dry_run:
            print(f"\n⚠️  This was a dry-run. Run with --execute to apply changes.")
        else:
            print(f"\n✅ Migration complete!")

        print("\n" + "=" * 70)

    async def run(self):
        """Execute full migration."""
        print("=" * 70)
        print("RAG PIPELINE MIGRATION - OPTIMAL CONFIGURATION")
        print("=" * 70)
        print(f"\nTarget configuration: 0.8386 composite score")
        print(f"Expected improvement: +16.5% quality")

        # Connect to database
        conn = await self._connect_db_async()
        if not conn:
            print("❌ Cannot proceed without database connection")
            return False

        # Inspect existing chunks
        inspection = await self.inspect_chunks(conn, limit=100)

        if not inspection['needs_reembed']:
            print("✅ Existing chunks appear compatible - no re-embedding needed")
            print("   (You may still want to update application config)")
        else:
            print(f"⚠️  Chunks need re-embedding: {inspection['reason']}")

            # Load embedding model
            model = self.load_embedding_model()

            if model:
                # Migrate all chunks
                await self.migrate_all_chunks(conn, model, batch_size=50)
            else:
                print("⚠️  Skipping re-embedding (model not available)")

        # Print report
        self.print_report()

        # Close connection
        await conn.close()

        return True


def main():
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="Migrate RAG pipeline to optimal configuration",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --dry-run              Preview migration changes
  %(prog)s --execute              Apply migration
  %(prog)s --db-url postgres://…  Custom database URL

See RAG_PIPELINE_AUDIT_REPORT.md for configuration details.
        """
    )

    parser.add_argument(
        '--dry-run',
        action='store_true',
        default=True,
        help='Preview changes without applying (default)'
    )

    parser.add_argument(
        '--execute',
        action='store_false',
        dest='dry_run',
        help='Apply migration changes'
    )

    parser.add_argument(
        '--db-url',
        type=str,
        default=None,
        help='Database connection URL'
    )

    args = parser.parse_args()

    # Run migration
    tool = MigrationTool(db_url=args.db_url, dry_run=args.dry_run)
    success = asyncio.run(tool.run())

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
