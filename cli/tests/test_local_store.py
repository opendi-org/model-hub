"""Tests for opendi.local_store (SQLite model cache)."""

import tempfile
from pathlib import Path
from unittest.mock import patch

from opendi import local_store


def _tmp_db(tmp_path: Path) -> Path:
    """Return a temp DB path and patch _db_path to use it."""
    return tmp_path / "models.db"


# ── save_model / get_model ────────────────────────────────────────────────────


def test_save_model_and_get(tmp_path: Path) -> None:
    """save_model stores content; get_model retrieves it."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        local_store.save_model("alice", "my-repo", "v1", '{"hello": "world"}')
        result = local_store.get_model("alice", "my-repo", "v1")
    assert result == '{"hello": "world"}'


def test_save_model_upsert(tmp_path: Path) -> None:
    """Saving the same owner/repo/tag twice updates content, no duplicate rows."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        local_store.save_model("alice", "my-repo", "v1", "old")
        local_store.save_model("alice", "my-repo", "v1", "new")
        result = local_store.get_model("alice", "my-repo", "v1")
        all_models = local_store.list_models()
    assert result == "new"
    assert len(all_models) == 1


def test_get_model_not_found(tmp_path: Path) -> None:
    """get_model returns None when the model is not cached."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        result = local_store.get_model("alice", "my-repo", "v1")
    assert result is None


# ── list_models ───────────────────────────────────────────────────────────────


def test_list_models_empty(tmp_path: Path) -> None:
    """list_models returns [] when the cache is empty."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        result = local_store.list_models()
    assert result == []


def test_list_models_returns_all(tmp_path: Path) -> None:
    """list_models returns one entry per cached model."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        local_store.save_model("alice", "repo-a", "v1", "{}")
        local_store.save_model("bob", "repo-b", "latest", "{}")
        result = local_store.list_models()
    assert len(result) == 2
    slugs = {r["repo"] for r in result}
    assert slugs == {"repo-a", "repo-b"}


def test_list_models_fields(tmp_path: Path) -> None:
    """list_models entries contain owner, repo, tag, pulled_at keys."""
    with patch("opendi.local_store._db_path", return_value=_tmp_db(tmp_path)):
        local_store.save_model("alice", "my-repo", "v1", "{}")
        result = local_store.list_models()
    assert set(result[0].keys()) == {"owner", "repo", "tag", "digest", "pulled_at"}
    assert result[0]["owner"] == "alice"
    assert result[0]["repo"] == "my-repo"
    assert result[0]["tag"] == "v1"
