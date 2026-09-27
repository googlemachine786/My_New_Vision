"""
Visionary RAG - DLQ Retry Processor
Processes failed ingestion attempts from dead letter queue
"""

import asyncio
import asyncpg
import json
import time
from typing import List, Dict
from datetime import datetime
import structlog

logger = structlog.get_logger()

class DLQRetryProcessor:
    """Processes failed ingestion attempts with retry logic."""
    
    def __init__(self, dsn: str, max_retries: int = 3):
        self.dsn = dsn
        self.max_retries = max_retries
        self.pool = None
        
    async def connect(self):
        """Create connection pool."""
        self.pool = await asyncpg.create_pool(
            self.dsn,
            min_size=2,
            max_size=10,
            command_timeout=60
        )
        logger.info("DLQ processor connected to database")
    
    async def close(self):
        """Close connection pool."""
        if self.pool:
            await self.pool.close()
            logger.info("DLQ processor disconnected")
    
    async def fetch_failed_items(self, limit: int = 100) -> List[Dict]:
        """Fetch failed items from DLQ."""
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(
                """
                SELECT dlq_id, payload, error_message, retry_count, created_at
                FROM ingestion_dlq
                WHERE failed = TRUE AND retry_count < $1
                ORDER BY created_at ASC
                LIMIT $2
                """,
                self.max_retries,
                limit
            )
            
            return [dict(row) for row in rows]
    
    async def retry_insert(self, item: Dict) -> bool:
        """Retry inserting a failed item."""
        try:
            payload = json.loads(item['payload'])
            
            # Extract parent and child data
            parent_data = payload.get('parent', {})
            children_data = payload.get('children', [])
            
            async with self.pool.acquire() as conn:
                async with conn.transaction():
                    # Insert parent
                    parent_id = await conn.fetchval(
                        """
                        INSERT INTO parent_chunks (
                            parent_id, taxonomy_id, content, extracted_keywords,
                            page_number, chapter, section, subsection, content_type
                        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                        RETURNING parent_id
                        """,
                        parent_data.get('parent_id'),
                        parent_data.get('taxonomy_id'),
                        parent_data.get('content'),
                        parent_data.get('extracted_keywords', []),
                        parent_data.get('page_number'),
                        parent_data.get('chapter'),
                        parent_data.get('section'),
                        parent_data.get('subsection'),
                        parent_data.get('content_type')
                    )
                    
                    # Insert children
                    for child in children_data:
                        await conn.execute(
                            """
                            INSERT INTO child_chunks (
                                child_id, parent_id, taxonomy_id, content,
                                embedding, page_number, content_type
                            ) VALUES ($1, $2, $3, $4, $5, $6, $7)
                            """,
                            child.get('child_id'),
                            parent_id,
                            child.get('taxonomy_id'),
                            child.get('content'),
                            child.get('embedding'),
                            child.get('page_number'),
                            child.get('content_type')
                        )
                    
                    # Mark as successful
                    await conn.execute(
                        """
                        UPDATE ingestion_dlq
                        SET failed = FALSE, processed_at = NOW()
                        WHERE dlq_id = $1
                        """,
                        item['dlq_id']
                    )
                    
                    logger.info("DLQ item retry successful", dlq_id=item['dlq_id'])
                    return True
                    
        except Exception as e:
            logger.error("DLQ retry failed", dlq_id=item['dlq_id'], error=str(e))
            
            # Increment retry count
            async with self.pool.acquire() as conn:
                await conn.execute(
                    """
                    UPDATE ingestion_dlq
                    SET retry_count = retry_count + 1, error_message = $2
                    WHERE dlq_id = $1
                    """,
                    item['dlq_id'],
                    str(e)
                )
            
            return False
    
    async def process_dlq(self):
        """Process all failed items in DLQ."""
        logger.info("Starting DLQ processing")
        
        failed_items = await self.fetch_failed_items()
        
        if not failed_items:
            logger.info("No failed items to process")
            return
        
        logger.info(f"Found {len(failed_items)} failed items to retry")
        
        success_count = 0
        failure_count = 0
        
        for item in failed_items:
            success = await self.retry_insert(item)
            if success:
                success_count += 1
            else:
                failure_count += 1
            
            # Small delay between retries
            await asyncio.sleep(0.5)
        
        logger.info(
            "DLQ processing complete",
            total=len(failed_items),
            success=success_count,
            failure=failure_count
        )
    
    async def get_dlq_stats(self) -> Dict:
        """Get DLQ statistics."""
        async with self.pool.acquire() as conn:
            stats = await conn.fetchrow(
                """
                SELECT
                    COUNT(*) FILTER (WHERE failed = TRUE) as failed_count,
                    COUNT(*) FILTER (WHERE failed = FALSE) as success_count,
                    COUNT(*) FILTER (WHERE retry_count >= 3) as max_retries_count,
                    AVG(retry_count) as avg_retries
                FROM ingestion_dlq
                """
            )
            return dict(stats)


async def main():
    """Run DLQ processor."""
    import os
    
    dsn = os.getenv('ALLOYDB_DSN')
    if not dsn:
        logger.error("ALLOYDB_DSN not set")
        return
    
    processor = DLQRetryProcessor(dsn)
    
    try:
        await processor.connect()
        
        # Process DLQ
        await processor.process_dlq()
        
        # Get stats
        stats = await processor.get_dlq_stats()
        logger.info("DLQ Statistics", **stats)
        
    finally:
        await processor.close()


if __name__ == "__main__":
    asyncio.run(main())
