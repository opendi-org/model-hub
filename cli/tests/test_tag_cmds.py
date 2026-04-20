"""Tests for tag subgroup commands: add tag, delete tag, delete local, list local."""

from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()

_TOKEN = "fake-access-token"


def _logged_in():
    return patch("opendi.main.credential_storage.load_access_token", return_value=_TOKEN)


def _logged_out():
    return patch("opendi.main.credential_storage.load_access_token", return_value=None)


def _ok(data: dict, status: int = 200):
    m = MagicMock()
    m.status_code = status
    m.ok = status < 400
    m.json.return_value = data
    m.text = ""
    return m


def _err(status: int, error: str = ""):
    m = MagicMock()
    m.status_code = status
    m.ok = False
    m.json.return_value = {"error": error} if error else {}
    m.text = error
    return m


# ── add tag ───────────────────────────────────────────────────────────────────


def test_add_tag_success() -> None:
    """opendi add tag creates a new alias tag."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", return_value=_ok({}, 201)),
    ):
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"])
    assert result.exit_code == 0
    assert "Added" in result.output
    assert "stable" in result.output


def test_add_tag_overwrite_confirmed() -> None:
    """opendi add tag prompts on 409 and retries with overwrite=true."""
    mock_409 = _err(409)
    mock_201 = _ok({}, 201)
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", side_effect=[mock_409, mock_201]),
    ):
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"], input="y\n")
    assert result.exit_code == 0
    assert "Added" in result.output


def test_add_tag_overwrite_declined() -> None:
    """opendi add tag aborts when user declines 409 overwrite."""
    mock_409 = _err(409)
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", return_value=mock_409),
    ):
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"], input="n\n")
    assert result.exit_code == 0
    assert "Operation canceled" in result.output


def test_add_tag_yes_flag_skips_overwrite_prompt() -> None:
    """opendi add tag --yes skips confirmation on 409."""
    mock_409 = _err(409)
    mock_201 = _ok({}, 201)
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", side_effect=[mock_409, mock_201]),
    ):
        result = runner.invoke(app, ["add", "tag", "--yes", "alice/my-model:v1", "stable"])
    assert result.exit_code == 0


def test_add_tag_requires_login() -> None:
    """opendi add tag exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_add_tag_source_not_found() -> None:
    """opendi add tag exits 1 when source tag doesn't exist."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", return_value=_err(404)),
    ):
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_add_tag_access_denied() -> None:
    """opendi add tag exits 1 on 403."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.put", return_value=_err(403)),
    ):
        result = runner.invoke(app, ["add", "tag", "alice/my-model:v1", "stable"])
    assert result.exit_code == 1
    assert "Not authorized" in result.output


def test_add_tag_invalid_ref() -> None:
    """opendi add tag exits 1 when source ref is missing tag."""
    with _logged_in():
        result = runner.invoke(app, ["add", "tag", "alice/my-model", "stable"])
    assert result.exit_code == 1


# ── delete tag ────────────────────────────────────────────────────────────────


def test_delete_tag_success() -> None:
    """opendi delete tag deletes remote tag and marks cache stale."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.delete", return_value=_ok({}, 204)),
        patch("opendi.cmds.tag_cmds.local_cache.mark_stale", return_value=True),
    ):
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1", "--yes"])
    assert result.exit_code == 0
    assert "Deleted" in result.output
    assert "v1" in result.output


def test_delete_tag_prompts_without_yes() -> None:
    """opendi delete tag prompts for confirmation when --yes is not passed."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.delete", return_value=_ok({}, 204)),
        patch("opendi.cmds.tag_cmds.local_cache.mark_stale", return_value=True),
    ):
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1"], input="y\n")
    assert result.exit_code == 0


def test_delete_tag_aborts_on_decline() -> None:
    """opendi delete tag aborts when user declines the confirmation prompt."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.delete") as mock_del,
    ):
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1"], input="n\n")
    assert result.exit_code != 0
    mock_del.assert_not_called()


def test_delete_tag_requires_login() -> None:
    """opendi delete tag exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_tag_not_found() -> None:
    """opendi delete tag exits 1 on 404."""
    with (
        _logged_in(),
        patch("opendi.cmds.tag_cmds.requests.delete", return_value=_err(404)),
    ):
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1", "--yes"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_delete_tag_connection_error() -> None:
    """opendi delete tag exits 1 when hub is unreachable."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.tag_cmds.requests.delete",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["delete", "tag", "alice/my-model:v1", "--yes"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


# ── delete local ──────────────────────────────────────────────────────────────


def test_delete_local_success() -> None:
    """opendi delete local removes a specific cached entry."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.remove_model", return_value=True),
    ):
        result = runner.invoke(app, ["delete", "local", "alice/my-model:v1"])
    assert result.exit_code == 0
    assert "Removed" in result.output


def test_delete_local_not_found() -> None:
    """opendi delete local exits 1 when entry is not in local cache."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.remove_model", return_value=False),
    ):
        result = runner.invoke(app, ["delete", "local", "alice/my-model:v1"])
    assert result.exit_code == 1
    assert "not in local cache" in result.output.lower()


def test_delete_local_all_confirmed() -> None:
    """opendi delete local --all clears entire cache after confirmation."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.remove_all", return_value=5),
    ):
        result = runner.invoke(app, ["delete", "local", "--all"], input="y\n")
    assert result.exit_code == 0
    assert "5 entries removed" in result.output


def test_delete_local_all_yes_flag() -> None:
    """opendi delete local --all --yes skips confirmation."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.remove_all", return_value=3),
    ):
        result = runner.invoke(app, ["delete", "local", "--all", "--yes"])
    assert result.exit_code == 0
    assert "3 entries removed" in result.output


def test_delete_local_both_ref_and_all_exits_1() -> None:
    """opendi delete local exits 1 when both REF and --all are specified."""
    with _logged_out():
        result = runner.invoke(app, ["delete", "local", "alice/my-model:v1", "--all"])
    assert result.exit_code == 1
    assert "not both" in result.output


def test_delete_local_no_args_exits_1() -> None:
    """opendi delete local exits 1 when neither REF nor --all is specified."""
    with _logged_out():
        result = runner.invoke(app, ["delete", "local"])
    assert result.exit_code == 1


# ── list local ────────────────────────────────────────────────────────────────


_CACHED_MODELS = [
    {
        "owner": "alice",
        "repo": "my-model",
        "tag": "v1",
        "digest": "abc123",
        "pulled_at": "2026-01-01T00:00:00+00:00",
        "stale": 0,
        "repo_id": 42,
    },
    {
        "owner": "bob",
        "repo": "other",
        "tag": "latest",
        "digest": "dead000",
        "pulled_at": "2026-01-02T00:00:00+00:00",
        "stale": 1,
        "repo_id": None,
    },
]


def test_list_local_shows_all_entries() -> None:
    """opendi list local shows all cached models including stale."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.list_models", return_value=_CACHED_MODELS),
    ):
        result = runner.invoke(app, ["list", "local"])
    assert result.exit_code == 0
    assert "alice/my-model" in result.output
    assert "bob/other" in result.output
    assert "stale" in result.output
    assert "v1" in result.output


def test_list_local_empty() -> None:
    """opendi list local prints helpful message when cache is empty."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.list_models", return_value=[]),
    ):
        result = runner.invoke(app, ["list", "local"])
    assert result.exit_code == 0
    assert "No models in local cache" in result.output


def test_list_local_filter_by_repo() -> None:
    """opendi list local filters by owner/repo argument."""
    with (
        _logged_out(),
        patch(
            "opendi.cmds.tag_cmds.local_cache.list_models",
            return_value=[_CACHED_MODELS[0]],
        ) as mock_list,
    ):
        result = runner.invoke(app, ["list", "local", "alice/my-model"])
    assert result.exit_code == 0
    assert "alice/my-model" in result.output
    _, kwargs = mock_list.call_args
    assert kwargs.get("filter_owner") == "alice"
    assert kwargs.get("filter_repo") == "my-model"


def test_list_local_clean_removes_stale() -> None:
    """opendi list local --clean removes stale entries and prints count."""
    with (
        _logged_out(),
        patch("opendi.cmds.tag_cmds.local_cache.remove_stale", return_value=2),
    ):
        result = runner.invoke(app, ["list", "local", "--clean"])
    assert result.exit_code == 0
    assert "2" in result.output
    assert "stale" in result.output


def test_list_local_invalid_filter_format() -> None:
    """opendi list local exits 1 when filter is not owner/repo format."""
    with _logged_out():
        result = runner.invoke(app, ["list", "local", "my-model"])
    assert result.exit_code == 1
    assert "owner/repo" in result.output
