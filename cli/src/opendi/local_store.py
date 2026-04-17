"""Local SQLite cache for pulled models.

Database location: OS-appropriate user data directory (e.g. %APPDATA%\\opendi\\models.db
on Windows, ~/.local/share/opendi/models.db on Linux, ~/Library/Application Support/opendi/models.db on macOS).

Schema:
  pulled_models
    id          INTEGER PRIMARY KEY AUTOINCREMENT
    owner       TEXT NOT NULL
    repo        TEXT NOT NULL
    tag         TEXT NOT NULL
    content     TEXT NOT NULL   -- raw JSON string
    digest      TEXT            -- ETag/UUID from hub (nullable for legacy rows)
    pulled_at   TEXT NOT NULL   -- ISO-8601 UTC timestamp
    UNIQUE(owner, repo, tag)    -- upsert on re-pull
"""

import sqlite3
from datetime import datetime, timezone
from pathlib import Path

from platformdirs import user_data_dir


def _db_path() -> Path:
    path = Path(user_data_dir("opendi", appauthor=False)) / "models.db"
    path.parent.mkdir(parents=True, exist_ok=True)
    return path


def _connect() -> sqlite3.Connection:
    conn = sqlite3.connect(_db_path())
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("""
        CREATE TABLE IF NOT EXISTS pulled_models (
            id        INTEGER PRIMARY KEY AUTOINCREMENT,
            owner     TEXT NOT NULL,
            repo      TEXT NOT NULL,
            tag       TEXT NOT NULL,
            content   TEXT NOT NULL,
            digest    TEXT,
            pulled_at TEXT NOT NULL,
            UNIQUE(owner, repo, tag)
        )
    """)
    # Migrate existing databases that lack the digest column
    cols = {row[1] for row in conn.execute("PRAGMA table_info(pulled_models)").fetchall()}
    if "digest" not in cols:
        conn.execute("ALTER TABLE pulled_models ADD COLUMN digest TEXT")
    conn.commit()
    return conn


def save_model(owner: str, repo: str, tag: str, content: str, digest: str | None = None) -> None:
    """Insert or replace a pulled model in the local cache."""
    pulled_at = datetime.now(timezone.utc).isoformat()
    with _connect() as conn:
        conn.execute(
            """
            INSERT INTO pulled_models (owner, repo, tag, content, digest, pulled_at)
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT(owner, repo, tag) DO UPDATE SET
                content   = excluded.content,
                digest    = excluded.digest,
                pulled_at = excluded.pulled_at
            """,
            (owner, repo, tag, content, digest, pulled_at),
        )


def get_model(owner: str, repo: str, tag: str) -> str | None:
    """Return the cached JSON content for owner/repo:tag, or None if not cached."""
    with _connect() as conn:
        row = conn.execute(
            "SELECT content FROM pulled_models WHERE owner=? AND repo=? AND tag=?",
            (owner, repo, tag),
        ).fetchone()
    return row[0] if row else None


def find_by_digest(digest: str) -> list[dict]:
    """Return all cached entries that share the given digest."""
    with _connect() as conn:
        rows = conn.execute(
            "SELECT owner, repo, tag, pulled_at FROM pulled_models WHERE digest=?",
            (digest,),
        ).fetchall()
    return [{"owner": r[0], "repo": r[1], "tag": r[2], "pulled_at": r[3]} for r in rows]


def list_models() -> list[dict]:
    """Return all cached models as a list of dicts (owner, repo, tag, digest, pulled_at)."""
    with _connect() as conn:
        rows = conn.execute(
            "SELECT owner, repo, tag, digest, pulled_at FROM pulled_models ORDER BY pulled_at DESC"
        ).fetchall()
    return [{"owner": r[0], "repo": r[1], "tag": r[2], "digest": r[3], "pulled_at": r[4]} for r in rows]


def remove_model(owner: str, repo: str, tag: str) -> bool:
    """Remove a cached model. Returns True if it existed and was deleted, False if not found."""
    with _connect() as conn:
        cursor = conn.execute(
            "DELETE FROM pulled_models WHERE owner=? AND repo=? AND tag=?",
            (owner, repo, tag),
        )
    return cursor.rowcount > 0
