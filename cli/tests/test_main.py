"""Tests for the CLI entry point (flat commands: login, logout, whoami, search, pull, push, save, diff, validate)."""

import json
import tempfile
from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()

_TOKEN = "fake-access-token"
# command to check with coverage
# pytest tests/ --cov=opendi --cov-report=term-missing -q


# ── Helpers ───────────────────────────────────────────────────────────────────


def _logged_in():
    """Patch load_access_token to return a valid token."""
    return patch("opendi.main.credential_storage.load_access_token", return_value=_TOKEN)


def _logged_out():
    """Patch load_access_token to return None (not logged in)."""
    return patch("opendi.main.credential_storage.load_access_token", return_value=None)


def _remote_ok(data: dict):
    """Mock a successful 200 response returning JSON data."""
    m = MagicMock()
    m.status_code = 200
    m.ok = True
    m.json.return_value = data
    return m


def _remote_err(status: int, error: str = ""):
    """Mock an error response."""
    m = MagicMock()
    m.status_code = status
    m.ok = False
    m.json.return_value = {"error": error} if error else {}
    m.text = error
    return m


# ── Help / smoke ──────────────────────────────────────────────────────────────


def test_app_help_exits_zero() -> None:
    """opendi --help exits with code 0."""
    result = runner.invoke(app, ["--help"])
    assert result.exit_code == 0
    assert "opendi" in result.output.lower()


def test_app_without_command_shows_help() -> None:
    """opendi (no subcommand) shows full help."""
    result = runner.invoke(app, [])
    assert "Usage:" in result.output
    assert "login" in result.output
    assert result.exit_code == 2


# ── Login ─────────────────────────────────────────────────────────────────────


def test_login_success() -> None:
    """opendi login stores token and prints success with username."""
    with (
        _logged_out(),
        patch("opendi.main.auth.start_cli_login", return_value=("code123", "/login", 300)),
        patch("opendi.main.auth.open_login_url", return_value="http://localhost/login"),
        patch("opendi.main.auth.poll_cli_token", return_value=_TOKEN),
        patch("opendi.main.auth.get_current_user", return_value={"username": "alice"}),
        patch("opendi.main.credential_storage.store_access_token") as mock_store,
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Please sign in using your browser" in result.output
    assert "this link" in result.output
    assert "Login successful" in result.output
    assert "alice" in result.output
    mock_store.assert_called_once_with(_TOKEN)


def test_login_success_no_username_fallback() -> None:
    """opendi login prints success without username when API returns no username."""
    with (
        _logged_out(),
        patch("opendi.main.auth.start_cli_login", return_value=("code123", "/login", 300)),
        patch("opendi.main.auth.open_login_url", return_value="http://localhost/login"),
        patch("opendi.main.auth.poll_cli_token", return_value=_TOKEN),
        patch("opendi.main.auth.get_current_user", return_value={}),
        patch("opendi.main.credential_storage.store_access_token"),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "this link" in result.output
    assert "Login successful." in result.output


def test_login_already_logged_in() -> None:
    """opendi login prints 'Already logged in' when token is valid."""
    with (
        _logged_in(),
        patch("opendi.main.auth.get_current_user", return_value={"username": "alice"}),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Already logged in" in result.output
    assert "alice" in result.output


def test_login_timeout() -> None:
    """opendi login exits 1 with timeout message when polling times out."""
    with (
        _logged_out(),
        patch("opendi.main.auth.start_cli_login", return_value=("code123", "/login", 300)),
        patch("opendi.main.auth.open_login_url", return_value="http://localhost/login"),
        patch("opendi.main.auth.poll_cli_token", side_effect=TimeoutError()),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "timed out" in result.output


def test_login_cancelled() -> None:
    """opendi login exits 1 with 'Login cancelled' on KeyboardInterrupt."""
    with (
        _logged_out(),
        patch("opendi.main.auth.start_cli_login", side_effect=KeyboardInterrupt()),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "cancelled" in result.output


def test_login_generic_error() -> None:
    """opendi login exits 1 with error message on unexpected failure."""
    with (
        _logged_out(),
        patch("opendi.main.auth.start_cli_login", side_effect=RuntimeError("network error")),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "Login failed" in result.output


# ── Whoami ────────────────────────────────────────────────────────────────────


def test_whoami_not_logged_in() -> None:
    """opendi whoami exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "Not logged in" in result.output
    assert "opendi login" in result.output


def test_whoami_shows_username() -> None:
    """opendi whoami prints username when logged in."""
    with (
        _logged_in(),
        patch("opendi.main.auth.get_current_user", return_value={"username": "alice"}),
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 0
    assert "alice" in result.output


def test_whoami_shows_email_fallback() -> None:
    """opendi whoami prints email when username is absent."""
    with (
        _logged_in(),
        patch("opendi.main.auth.get_current_user", return_value={"email": "alice@example.com"}),
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 0
    assert "alice@example.com" in result.output


def test_whoami_session_expired() -> None:
    """opendi whoami exits 1 and clears token when /me call fails."""
    with (
        _logged_in(),
        patch("opendi.main.auth.get_current_user", side_effect=Exception("unauthorized")),
        patch("opendi.main.credential_storage.delete_access_token") as mock_delete,
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "Session expired" in result.output
    assert "opendi login" in result.output
    mock_delete.assert_called_once()


# ── Logout ────────────────────────────────────────────────────────────────────


def test_logout_success() -> None:
    """opendi logout deletes stored token."""
    with patch("opendi.main.credential_storage.delete_access_token", return_value=True):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Logged out" in result.output


def test_logout_not_logged_in() -> None:
    """opendi logout prints 'Not logged in' when no token stored."""
    with patch("opendi.main.credential_storage.delete_access_token", return_value=False):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Not logged in" in result.output


# ── Diff ──────────────────────────────────────────────────────────────────────


def test_diff_two_identical_remote_models() -> None:
    """opendi diff prints 'No differences' when both sides are identical."""
    model = {"name": "my-model", "version": "1"}
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_ok(model)),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v1"])
    assert result.exit_code == 0
    assert "No differences" in result.output


def test_diff_two_different_remote_models() -> None:
    """opendi diff prints a unified diff when models differ."""
    left = {"name": "model-a"}
    right = {"name": "model-b"}
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=[_remote_ok(left), _remote_ok(right)]),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 0
    assert "model-a" in result.output
    assert "model-b" in result.output


def test_diff_local_file_vs_remote(tmp_path) -> None:
    """opendi diff works with a local JSON file on one side."""
    local = tmp_path / "model.json"
    local.write_text('{"name": "local"}', encoding="utf-8")
    remote = {"name": "remote"}
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_ok(remote)),
    ):
        result = runner.invoke(app, ["diff", str(local), "alice/repo:v1"])
    assert result.exit_code == 0


def test_diff_two_local_files_identical(tmp_path) -> None:
    """opendi diff works with two local files."""
    f = tmp_path / "model.json"
    f.write_text('{"x": 1}', encoding="utf-8")
    result = runner.invoke(app, ["diff", str(f), str(f)])
    assert result.exit_code == 0
    assert "No differences" in result.output


def test_diff_local_file_invalid_json(tmp_path) -> None:
    """opendi diff exits 1 when a local file contains invalid JSON."""
    f = tmp_path / "bad.json"
    f.write_text("not json", encoding="utf-8")
    result = runner.invoke(app, ["diff", str(f), str(f)])
    assert result.exit_code == 1
    assert "Invalid JSON" in result.output


def test_diff_local_file_not_found() -> None:
    """opendi diff exits 1 when a local file path does not exist."""
    result = runner.invoke(app, ["diff", "/no/such/file.json", "/no/such/file.json"])
    assert result.exit_code == 1
    assert "File not found" in result.output


def test_diff_remote_401() -> None:
    """opendi diff exits 1 with auth message on 401."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_err(401)),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_diff_remote_403() -> None:
    """opendi diff exits 1 with access denied message on 403."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_err(403)),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "Not authorized" in result.output


def test_diff_remote_404_tag_not_found() -> None:
    """opendi diff exits 1 with tag-not-found message on 404."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_err(404, "tag not found")),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "Tag not found" in result.output


def test_diff_remote_404_repo_not_found() -> None:
    """opendi diff exits 1 with repo-not-found message on 404."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_err(404, "repository not found")),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "Repository not found" in result.output


def test_diff_remote_500() -> None:
    """opendi diff exits 1 with HTTP status on unexpected error."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=_remote_err(500)),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_diff_connection_error() -> None:
    """opendi diff exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_diff_timeout() -> None:
    """opendi diff exits 1 when the request times out."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["diff", "alice/repo:v1", "alice/repo:v2"])
    assert result.exit_code == 1
    assert "timed out" in result.output


def test_diff_no_differences(tmp_path) -> None:
    """opendi diff prints a message when normalized JSON matches."""
    doc = {"meta": {"name": "m"}, "$schema": "x"}
    a = tmp_path / "a.json"
    b = tmp_path / "b.json"
    a.write_text(json.dumps(doc), encoding="utf-8")
    b.write_text(json.dumps(doc), encoding="utf-8")
    result = runner.invoke(app, ["diff", str(a), str(b)])
    assert result.exit_code == 0
    assert "No differences" in result.output


def test_diff_prints_unified_diff(tmp_path) -> None:
    """opendi diff prints unified diff when JSON differs."""
    (tmp_path / "a.json").write_text(json.dumps({"a": 1}), encoding="utf-8")
    (tmp_path / "b.json").write_text(json.dumps({"a": 2}), encoding="utf-8")
    result = runner.invoke(app, ["diff", str(tmp_path / "a.json"), str(tmp_path / "b.json")])
    assert result.exit_code == 0
    assert "@@" in result.output or "--- " in result.output


def test_diff_access_denied() -> None:
    """HTTP 401 shows the same sign-in hint as other commands."""
    response = MagicMock()
    response.status_code = 401
    response.ok = False
    response.json.return_value = {"error": "unauthorized"}
    with patch("opendi.main.requests.get", return_value=response):
        result = runner.invoke(app, ["diff", "a/b:c", "x/y:z"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_diff_tag_not_found() -> None:
    """404 with tag not found from the API is reported."""
    ok = MagicMock()
    ok.status_code = 200
    ok.ok = True
    ok.json.return_value = {"k": 1}
    missing = MagicMock()
    missing.status_code = 404
    missing.ok = False
    missing.json.return_value = {"error": "tag not found"}
    with patch("opendi.main.requests.get", side_effect=[ok, missing]):
        result = runner.invoke(app, ["diff", "a/b:c", "owner/repo:bad"])
    assert result.exit_code == 1
    assert "Tag not found" in result.output


def test_diff_invalid_ref(tmp_path) -> None:
    """Non-file path that is not owner/slug:tag exits with a clear error."""
    f = tmp_path / "b.json"
    f.write_text("{}", encoding="utf-8")
    result = runner.invoke(app, ["diff", "not-a-file", str(f)])
    assert result.exit_code == 1
    assert "Expected owner/repo:tag" in result.output


# ── Pull ──────────────────────────────────────────────────────────────────────


def _repo_ok(owner="alice", slug="my-model", repo_id=42):
    """Mock repo resolve response."""
    m = MagicMock()
    m.status_code = 200
    m.ok = True
    m.json.return_value = {"id": repo_id, "owner": owner, "slug": slug, "tags": []}
    return m


def _repo_ok_with_tag_digest(
    owner="alice",
    slug="my-model",
    repo_id=42,
    tag="v1.0",
    digest="digest-abc",
):
    """Mock repo response that includes tag metadata digest."""
    m = MagicMock()
    m.status_code = 200
    m.ok = True
    m.json.return_value = {
        "id": repo_id,
        "owner": owner,
        "slug": slug,
        "tags": [{"name": tag, "digest": digest}],
    }
    return m


def _model_ok(content='{"meta": {}}', digest="digest-abc"):
    """Mock model fetch response."""
    m = MagicMock()
    m.status_code = 200
    m.ok = True
    m.text = content
    m.headers = {"ETag": digest}
    return m


def test_pull_invalid_format() -> None:
    """opendi pull exits 1 when name is not a valid ref."""
    with _logged_out():
        result = runner.invoke(app, ["pull", ""])
    assert result.exit_code == 1


def test_pull_missing_tag() -> None:
    """opendi pull exits 1 when tag is omitted."""
    with _logged_out():
        result = runner.invoke(app, ["pull", "alice/my-model"])
    assert result.exit_code == 1
    assert "Tag is required" in result.output


def test_pull_unauthenticated_public_repo() -> None:
    """opendi pull succeeds without credentials for public repos."""
    with (
        _logged_out(),
        patch("opendi.main.requests.get", side_effect=[_repo_ok(), _model_ok()]),
        patch("opendi.main.local_cache.save_model") as mock_save,
        patch("opendi.main.local_cache.get_model_info", return_value=None),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "Pulled" in result.output
    mock_save.assert_called_once()


def test_pull_fetches_model_and_writes_cache() -> None:
    """opendi pull fetches model payload and writes cache metadata/content."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=[_repo_ok(), _model_ok()]),
        patch("opendi.main.local_cache.get_model_info", return_value=None),
        patch("opendi.main.local_cache.save_model") as mock_save,
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "Pulled" in result.output
    mock_save.assert_called_once()


def test_pull_up_to_date_uses_metadata_digest_without_model_fetch() -> None:
    """opendi pull skips model download when metadata digest matches cache."""
    cached = {"digest": "digest-abc", "content": '{"cached": true}'}
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=[_repo_ok_with_tag_digest()]) as mock_get,
        patch("opendi.main.local_cache.get_model_info", return_value=cached),
        patch("opendi.main.local_cache.save_model") as mock_save,
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "up to date" in result.output
    assert mock_get.call_count == 1
    mock_save.assert_called_once_with(
        "alice",
        "my-model",
        "v1.0",
        '{"cached": true}',
        digest="digest-abc",
        repo_id=42,
        stale=0,
    )


def test_pull_up_to_date_prints_slug_owner_metadata_updates() -> None:
    """opendi pull reports slug/owner metadata updates on digest match."""
    cached = {
        "digest": "digest-abc",
        "content": '{"cached": true}',
        "owner": "old-owner",
        "repo": "old-slug",
    }
    with (
        _logged_in(),
        patch(
            "opendi.main.requests.get",
            side_effect=[_repo_ok_with_tag_digest(owner="alice", slug="my-model", digest="digest-abc")],
        ),
        patch("opendi.main.local_cache.get_model_info", return_value=cached),
        patch("opendi.main.local_cache.save_model"),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "metadata updated: slug old-slug -> my-model" in result.output
    assert "metadata updated: owner old-owner -> alice" in result.output


def test_pull_reports_digest_metadata_update_on_changed_model() -> None:
    """opendi pull reports digest metadata update when model changes."""
    cached = {"digest": "digest-old", "content": '{"cached": true}', "owner": "alice", "repo": "my-model"}
    with (
        _logged_in(),
        patch(
            "opendi.main.requests.get",
            side_effect=[_repo_ok_with_tag_digest(digest="digest-new"), _model_ok(digest="digest-new")],
        ),
        patch("opendi.main.local_cache.get_model_info", return_value=cached),
        patch("opendi.main.local_cache.save_model"),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "metadata updated: digest changed" in result.output


def test_pull_not_found() -> None:
    """opendi pull exits 1 when server returns 404 on repo resolve."""
    repo_404 = MagicMock(status_code=404, ok=False)
    with (
        _logged_out(),
        patch("opendi.main.requests.get", return_value=repo_404),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_pull_access_denied() -> None:
    """opendi pull exits 1 when server returns 403 on repo resolve."""
    repo_403 = MagicMock(status_code=403, ok=False)
    with (
        _logged_out(),
        patch("opendi.main.requests.get", return_value=repo_403),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "Access denied" in result.output


def test_pull_connection_error() -> None:
    """opendi pull exits 1 when server is unreachable."""
    import requests as req_lib
    with (
        _logged_out(),
        patch("opendi.main.requests.get", side_effect=req_lib.ConnectionError()),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


# ── Push ──────────────────────────────────────────────────────────────────────


def test_push_file_not_found() -> None:
    """opendi push exits 1 when the local file does not exist."""
    with _logged_in():
        result = runner.invoke(app, ["push", "alice/my-repo:v1", "/no/such/file.json"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_push_invalid_ref() -> None:
    """opendi push exits 1 when REF is not owner/repo:tag format."""
    with _logged_in():
        result = runner.invoke(app, ["push", "bad-ref", "file.json"])
    assert result.exit_code == 1


def test_push_requires_login() -> None:
    """opendi push exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["push", "alice/my-repo:v1", "file.json"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_push_success() -> None:
    """opendi push uploads file and prints digest and size."""
    mock_put = MagicMock(ok=True, status_code=200)
    mock_put.json.return_value = {"digest": "abc123", "size": 42}
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{"meta": {"uuid": "abc"}}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", return_value=mock_put),
        patch("opendi.main._pull_to_cache"),  # suppress implicit pull
    ):
        result = runner.invoke(app, ["push", "alice/repo:v1.0", tmp_path])
    assert result.exit_code == 0
    assert "abc123" in result.output


def test_push_with_id_at_ref() -> None:
    """opendi push resolves id@repo_id:tag via GET /v0/repo/:id before PUT."""
    mock_put = MagicMock(ok=True, status_code=200)
    mock_put.json.return_value = {"digest": "xyz", "size": 10}
    repo_resolve = _remote_ok({"owner": "alice", "slug": "r-from-id", "id": 99, "tags": []})
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b"{}")
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.cmds.shared.requests.get", return_value=repo_resolve),
        patch("opendi.main.requests.put", return_value=mock_put) as put_fn,
        patch("opendi.main._pull_to_cache"),
    ):
        result = runner.invoke(app, ["push", "id@99:v1", tmp_path, "--yes"])
    assert result.exit_code == 0
    assert "Pushed" in result.output
    put_url = put_fn.call_args[0][0]
    assert "alice" in put_url and "r-from-id" in put_url


def test_push_overwrite_confirmation_yes() -> None:
    """opendi push retries with ?overwrite=true after user confirms 409."""
    mock_409 = MagicMock(ok=False, status_code=409)
    mock_200 = MagicMock(ok=True, status_code=200)
    mock_200.json.return_value = {"digest": "xyz", "size": 10}
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", side_effect=[mock_409, mock_200]),
        patch("opendi.main._pull_to_cache"),
    ):
        result = runner.invoke(app, ["push", "alice/repo:v1", tmp_path], input="y\n")
    assert result.exit_code == 0


def test_push_overwrite_declined() -> None:
    """opendi push aborts when user declines 409 overwrite."""
    mock_409 = MagicMock(ok=False, status_code=409)
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", return_value=mock_409),
    ):
        result = runner.invoke(app, ["push", "alice/repo:v1", tmp_path], input="n\n")
    assert result.exit_code == 0
    assert "Operation canceled" in result.output


def test_push_access_denied() -> None:
    """opendi push exits 1 when server returns 403."""
    mock_403 = MagicMock(ok=False, status_code=403)
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", return_value=mock_403),
    ):
        result = runner.invoke(app, ["push", "alice/repo:v1", tmp_path])
    assert result.exit_code == 1
    assert "Not authorized" in result.output


def test_push_validation_error() -> None:
    """opendi push exits 1 when server returns 400 validation error."""
    mock_400 = MagicMock(ok=False, status_code=400)
    mock_400.json.return_value = {"error": "schema mismatch"}
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", return_value=mock_400),
    ):
        result = runner.invoke(app, ["push", "alice/repo:v1", tmp_path])
    assert result.exit_code == 1
    assert "Validation error" in result.output


def test_push_invalid_json_file() -> None:
    """opendi push exits 1 when the file contains invalid JSON."""
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b"not json")
        tmp_path = f.name
    with _logged_in():
        result = runner.invoke(app, ["push", "alice/repo:v1", tmp_path])
    assert result.exit_code == 1
    assert "Invalid JSON" in result.output


def test_push_yes_flag_skips_confirmation() -> None:
    """opendi push --yes skips overwrite confirmation on 409."""
    mock_409 = MagicMock(ok=False, status_code=409)
    mock_200 = MagicMock(ok=True, status_code=200)
    mock_200.json.return_value = {"digest": "d", "size": 1}
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        _logged_in(),
        patch("opendi.main.requests.put", side_effect=[mock_409, mock_200]),
        patch("opendi.main._pull_to_cache"),
    ):
        result = runner.invoke(app, ["push", "--yes", "alice/repo:v1", tmp_path])
    assert result.exit_code == 0


# ── Save ──────────────────────────────────────────────────────────────────────


def test_save_from_cache(tmp_path) -> None:
    """opendi save writes content from local cache to a file."""
    with (
        _logged_out(),
        patch("opendi.main.local_cache.get_model", return_value='{"k": 1}'),
    ):
        result = runner.invoke(app, ["save", "alice/my-model:v1", str(tmp_path)])
    assert result.exit_code == 0
    out_file = tmp_path / "v1.json"
    assert out_file.exists()
    assert out_file.read_text() == '{"k": 1}'


def test_save_implicit_pull_on_cache_miss(tmp_path) -> None:
    """opendi save pulls from hub when model is not in local cache."""
    with (
        _logged_out(),
        patch("opendi.main.local_cache.get_model", return_value=None),
        patch("opendi.main._pull_to_cache", return_value='{"pulled": true}') as mock_pull,
    ):
        result = runner.invoke(app, ["save", "alice/my-model:v1", str(tmp_path)])
    assert result.exit_code == 0
    mock_pull.assert_called_once()


def test_save_with_output_option(tmp_path) -> None:
    """opendi save -o writes to the exact path specified."""
    out = tmp_path / "my-output.json"
    with (
        _logged_out(),
        patch("opendi.main.local_cache.get_model", return_value='{"k": 1}'),
    ):
        result = runner.invoke(app, ["save", "alice/my-model:v1", "--output", str(out)])
    assert result.exit_code == 0
    assert out.exists()


def test_save_both_output_and_dir_exits_1(tmp_path) -> None:
    """opendi save exits 1 when both OUTPUT_DIR and --output are specified."""
    with _logged_out():
        result = runner.invoke(
            app, ["save", "alice/my-model:v1", str(tmp_path), "--output", "x.json"]
        )
    assert result.exit_code == 1
    assert "not both" in result.output


# ── Validate ──────────────────────────────────────────────────────────────────


def test_validate_valid_file(tmp_path) -> None:
    """opendi validate prints 'Valid CDM.' for a valid file."""
    f = tmp_path / "model.json"
    f.write_text('{"k": 1}', encoding="utf-8")
    mock_resp = MagicMock(status_code=200, ok=True)
    with (
        _logged_out(),
        patch("opendi.main.requests.post", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["validate", str(f)])
    assert result.exit_code == 0
    assert "Valid CDM." in result.output


def test_validate_invalid_json(tmp_path) -> None:
    """opendi validate exits 1 for a file with invalid JSON."""
    f = tmp_path / "bad.json"
    f.write_text("not json", encoding="utf-8")
    result = runner.invoke(app, ["validate", str(f)])
    assert result.exit_code == 1
    assert "Invalid JSON" in result.output


def test_validate_file_not_found() -> None:
    """opendi validate exits 1 when file does not exist."""
    result = runner.invoke(app, ["validate", "/no/such/file.json"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_validate_server_error(tmp_path) -> None:
    """opendi validate exits 1 with error message on API failure."""
    f = tmp_path / "model.json"
    f.write_text('{"k": 1}', encoding="utf-8")
    mock_resp = MagicMock(ok=False, status_code=422)
    mock_resp.json.return_value = {"error": "schema mismatch"}
    with (
        _logged_out(),
        patch("opendi.main.requests.post", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["validate", str(f)])
    assert result.exit_code == 1
    assert "Validation failed" in result.output


def test_validate_server_error_with_details(tmp_path) -> None:
    """opendi validate prints structured issue details (path + line/column) when provided."""
    f = tmp_path / "model.json"
    f.write_text('{"k": 1}', encoding="utf-8")
    mock_resp = MagicMock(ok=False, status_code=400)
    mock_resp.json.return_value = {
        "error": "CDM validation failed (1 issue).",
        "details": [
            {
                "phase": "schema",
                "instancePath": "/meta/uuid",
                "message": "not valid uuid",
                "line": 3,
                "column": 12,
            }
        ],
    }
    with (
        _logged_out(),
        patch("opendi.main.requests.post", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["validate", str(f)])
    assert result.exit_code == 1
    assert "Validation failed" in result.output
    assert "/meta/uuid" in result.output
    assert "line 3, col 12" in result.output


def test_validate_connection_error(tmp_path) -> None:
    """opendi validate exits 1 when hub is unreachable."""
    import requests as req_lib
    f = tmp_path / "model.json"
    f.write_text('{"k": 1}', encoding="utf-8")
    with (
        _logged_out(),
        patch("opendi.main.requests.post", side_effect=req_lib.ConnectionError()),
    ):
        result = runner.invoke(app, ["validate", str(f)])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


# ── Search ────────────────────────────────────────────────────────────────────


def test_search_success() -> None:
    """opendi search prints results in a table."""
    repos = [
        {"owner": "alice", "slug": "my-model", "visibility": "public", "description": "A model"},
        {"owner": "bob", "slug": "other", "visibility": "private", "description": ""},
    ]
    mock_resp = MagicMock(ok=True, status_code=200)
    mock_resp.json.return_value = {"repositories": repos}
    with (
        _logged_out(),
        patch("opendi.main.requests.get", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["search", "model"])
    assert result.exit_code == 0
    assert "alice/my-model" in result.output
    assert "bob/other" in result.output


def test_search_no_results() -> None:
    """opendi search prints 'No repositories found' when API returns empty list."""
    mock_resp = MagicMock(ok=True, status_code=200)
    mock_resp.json.return_value = {"repositories": []}
    with (
        _logged_out(),
        patch("opendi.main.requests.get", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["search", "noresults"])
    assert result.exit_code == 0
    assert "No repositories found" in result.output


def test_search_connection_error() -> None:
    """opendi search exits 1 when the hub is unreachable."""
    with (
        _logged_out(),
        patch("opendi.main.requests.get", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["search", "x"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_search_limit() -> None:
    """opendi search --limit caps the number of results shown."""
    repos = [{"owner": "alice", "slug": f"repo-{i}", "visibility": "public", "description": ""} for i in range(10)]
    mock_resp = MagicMock(ok=True, status_code=200)
    mock_resp.json.return_value = {"repositories": repos}
    with (
        _logged_out(),
        patch("opendi.main.requests.get", return_value=mock_resp),
    ):
        result = runner.invoke(app, ["search", "x", "--limit", "3"])
    assert result.exit_code == 0
    # Only 3 repos should appear
    assert result.output.count("alice/repo-") == 3
