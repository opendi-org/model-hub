"""Tests for opendi.local_cache (SQLite model cache)."""

from pathlib import Path
from unittest.mock import patch

from opendi import local_cache


def _tmp_db(tmp_path: Path) -> Path:
    """Return a temp DB path for patching _db_path."""
    return tmp_path / "models.db"


# ── save_model / get_model ────────────────────────────────────────────────────


def test_save_model_and_get(tmp_path: Path) -> None:
    """save_model stores content; get_model retrieves it."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", '{"hello": "world"}')
        result = local_cache.get_model("alice", "my-repo", "v1")
    assert result == '{"hello": "world"}'


def test_save_model_upsert(tmp_path: Path) -> None:
    """Saving the same owner/repo/tag twice updates content, no duplicate rows."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "old")
        local_cache.save_model("alice", "my-repo", "v1", "new")
        result = local_cache.get_model("alice", "my-repo", "v1")
        all_models = local_cache.list_models()
    assert result == "new"
    assert len(all_models) == 1


def test_get_model_not_found(tmp_path: Path) -> None:
    """get_model returns None when the model is not cached."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        result = local_cache.get_model("alice", "my-repo", "v1")
    assert result is None


# ── list_models ───────────────────────────────────────────────────────────────


def test_list_models_empty(tmp_path: Path) -> None:
    """list_models returns [] when the cache is empty."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        result = local_cache.list_models()
    assert result == []


def test_list_models_returns_all(tmp_path: Path) -> None:
    """list_models returns one entry per cached model."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "repo-a", "v1", "{}")
        local_cache.save_model("bob", "repo-b", "latest", "{}")
        result = local_cache.list_models()
    assert len(result) == 2
    slugs = {r["repo"] for r in result}
    assert slugs == {"repo-a", "repo-b"}


def test_list_models_fields(tmp_path: Path) -> None:
    """list_models entries contain owner, repo, tag, digest, pulled_at keys."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        result = local_cache.list_models()
    assert set(result[0].keys()) == {"owner", "repo", "tag", "digest", "pulled_at"}
    assert result[0]["owner"] == "alice"
    assert result[0]["repo"] == "my-repo"
    assert result[0]["tag"] == "v1"


# ── find_by_digest ────────────────────────────────────────────────────────────


def test_find_by_digest_returns_matches(tmp_path: Path) -> None:
    """find_by_digest returns all entries with the given digest."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "repo-a", "v1", "{}", digest="abc123")
        local_cache.save_model("alice", "repo-a", "v2", "{}", digest="abc123")
        local_cache.save_model("bob", "repo-b", "v1", "{}", digest="other")
        result = local_cache.find_by_digest("abc123")
    assert len(result) == 2
    tags = {r["tag"] for r in result}
    assert tags == {"v1", "v2"}


def test_find_by_digest_no_match(tmp_path: Path) -> None:
    """find_by_digest returns [] when no entry matches the digest."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        result = local_cache.find_by_digest("nonexistent")
    assert result == []


# ── remove_model ──────────────────────────────────────────────────────────────


def test_remove_model_returns_true_and_deletes(tmp_path: Path) -> None:
    """remove_model returns True and removes the entry when it exists."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        result = local_cache.remove_model("alice", "my-repo", "v1")
        after = local_cache.get_model("alice", "my-repo", "v1")
    assert result is True
    assert after is None


def test_remove_model_returns_false_when_not_found(tmp_path: Path) -> None:
    """remove_model returns False when the entry does not exist."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        result = local_cache.remove_model("alice", "my-repo", "v1")
    assert result is False
