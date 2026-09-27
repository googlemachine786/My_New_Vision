"""
AlloyDB Async Writer

Writes parent-child chunks to AlloyDB with proper transaction handling
and dead letter queue (DLQ) routing on failure.

Critical:
- Use psycopg3 (psycopg), NOT psycopg2 (no async support)
- TEXT[] marshalling: Python list → psycopg3 maps to TEXT[]
- VECTOR: use pgvector's register_vector for VECTOR type
- Transaction pattern: parent+child insert atomic (both or neither)
- DLQ: if child insert fails, parent insert is rolled back
"""

import asyncio
from typing import List, Optional, Dict, Any, Tuple
from dataclasses import dataclass, field
from contextlib import asynccontextmanager
import structlog

import psycopg
from psycopg import AsyncConnection, sql
from psycopg.rows import dict_row
from pgvector.psycopg import register_vector

from ..chunker.parent_child import ParentChunk, ChildChunk

logger = structlog.get_logger()


@dataclass
class WriterConfig:
    """AlloyDB writer configuration.
    
    Attributes:
        dsn: Database connection string
        pool_min_size: Minimum pool size (default: 2)
        pool_max_size: Maximum pool size (default: 10)
        statement_timeout: Statement timeout in seconds (default: 60)
    """
    dsn: str
    pool_min_size: int = 2
    pool_max_size: int = 10
    statement_timeout: int = 60
    
    def __post_init__(self):
        """Validate configuration."""
        if self.pool_max_size < self.pool_min_size:
            raise ValueError("pool_max_size must be >= pool_min_size")


class AlloyDBWriter:
    """Async AlloyDB writer with transaction support.
    
    Attributes:
        config: WriterConfig instance
        pool: Async connection pool
    """
    
    def __init__(self, config: WriterConfig):
        """Initialize writer.
        
        Args:
            config: WriterConfig with connection settings
        """
        self.config = config
        self.pool: Optional[psycopg.AsyncPool] = None
        
        logger.info("AlloyDB writer initialized",
                   dsn_host=config.dsn.split("host=")[1].split()[0] if "host=" in config.dsn else "unknown",
                   pool_size=f"{config.pool_min_size}-{config.pool_max_size}")
    
    async def connect(self):
        """Create connection pool."""
        if self.pool is not None:
            logger.warning("Connection pool already exists")
            return
        
        self.pool = psycopg.AsyncPool(
            self.config.dsn,
            min_size=self.config.pool_min_size,
            max_size=self.config.pool_max_size,
            statement_timeout=self.config.statement_timeout * 1000,  # ms
        )
        
        # Register pgvector types
        async with self.pool.connection() as conn:
            await register_vector(conn)
        
        logger.info("Connection pool created",
                   pool_size=self.pool.min_size)
    
    async def close(self):
        """Close connection pool."""
        if self.pool:
            await self.pool.close()
            self.pool = None
            logger.info("Connection pool closed")
    
    @asynccontextmanager
    async def connection(self):
        """Get connection from pool.
        
        Yields:
            AsyncConnection object
        """
        if self.pool is None:
            raise RuntimeError("Not connected. Call connect() first.")
        
        async with self.pool.connection() as conn:
            await register_vector(conn)
            yield conn
    
    async def insert_parent_child(
        self,
        parent: ParentChunk,
        children: List[ChildChunk],
    ) -> Tuple[bool, Optional[str]]:
        """Insert parent and child chunks atomically.
        
        Transaction pattern:
        1. Begin transaction
        2. Insert parent
        3. Insert all children
        4. Commit
        
        On failure:
        1. Rollback parent insert
        2. Insert to DLQ
        3. Re-raise exception
        
        Args:
            parent: ParentChunk to insert
            children: List of ChildChunk objects
        
        Returns:
            Tuple of (success, error_message)
        """
        try:
            async with self.connection() as conn:
                async with conn.transaction():
                    # Insert parent
                    await self._insert_parent(conn, parent)
                    
                    # Insert all children
                    for child in children:
                        await self._insert_child(conn, child)
                    
                    logger.info("Parent-child inserted",
                               parent_id=parent.parent_id,
                               children=len(children))
                    
                    return True, None
                    
        except Exception as e:
            # Insert to DLQ
            error_msg = str(e)
            await self._insert_dlq(
                payload=parent.__dict__,
                error_message=error_msg,
                error_type=type(e).__name__,
            )
            
            logger.error("Parent-child insert failed, rolled back to DLQ",
                        parent_id=parent.parent_id,
                        error=error_msg,
                        error_type=type(e).__name__)
            
            return False, error_msg
    
    async def _insert_parent(
        self,
        conn: AsyncConnection,
        parent: ParentChunk,
    ):
        """Insert parent chunk.
        
        Args:
            conn: Database connection
            parent: ParentChunk to insert
        """
        query = """
            INSERT INTO parent_chunks (
                parent_id,
                taxonomy_id,
                content,
                extracted_keywords,
                page_number,
                chapter,
                section,
                subsection,
                content_type
            ) VALUES (
                %s, %s, %s, %s, %s, %s, %s, %s, %s
            )
        """
        
        await conn.execute(
            query,
            (
                parent.parent_id,
                parent.taxonomy_id,
                parent.content,
                parent.extracted_keywords,  # Python list → TEXT[]
                parent.page_number,
                parent.chapter,
                parent.section,
                parent.subsection,
                parent.content_type,
            ),
        )
    
    async def _insert_child(
        self,
        conn: AsyncConnection,
        child: ChildChunk,
    ):
        """Insert child chunk.
        
        Args:
            conn: Database connection
            child: ChildChunk to insert
        """
        query = """
            INSERT INTO child_chunks (
                child_id,
                parent_id,
                taxonomy_id,
                content,
                embedding,
                page_number,
                content_type
            ) VALUES (
                %s, %s, %s, %s, %s, %s, %s
            )
        """
        
        # Get embedding from metadata (set by pipeline before calling)
        embedding = child.metadata.get("embedding")
        
        if embedding is None:
            raise ValueError(f"Child chunk {child.child_id} missing embedding")
        
        await conn.execute(
            query,
            (
                child.child_id,
                child.parent_id,
                child.taxonomy_id,
                child.content,
                embedding,  # pgvector handles List[float] → VECTOR
                child.page_number,
                child.content_type,
            ),
        )
    
    async def _insert_dlq(
        self,
        payload: Dict[str, Any],
        error_message: str,
        error_type: str,
        retry_count: int = 0,
    ):
        """Insert failed record to dead letter queue.
        
        Args:
            payload: Original data that failed
            error_message: Error message
            error_type: Exception type name
            retry_count: Number of retry attempts
        """
        async with self.connection() as conn:
            query = """
                INSERT INTO ingestion_dlq (
                    payload,
                    error_message,
                    error_type,
                    retry_count
                ) VALUES (%s, %s, %s, %s)
            """
            
            await conn.execute(
                query,
                (
                    psycopg.types.json.Jsonb(payload),
                    error_message,
                    error_type,
                    retry_count,
                ),
            )
    
    async def insert_feedback(
        self,
        session_id: str,
        user_query: str,
        retrieved_context: List[str],  # UUIDs
        llm_response: str,
        feedback_score: Optional[int] = None,
    ) -> str:
        """Insert feedback loop record.
        
        Args:
            session_id: Session identifier
            user_query: User's query
            retrieved_context: List of chunk UUIDs
            llm_response: LLM's response
            feedback_score: Optional score (-1, 0, 1)
        
        Returns:
            Feedback ID (UUID)
        """
        import uuid
        
        feedback_id = str(uuid.uuid4())
        
        async with self.connection() as conn:
            query = """
                INSERT INTO ai_feedback_loop (
                    feedback_id,
                    session_id,
                    user_query,
                    retrieved_context,
                    llm_response,
                    feedback_score
                ) VALUES (%s, %s, %s, %s, %s, %s)
                RETURNING feedback_id
            """
            
            result = await conn.execute(
                query,
                (
                    feedback_id,
                    session_id,
                    user_query,
                    retrieved_context,  # Python list → UUID[]
                    llm_response,
                    feedback_score,
                ),
            )
            
            logger.info("Feedback inserted",
                       feedback_id=feedback_id,
                       session_id=session_id,
                       context_count=len(retrieved_context))
        
        return feedback_id


async def insert_parent_child(
    writer: AlloyDBWriter,
    parent: ParentChunk,
    children: List[ChildChunk],
) -> bool:
    """Convenience function to insert parent-child chunks.
    
    Args:
        writer: AlloyDBWriter instance
        parent: ParentChunk to insert
        children: List of ChildChunk objects
    
    Returns:
        True if successful
    """
    success, error = await writer.insert_parent_child(parent, children)
    return success


async def insert_parent_child_batch(
    writer: AlloyDBWriter,
    parent_child_pairs: List[Tuple[ParentChunk, List[ChildChunk]]],
    concurrency: int = 5,
) -> Dict[str, int]:
    """Insert multiple parent-child pairs concurrently.
    
    Args:
        writer: AlloyDBWriter instance
        parent_child_pairs: List of (ParentChunk, [ChildChunk]) tuples
        concurrency: Maximum concurrent inserts
    
    Returns:
        Dict with success/failure counts
    """
    semaphore = asyncio.Semaphore(concurrency)
    
    async def insert_with_semaphore(parent, children):
        async with semaphore:
            return await writer.insert_parent_child(parent, children)
    
    tasks = [
        insert_with_semaphore(parent, children)
        for parent, children in parent_child_pairs
    ]
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    success_count = sum(1 for r in results if r is True)
    failure_count = sum(1 for r in results if r is False or isinstance(r, Exception))
    
    logger.info("Batch insert complete",
               total=len(parent_child_pairs),
               success=success_count,
               failure=failure_count)
    
    return {
        "total": len(parent_child_pairs),
        "success": success_count,
        "failure": failure_count,
    }


# Example usage:
if __name__ == "__main__":
    import os
    
    async def main():
        dsn = os.getenv("ALLOYDB_DSN")
        if not dsn:
            print("Set ALLOYDB_DSN environment variable to test")
            return
        
        config = WriterConfig(dsn=dsn)
        writer = AlloyDBWriter(config)
        
        await writer.connect()
        
        # Test insert would go here
        # (requires actual ParentChunk/ChildChunk with embeddings)
        
        await writer.close()
    
    asyncio.run(main())
