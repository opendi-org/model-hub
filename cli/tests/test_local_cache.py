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
    """list_models entries contain at least owner, repo, tag, digest, pulled_at, stale, repo_id."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        result = local_cache.list_models()
    required = {"owner", "repo", "tag", "digest", "pulled_at", "stale", "repo_id"}
    assert required.issubset(set(result[0].keys()))
    assert result[0]["owner"] == "alice"
    assert result[0]["repo"] == "my-repo"
    assert result[0]["tag"] == "v1"
    assert result[0]["stale"] == 0
    assert result[0]["repo_id"] is None


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


# ── repo_id / stale ───────────────────────────────────────────────────────────


def test_save_model_stores_repo_id(tmp_path: Path) -> None:
    """save_model stores repo_id and retrieval works by repo_id."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", '{"x": 1}', repo_id=42)
        info = local_cache.get_model_info("alice", "my-repo", "v1", repo_id=42)
    assert info is not None
    assert info["repo_id"] == 42
    assert info["content"] == '{"x": 1}'


def test_get_model_info_by_repo_id_rename(tmp_path: Path) -> None:
    """get_model_info finds entry by repo_id even when owner/slug have changed."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "content", repo_id=42)
        # Simulate rename: look up with new owner/slug but same repo_id
        info = local_cache.get_model_info("bob", "renamed-repo", "v1", repo_id=42)
    assert info is not None
    assert info["content"] == "content"


def test_save_model_rename_updates_owner_slug(tmp_path: Path) -> None:
    """save_model updates owner/slug when repo_id+tag matches but owner/slug differ."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "old-repo", "v1", "old", repo_id=42)
        local_cache.save_model("alice", "new-repo", "v1", "new", repo_id=42)
        result = local_cache.list_models()
        content = local_cache.get_model("alice", "new-repo", "v1")
    assert len(result) == 1
    assert result[0]["repo"] == "new-repo"
    assert content == "new"


def test_mark_stale(tmp_path: Path) -> None:
    """mark_stale sets stale=1; list_models shows stale flag."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        found = local_cache.mark_stale("alice", "my-repo", "v1")
        result = local_cache.list_models()
    assert found is True
    assert result[0]["stale"] == 1


def test_mark_stale_not_found(tmp_path: Path) -> None:
    """mark_stale returns False when entry does not exist."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        found = local_cache.mark_stale("alice", "no-repo", "v1")
    assert found is False


def test_mark_stale_by_repo(tmp_path: Path) -> None:
    """mark_stale_by_repo marks all entries for a repo as stale."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        local_cache.save_model("alice", "my-repo", "v2", "{}")
        local_cache.save_model("bob", "other", "v1", "{}")
        count = local_cache.mark_stale_by_repo("alice", "my-repo")
        models = local_cache.list_models()
    assert count == 2
    stale = [m for m in models if m["stale"] == 1]
    not_stale = [m for m in models if m["stale"] == 0]
    assert len(stale) == 2
    assert len(not_stale) == 1
    assert not_stale[0]["owner"] == "bob"


def test_remove_stale(tmp_path: Path) -> None:
    """remove_stale deletes stale entries and leaves ok entries."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "my-repo", "v1", "{}")
        local_cache.save_model("alice", "my-repo", "v2", "{}")
        local_cache.mark_stale("alice", "my-repo", "v1")
        count = local_cache.remove_stale()
        remaining = local_cache.list_models()
    assert count == 1
    assert len(remaining) == 1
    assert remaining[0]["tag"] == "v2"


def test_remove_all(tmp_path: Path) -> None:
    """remove_all deletes every entry in the cache."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "repo-a", "v1", "{}")
        local_cache.save_model("bob", "repo-b", "latest", "{}")
        count = local_cache.remove_all()
        remaining = local_cache.list_models()
    assert count == 2
    assert remaining == []


def test_list_models_filter_by_repo(tmp_path: Path) -> None:
    """list_models with filter_owner/filter_repo returns only matching entries."""
    with patch("opendi.local_cache._db_path", return_value=_tmp_db(tmp_path)):
        local_cache.save_model("alice", "repo-a", "v1", "{}")
        local_cache.save_model("alice", "repo-a", "v2", "{}")
        local_cache.save_model("bob", "repo-b", "v1", "{}")
        result = local_cache.list_models(filter_owner="alice", filter_repo="repo-a")
    assert len(result) == 2
    assert all(r["owner"] == "alice" and r["repo"] == "repo-a" for r in result)
