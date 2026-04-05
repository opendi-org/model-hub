"""Tests for the CLI entry point."""

from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()

_TOKEN = "fake-access-token"


# ── Helpers ───────────────────────────────────────────────────────────────────


def _logged_in():
    """Patch load_access_token to return a valid token."""
    return patch("opendi.main.credential_storage.load_access_token", return_value=_TOKEN)


def _logged_out():
    """Patch load_access_token to return None (not logged in)."""
    return patch("opendi.main.credential_storage.load_access_token", return_value=None)


# ── Help ──────────────────────────────────────────────────────────────────────


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
        patch("opendi.main.credential_storage.delete") as mock_delete,
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "Session expired" in result.output
    mock_delete.assert_called_once()


# ── Logout ────────────────────────────────────────────────────────────────────


def test_logout_success() -> None:
    """opendi logout deletes stored token."""
    with patch("opendi.main.credential_storage.delete", return_value=True):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Logged out" in result.output


def test_logout_not_logged_in() -> None:
    """opendi logout prints 'Not logged in' when no token stored."""
    with patch("opendi.main.credential_storage.delete", return_value=False):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Not logged in" in result.output

# ── Create repo ───────────────────────────────────────────────────────────────


def test_create_repo_requires_login() -> None:
    """opendi create-repo exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_create_repo_success_private() -> None:
    """opendi create-repo posts to the API and prints success (private by default)."""
    response = MagicMock()
    response.status_code = 201
    with (
        _logged_in(),
        patch("opendi.main.requests.post", return_value=response) as mock_post,
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 0
    assert "my-repo" in result.output
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["slug"] == "my-repo"
    assert kwargs["json"]["visibility"] == "private"


def test_create_repo_success_public() -> None:
    """opendi create-repo creates a public repo when --public is passed."""
    response = MagicMock()
    response.status_code = 201
    with (
        _logged_in(),
        patch("opendi.main.requests.post", return_value=response) as mock_post,
    ):
        result = runner.invoke(app, ["create-repo", "my-repo", "--public"])
    assert result.exit_code == 0
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["visibility"] == "public"


def test_create_repo_success_with_description() -> None:
    """opendi create-repo forwards --description to the API."""
    response = MagicMock()
    response.status_code = 201
    with (
        _logged_in(),
        patch("opendi.main.requests.post", return_value=response) as mock_post,
    ):
        result = runner.invoke(app, ["create-repo", "my-repo", "--description", "A test repo"])
    assert result.exit_code == 0
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["description"] == "A test repo"


def test_create_repo_conflict() -> None:
    """opendi create-repo exits 1 when the repo name already exists (HTTP 409)."""
    response = MagicMock()
    response.status_code = 409
    with (_logged_in(), patch("opendi.main.requests.post", return_value=response)):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "already exists" in result.output


def test_create_repo_unauthorized() -> None:
    """opendi create-repo exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (_logged_in(), patch("opendi.main.requests.post", return_value=response)):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_create_repo_unexpected_status() -> None:
    """opendi create-repo exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (_logged_in(), patch("opendi.main.requests.post", return_value=response)):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_create_repo_connection_error() -> None:
    """opendi create-repo exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch("opendi.main.requests.post", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_create_repo_timeout() -> None:
    """opendi create-repo exits 1 with timeout message when the request times out."""
    with (
        _logged_in(),
        patch("opendi.main.requests.post", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── List repos ────────────────────────────────────────────────────────────────


def test_list_repos_requires_login() -> None:
    """opendi list-repos exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_list_repos_all() -> None:
    """opendi list-repos prints all repos when no owner is given."""
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = {
        "repositories": [
            {"slug": "repo-a", "visibility": "public", "description": "First repo"},
            {"slug": "repo-b", "visibility": "private", "description": ""},
        ]
    }
    with (_logged_in(), patch("opendi.main.requests.get", return_value=response)):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 0
    assert "repo-a" in result.output
    assert "repo-b" in result.output


def test_list_repos_by_owner() -> None:
    """opendi list-repos sends owner param when given."""
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = {"repositories": [{"slug": "repo-a", "visibility": "public", "description": ""}]}
    with (
        _logged_in(),
        patch("opendi.main.requests.get", return_value=response) as mock_get,
    ):
        result = runner.invoke(app, ["list-repos", "alice"])
    assert result.exit_code == 0
    _, kwargs = mock_get.call_args
    assert kwargs["params"] == {"owner": "alice"}


def test_list_repos_empty() -> None:
    """opendi list-repos prints 'No repositories found' when API returns empty list."""
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = {"repositories": []}
    with (_logged_in(), patch("opendi.main.requests.get", return_value=response)):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 0
    assert "No repositories found" in result.output


def test_list_repos_owner_not_found() -> None:
    """opendi list-repos exits 1 when the owner does not exist (HTTP 404)."""
    response = MagicMock()
    response.status_code = 404
    with (_logged_in(), patch("opendi.main.requests.get", return_value=response)):
        result = runner.invoke(app, ["list-repos", "alice"])
    assert result.exit_code == 1
    assert "alice" in result.output


def test_list_repos_unauthorized() -> None:
    """opendi list-repos exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (_logged_in(), patch("opendi.main.requests.get", return_value=response)):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_list_repos_unexpected_status() -> None:
    """opendi list-repos exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (_logged_in(), patch("opendi.main.requests.get", return_value=response)):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_list_repos_connection_error() -> None:
    """opendi list-repos exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_list_repos_timeout() -> None:
    """opendi list-repos exits 1 with timeout message when the request times out."""
    with (
        _logged_in(),
        patch("opendi.main.requests.get", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── Delete repo ───────────────────────────────────────────────────────────────


def test_delete_repo_requires_login() -> None:
    """opendi delete-repo exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_repo_invalid_format() -> None:
    """opendi delete-repo exits 1 when repo is not in owner/slug format."""
    with _logged_out():
        result = runner.invoke(app, ["delete-repo", "my-repo", "--yes"])
    assert result.exit_code == 1
    assert "owner/slug" in result.output


def test_delete_repo_success() -> None:
    """opendi delete-repo deletes the repo and prints success (HTTP 204)."""
    response = MagicMock()
    response.status_code = 204
    with (
        _logged_in(),
        patch("opendi.main.requests.delete", return_value=response) as mock_delete,
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output
    args, _ = mock_delete.call_args
    assert "alice/my-repo" in args[0]


def test_delete_repo_prompts_without_yes_flag() -> None:
    """opendi delete-repo prompts for confirmation when --yes is not passed."""
    response = MagicMock()
    response.status_code = 204
    with (
        _logged_in(),
        patch("opendi.main.requests.delete", return_value=response),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo"], input="y\n")
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output


def test_delete_repo_aborts_on_prompt_decline() -> None:
    """opendi delete-repo aborts when the user declines the confirmation prompt."""
    with (
        _logged_in(),
        patch("opendi.main.requests.delete") as mock_delete,
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo"], input="n\n")
    assert result.exit_code != 0
    mock_delete.assert_not_called()


def test_delete_repo_not_found_or_forbidden() -> None:
    """opendi delete-repo exits 1 when repo does not exist or user is not the owner (403/404)."""
    for status in (403, 404):
        response = MagicMock()
        response.status_code = status
        with (_logged_in(), patch("opendi.main.requests.delete", return_value=response)):
            result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
        assert result.exit_code == 1
        assert "does not exist or you are not the owner" in result.output


def test_delete_repo_unauthorized() -> None:
    """opendi delete-repo exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (_logged_in(), patch("opendi.main.requests.delete", return_value=response)):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_repo_unexpected_status() -> None:
    """opendi delete-repo exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (_logged_in(), patch("opendi.main.requests.delete", return_value=response)):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_delete_repo_connection_error() -> None:
    """opendi delete-repo exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch("opendi.main.requests.delete", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_delete_repo_timeout() -> None:
    """opendi delete-repo exits 1 with timeout message when the request times out."""
    with (
        _logged_in(),
        patch("opendi.main.requests.delete", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "timed out" in result.output

# ── Pull ─────────────────────────────────────────────────────────────────────


def test_pull_invalid_format() -> None:
    """opendi pull exits 1 when name is not owner/repo:tag format."""
    result = runner.invoke(app, ["pull", "invalid-name"])
    assert result.exit_code == 1
    assert "Invalid format" in result.output


def test_pull_unauthenticated_public_repo() -> None:
    """opendi pull succeeds without credentials for public repos."""
    mock_response = MagicMock()
    mock_response.ok = True
    mock_response.status_code = 200
    mock_response.text = '{"meta": {"uuid": "abc"}}'
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.get", return_value=mock_response),
        patch("opendi.main.local_store.save_model") as mock_save,
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 0
    assert "Pulled alice/my-model:v1.0 into local cache" in result.output
    mock_save.assert_called_once_with("alice", "my-model", "v1.0", mock_response.text)


def test_pull_not_found() -> None:
    """opendi pull exits 1 when server returns 404."""
    mock_response = MagicMock()
    mock_response.status_code = 404
    mock_response.ok = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.get", return_value=mock_response),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "Not found" in result.output


def test_pull_access_denied() -> None:
    """opendi pull exits 1 when server returns 403."""
    mock_response = MagicMock()
    mock_response.status_code = 403
    mock_response.ok = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.get", return_value=mock_response),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "Access denied" in result.output


def test_pull_connection_error() -> None:
    """opendi pull exits 1 when server is unreachable."""
    import requests as req_lib
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.get", side_effect=req_lib.ConnectionError()),
    ):
        result = runner.invoke(app, ["pull", "alice/my-model:v1.0"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


# ── Push ─────────────────────────────────────────────────────────────────────


def test_push_invalid_name_format() -> None:
    """opendi push exits 1 when --name is not owner/repo:tag format."""
    result = runner.invoke(app, ["push", "model.json", "--name", "bad-name"])
    assert result.exit_code == 1
    assert "Invalid --name format" in result.output


def test_push_file_not_found() -> None:
    """opendi push exits 1 when the local file does not exist."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["push", "/nonexistent/model.json", "--name", "alice/repo:v1.0"])
    assert result.exit_code == 1
    assert "File not found" in result.output


def test_push_success() -> None:
    """opendi push uploads file and prints digest and size."""
    import tempfile, json as _json
    mock_response = MagicMock()
    mock_response.ok = True
    mock_response.status_code = 200
    mock_response.json.return_value = {"digest": "abc123", "size": 42}
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{"meta": {"uuid": "abc"}}')
        tmp_path = f.name
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.put", return_value=mock_response),
    ):
        result = runner.invoke(app, ["push", tmp_path, "--name", "alice/repo:v1.0"])
    assert result.exit_code == 0
    assert "abc123" in result.output
    assert "42" in result.output


def test_push_access_denied() -> None:
    """opendi push exits 1 when server returns 403."""
    import tempfile
    mock_response = MagicMock()
    mock_response.status_code = 403
    mock_response.ok = False
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        f.write(b'{}')
        tmp_path = f.name
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main.requests.put", return_value=mock_response),
    ):
        result = runner.invoke(app, ["push", tmp_path, "--name", "alice/repo:v1.0"])
    assert result.exit_code == 1
    assert "Access denied" in result.output