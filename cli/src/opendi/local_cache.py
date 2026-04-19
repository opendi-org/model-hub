"""Local SQLite cache for pulled models (cache-aside pattern).

Database location: OS-appropriate user data directory
  Windows  : %APPDATA%\\opendi\\models.db
  Linux    : ~/.local/share/opendi/models.db
  macOS    : ~/Library/Application Support/opendi/models.db

Schema (v2):
  pulled_models
    id          INTEGER PRIMARY KEY AUTOINCREMENT
    repo_id     INTEGER                -- hub repo ID; NULL for pre-migration rows
    owner       TEXT NOT NULL          -- display only (rename-resilient via repo_id)
    repo        TEXT NOT NULL          -- display only (slug)
    tag         TEXT NOT NULL
    content     TEXT NOT NULL          -- raw JSON string
    digest      TEXT                   -- ETag from hub (used for If-None-Match on re-pull)
    pulled_at   TEXT NOT NULL          -- ISO-8601 UTC timestamp
    stale       INTEGER NOT NULL DEFAULT 0  -- set to 1 by delete tag/repo; cleared on re-pull
    UNIQUE(owner, repo, tag)           -- primary dedup key; also see partial index below
"""

import logging
import sqlite3
from contextlib import contextmanager
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterator

from platformdirs import user_data_dir

logger = logging.getLogger(__name__)


def _db_path() -> Path:
    path = Path(user_data_dir("opendi", appauthor=False)) / "models.db"
    path.parent.mkdir(parents=True, exist_ok=True)
    return path


def _init_schema(conn: sqlite3.Connection) -> None:
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("""
        CREATE TABLE IF NOT EXISTS pulled_models (
            id        INTEGER PRIMARY KEY AUTOINCREMENT,
            repo_id   INTEGER,
            owner     TEXT NOT NULL,
            repo      TEXT NOT NULL,
            tag       TEXT NOT NULL,
            content   TEXT NOT NULL,
            digest    TEXT,
            pulled_at TEXT NOT NULL,
            stale     INTEGER NOT NULL DEFAULT 0,
            UNIQUE(owner, repo, tag)
        )
    """)
    # Partial unique index so repo_id+tag is also unique when repo_id is known
    conn.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS idx_pulled_models_repo_tag
        ON pulled_models(repo_id, tag)
        WHERE repo_id IS NOT NULL
    """)

    # Migrate existing databases that are missing new columns
    cols = {row["name"] for row in conn.execute("PRAGMA table_info(pulled_models)").fetchall()}
    if "digest" not in cols:
        conn.execute("ALTER TABLE pulled_models ADD COLUMN digest TEXT")
    if "repo_id" not in cols:
        conn.execute("ALTER TABLE pulled_models ADD COLUMN repo_id INTEGER")
    if "stale" not in cols:
        conn.execute("ALTER TABLE pulled_models ADD COLUMN stale INTEGER NOT NULL DEFAULT 0")
    conn.commit()


@contextmanager
def _connect() -> Iterator[sqlite3.Connection]:
    """Open a database connection, initialize schema, and ensure it is closed."""
    conn = sqlite3.connect(_db_path())
    conn.row_factory = sqlite3.Row
    _init_schema(conn)
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()


def save_model(
    owner: str,
    repo: str,
    tag: str,
    content: str,
    digest: str | None = None,
    repo_id: int | None = None,
    stale: int = 0,
) -> None:
    """Insert or update a pulled model in the local cache.

    If repo_id is provided and an existing row for (repo_id, tag) is found
    under a different owner/repo (rename case), that row is updated in place
    so stale refs are not left behind.
    """
    pulled_at = datetime.now(timezone.utc).isoformat()
    with _connect() as conn:
        # Handle rename: if repo_id is known and the row exists under a different slug
        if repo_id is not None:
            cursor = conn.execute(
                """
                UPDATE pulled_models
                SET owner=?, repo=?, content=?, digest=?, pulled_at=?, stale=0
                WHERE repo_id=? AND tag=? AND (owner!=? OR repo!=?)
                """,
                (owner, repo, content, digest, pulled_at, repo_id, tag, owner, repo),
            )
            if cursor.rowcount > 0:
                logger.debug(
                    "Renamed cache entry for repo_id=%s tag=%s to %s/%s",
                    repo_id, tag, owner, repo,
                )
                return

        # Standard upsert by (owner, repo, tag)
        conn.execute(
            """
            INSERT INTO pulled_models (repo_id, owner, repo, tag, content, digest, pulled_at, stale)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(owner, repo, tag) DO UPDATE SET
                repo_id   = COALESCE(excluded.repo_id, repo_id),
                content   = excluded.content,
                digest    = excluded.digest,
                pulled_at = excluded.pulled_at,
                stale     = excluded.stale
            """,
            (repo_id, owner, repo, tag, content, digest, pulled_at, stale),
        )
    logger.debug("Cached %s/%s:%s (repo_id=%s, digest=%s)", owner, repo, tag, repo_id, digest)


def get_model(owner: str, repo: str, tag: str) -> str | None:
    """Return the cached JSON content for owner/repo:tag, or None if not cached."""
    info = get_model_info(owner, repo, tag)
    return info["content"] if info else None


def get_model_info(
    owner: str,
    repo: str,
    tag: str,
    repo_id: int | None = None,
) -> dict | None:
    """Return the full cached row dict for a model, or None if not cached.

    When repo_id is provided, it also tries lookup by (repo_id, tag) for
    rename-resilient cache hits.
    """
    with _connect() as conn:
        row = None

        if repo_id is not None:
            row = conn.execute(
                "SELECT * FROM pulled_models WHERE repo_id=? AND tag=?",
                (repo_id, tag),
            ).fetchone()

        if row is None:
            row = conn.execute(
                "SELECT * FROM pulled_models WHERE owner=? AND repo=? AND tag=?",
                (owner, repo, tag),
            ).fetchone()

    if row:
        logger.debug("Cache hit for %s/%s:%s (repo_id=%s)", owner, repo, tag, repo_id)
        return dict(row)
    logger.debug("Cache miss for %s/%s:%s (repo_id=%s)", owner, repo, tag, repo_id)
    return None


def mark_stale(owner: str, repo: str, tag: str, repo_id: int | None = None) -> bool:
    """Mark a single cache entry as stale (e.g. after remote tag deletion).

    Returns True if an entry was found and marked.
    """
    with _connect() as conn:
        if repo_id is not None:
            cursor = conn.execute(
                "UPDATE pulled_models SET stale=1 WHERE repo_id=? AND tag=?",
                (repo_id, tag),
            )
            if cursor.rowcount > 0:
                logger.debug("Marked stale: repo_id=%s tag=%s", repo_id, tag)
                return True
        cursor = conn.execute(
            "UPDATE pulled_models SET stale=1 WHERE owner=? AND repo=? AND tag=?",
            (owner, repo, tag),
        )
        found = cursor.rowcount > 0
        if found:
            logger.debug("Marked stale: %s/%s:%s", owner, repo, tag)
        return found


def mark_stale_by_repo(owner: str, repo: str, repo_id: int | None = None) -> int:
    """Mark all cache entries for a repository as stale (e.g. after repo deletion).

    Returns the number of entries marked.
    """
    with _connect() as conn:
        if repo_id is not None:
            cursor = conn.execute(
                "UPDATE pulled_models SET stale=1 WHERE repo_id=?",
                (repo_id,),
            )
        else:
            cursor = conn.execute(
                "UPDATE pulled_models SET stale=1 WHERE owner=? AND repo=?",
                (owner, repo),
            )
        count = cursor.rowcount
    logger.debug("Marked %d entries stale for %s/%s", count, owner, repo)
    return count


def remove_stale() -> int:
    """Delete all stale cache entries. Returns the number deleted."""
    with _connect() as conn:
        cursor = conn.execute("DELETE FROM pulled_models WHERE stale=1")
        count = cursor.rowcount
    logger.debug("Removed %d stale cache entries", count)
    return count


def remove_all() -> int:
    """Delete ALL cache entries. Returns the number deleted."""
    with _connect() as conn:
        cursor = conn.execute("DELETE FROM pulled_models")
        count = cursor.rowcount
    logger.debug("Cleared entire local cache (%d entries)", count)
    return count


def list_models(filter_owner: str | None = None, filter_repo: str | None = None) -> list[dict]:
    """Return cached models as a list of dicts, newest first.

    When filter_owner and filter_repo are both provided, only entries for that
    owner/repo are returned.
    """
    with _connect() as conn:
        if filter_owner and filter_repo:
            rows = conn.execute(
                """
                SELECT repo_id, owner, repo, tag, digest, pulled_at, stale
                FROM pulled_models WHERE owner=? AND repo=? ORDER BY pulled_at DESC
                """,
                (filter_owner, filter_repo),
            ).fetchall()
        else:
            rows = conn.execute(
                """
                SELECT repo_id, owner, repo, tag, digest, pulled_at, stale
                FROM pulled_models ORDER BY pulled_at DESC
                """
            ).fetchall()
    return [dict(r) for r in rows]


def remove_model(owner: str, repo: str, tag: str, repo_id: int | None = None) -> bool:
    """Remove a cached model. Returns True if it existed and was deleted."""
    with _connect() as conn:
        if repo_id is not None:
            cursor = conn.execute(
                "DELETE FROM pulled_models WHERE repo_id=? AND tag=?",
                (repo_id, tag),
            )
            if cursor.rowcount > 0:
                logger.debug("Removed repo_id=%s tag=%s from cache", repo_id, tag)
                return True
        cursor = conn.execute(
            "DELETE FROM pulled_models WHERE owner=? AND repo=? AND tag=?",
            (owner, repo, tag),
        )
        found = cursor.rowcount > 0
    logger.debug("Removed %s/%s:%s from cache: %s", owner, repo, tag, found)
    return found


def find_by_digest(digest: str) -> list[dict]:
    """Return all cached entries that share the given digest."""
    with _connect() as conn:
        rows = conn.execute(
            "SELECT owner, repo, tag, pulled_at FROM pulled_models WHERE digest=?",
            (digest,),
        ).fetchall()
    return [dict(r) for r in rows]
