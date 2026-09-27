#!/usr/bin/env python3
"""
Database Backup and Recovery Automation (P3-33)
Automates pg_dump backups with scheduled execution and validation.

Usage:
    python scripts/backup_db.py --backup
    python scripts/backup_db.py --restore --backup-file backup_20260406.sql
    python scripts/backup_db.py --list
    python scripts/backup_db.py --cleanup --keep-days 30

Cron schedule (daily at 2am):
    0 2 * * * python /path/to/scripts/backup_db.py --backup
"""

import argparse
import datetime
import gzip
import logging
import os
import subprocess
import sys

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Configuration
DB_URL = os.environ.get("DATABASE_URL", "")
DB_NAME = os.environ.get("DB_NAME", "visionary")
BACKUP_DIR = os.environ.get("BACKUP_DIR", "./backups")
BACKUP_PREFIX = "backup"
KEEP_DAYS = 30


def ensure_backup_dir():
    """Create backup directory if it doesn't exist."""
    os.makedirs(BACKUP_DIR, exist_ok=True)


def get_backup_filename() -> str:
    """Generate a timestamped backup filename."""
    timestamp = datetime.datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    return f"{BACKUP_PREFIX}_{timestamp}.sql.gz"


def run_backup() -> str:
    """Create a compressed database backup."""
    ensure_backup_dir()
    filename = get_backup_filename()
    filepath = os.path.join(BACKUP_DIR, filename)

    logger.info(f"Starting backup to {filepath}")

    try:
        # Use pg_dump with compression
        cmd = f"pg_dump {DB_URL}"
        with open(filepath.replace('.gz', ''), 'w') as f:
            result = subprocess.run(
                cmd.split(),
                stdout=f,
                stderr=subprocess.PIPE,
                text=True,
                timeout=3600  # 1 hour timeout
            )

        if result.returncode != 0:
            logger.error(f"pg_dump failed: {result.stderr}")
            return ""

        # Compress
        with open(filepath.replace('.gz', ''), 'rb') as f_in:
            with gzip.open(filepath, 'wb') as f_out:
                f_out.writelines(f_in)

        # Remove uncompressed file
        os.remove(filepath.replace('.gz', ''))

        file_size = os.path.getsize(filepath)
        logger.info(f"Backup completed: {filepath} ({file_size / 1024 / 1024:.1f} MB)")
        return filepath

    except subprocess.TimeoutExpired:
        logger.error("Backup timed out")
        return ""
    except Exception as e:
        logger.error(f"Backup failed: {e}")
        return ""


def restore_backup(backup_file: str) -> bool:
    """Restore database from a backup file."""
    if not os.path.exists(backup_file):
        logger.error(f"Backup file not found: {backup_file}")
        return False

    logger.info(f"Restoring from {backup_file}")

    try:
        # Decompress if needed
        if backup_file.endswith('.gz'):
            sql_file = backup_file.replace('.gz', '')
            with gzip.open(backup_file, 'rb') as f_in:
                with open(sql_file, 'wb') as f_out:
                    f_out.write(f_in.read())
        else:
            sql_file = backup_file

        # Restore using psql
        cmd = f"psql {DB_URL} -f {sql_file}"
        result = subprocess.run(
            cmd.split(),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            timeout=7200  # 2 hour timeout
        )

        # Cleanup decompressed file
        if backup_file.endswith('.gz') and os.path.exists(sql_file):
            os.remove(sql_file)

        if result.returncode != 0:
            logger.error(f"Restore failed: {result.stderr}")
            return False

        logger.info("Restore completed successfully")
        return True

    except subprocess.TimeoutExpired:
        logger.error("Restore timed out")
        return False
    except Exception as e:
        logger.error(f"Restore failed: {e}")
        return False


def list_backups() -> list:
    """List available backups sorted by date."""
    ensure_backup_dir()
    files = []

    for filename in os.listdir(BACKUP_DIR):
        if filename.startswith(BACKUP_PREFIX) and (filename.endswith('.sql') or filename.endswith('.sql.gz')):
            filepath = os.path.join(BACKUP_DIR, filename)
            stat = os.stat(filepath)
            files.append({
                'filename': filename,
                'path': filepath,
                'size_mb': stat.st_size / 1024 / 1024,
                'created': datetime.datetime.fromtimestamp(stat.st_mtime),
            })

    # Sort by date descending
    files.sort(key=lambda x: x['created'], reverse=True)
    return files


def cleanup_old_backups(keep_days: int = KEEP_DAYS):
    """Remove backups older than keep_days."""
    ensure_backup_dir()
    cutoff = datetime.datetime.utcnow() - datetime.timedelta(days=keep_days)

    removed = 0
    for filename in os.listdir(BACKUP_DIR):
        if not filename.startswith(BACKUP_PREFIX):
            continue

        filepath = os.path.join(BACKUP_DIR, filename)
        stat = os.stat(filepath)
        created = datetime.datetime.fromtimestamp(stat.st_mtime)

        if created < cutoff:
            os.remove(filepath)
            removed += 1
            logger.info(f"Removed old backup: {filename}")

    logger.info(f"Cleanup complete: {removed} old backups removed")


def main():
    parser = argparse.ArgumentParser(description="Database Backup and Recovery")
    parser.add_argument("--backup", action="store_true", help="Create a new backup")
    parser.add_argument("--restore", action="store_true", help="Restore from backup")
    parser.add_argument("--backup-file", help="Path to backup file for restore")
    parser.add_argument("--list", action="store_true", help="List available backups")
    parser.add_argument("--cleanup", action="store_true", help="Remove old backups")
    parser.add_argument("--keep-days", type=int, default=KEEP_DAYS, help="Days to keep backups")
    args = parser.parse_args()

    if not DB_URL:
        logger.error("DATABASE_URL environment variable not set")
        sys.exit(1)

    if args.backup:
        filepath = run_backup()
        if not filepath:
            sys.exit(1)

    elif args.restore:
        if not args.backup_file:
            logger.error("--backup-file required for restore")
            sys.exit(1)
        if not restore_backup(args.backup_file):
            sys.exit(1)

    elif args.list:
        backups = list_backups()
        if not backups:
            print("No backups found")
        else:
            print(f"{'Filename':<40} {'Size (MB)':<12} {'Created':<20}")
            print("-" * 72)
            for b in backups:
                print(f"{b['filename']:<40} {b['size_mb']:<12.1f} {b['created'].strftime('%Y-%m-%d %H:%M')}")

    elif args.cleanup:
        cleanup_old_backups(args.keep_days)

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
