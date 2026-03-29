"""Tests for the CLI entry point."""

from pathlib import Path
from unittest.mock import MagicMock, patch

from oauthlib.oauth2.rfc6749.errors import AccessDeniedError, OAuth2Error
from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()

# A real-ish JWT whose payload decodes to {"email": "user@example.com"}.
_JWT_WITH_EMAIL = "header.eyJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20ifQ.sig"

# A real-ish JWT whose payload has no email claim: {"sub": "12345"}.
_JWT_NO_EMAIL = "header.eyJzdWIiOiIxMjM0NSJ9.sig"


# ── Help ──────────────────────────────────────────────────────────────────────


def test_app_help_exits_zero() -> None:
    """opendi --help exits with code 0."""
    result = runner.invoke(app, ["--help"])
    assert result.exit_code == 0
    assert "opendi" in result.output.lower()


def test_app_without_command_shows_help() -> None:
    """opendi (no subcommand) shows full help, same as --help."""
    result = runner.invoke(app, [])
    assert "Usage:" in result.output
    assert "login" in result.output
    assert "pull" in result.output
    assert result.exit_code == 2


# ── Login: full OAuth flow ────────────────────────────────────────────────────


def test_login_success() -> None:
    """opendi login stores credentials and prints success with email."""
    creds = MagicMock()
    creds.id_token = _JWT_WITH_EMAIL
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", return_value=creds) as mock_flow,
        patch("opendi.main.credential_storage.store_all") as mock_store,
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Login successful" in result.output
    assert "user@example.com" in result.output
    mock_flow.assert_called_once_with(fake_path)
    mock_store.assert_called_once_with(creds)


def test_login_success_no_email_fallback() -> None:
    """opendi login prints success without email when id_token has no email."""
    creds = MagicMock()
    creds.id_token = None
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", return_value=creds),
        patch("opendi.main.credential_storage.store_all"),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Login successful" in result.output
    assert "Logged in as" not in result.output


def test_login_file_not_found() -> None:
    """opendi login exits 1 when client secrets are missing."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets") as mock_resolve,
    ):
        mock_resolve.side_effect = FileNotFoundError("OAuth client secrets not found.")
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "OAuth client secrets not found" in result.output


def test_login_access_denied() -> None:
    """opendi login exits 1 with 'Login cancelled' when user denies consent."""
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", side_effect=AccessDeniedError()),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "Login cancelled" in result.output


def test_login_oauth_error() -> None:
    """opendi login exits 1 with generic message on other OAuth errors."""
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", side_effect=OAuth2Error()),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "Login failed. Please try again" in result.output


def test_login_timeout_or_no_redirect() -> None:
    """opendi login exits 1 with timeout message when no redirect."""
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", side_effect=AttributeError()),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 1
    assert "timed out or cancelled" in result.output


# ── Login: existing credentials ───────────────────────────────────────────────


def test_login_already_logged_in() -> None:
    """opendi login prints 'Already logged in' with email from stored id_token."""
    creds = MagicMock()
    creds.valid = True
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Already logged in" in result.output
    assert "user@example.com" in result.output


def test_login_creds_without_id_token_does_full_oauth() -> None:
    """opendi login does full OAuth when creds exist but id_token is missing."""
    old_creds = MagicMock()
    old_creds.expired = False
    new_creds = MagicMock()
    new_creds.id_token = _JWT_WITH_EMAIL
    fake_path = Path("/fake/client_secret.json")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=old_creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=None),
        patch("opendi.main._resolve_oauth_client_secrets", return_value=fake_path),
        patch("opendi.main.auth.run_login_flow", return_value=new_creds) as mock_flow,
        patch("opendi.main.credential_storage.store_all") as mock_store,
    ):
        result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "Login successful" in result.output
    assert "user@example.com" in result.output
    mock_flow.assert_called_once_with(fake_path)
    mock_store.assert_called_once_with(new_creds)


# ── Whoami ────────────────────────────────────────────────────────────────────


def test_whoami_not_logged_in() -> None:
    """opendi whoami exits 1 when no credentials stored (via _require_credentials)."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_whoami_shows_email() -> None:
    """opendi whoami prints the logged-in user's email from id_token."""
    creds = MagicMock()
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 0
    assert "user@example.com" in result.output


def test_whoami_refreshes_expired_then_shows_email() -> None:
    """opendi whoami refreshes an expired token and prints email."""
    creds = MagicMock()
    creds.expired = True
    creds.refresh_token = "refresh_tok"
    creds.id_token = _JWT_WITH_EMAIL  # id_token after refresh (in memory)
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value="old.jwt.token"),
        patch("opendi.main.credential_storage.store_all"),
        patch("opendi.main.Request"),
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 0
    assert "user@example.com" in result.output
    creds.refresh.assert_called_once()


def test_whoami_expired_refresh_fails_deletes_storage() -> None:
    """opendi whoami exits 1 and clears stored state when token is expired and refresh fails."""
    creds = MagicMock()
    creds.expired = True
    creds.refresh_token = "refresh_tok"
    creds.refresh.side_effect = Exception("refresh failed")
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.credential_storage.delete") as mock_delete,
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "Session expired" in result.output
    mock_delete.assert_called_once()


def test_whoami_no_email_fallback() -> None:
    """opendi whoami prints fallback when id_token has no email claim."""
    creds = MagicMock()
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_NO_EMAIL),
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 0
    assert "could not determine email" in result.output


def test_whoami_no_id_token_exits_and_deletes_storage() -> None:
    """opendi whoami exits 1 and clears stored state when no id_token is stored."""
    creds = MagicMock()
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=None),
        patch("opendi.main.credential_storage.delete") as mock_delete,
    ):
        result = runner.invoke(app, ["whoami"])
    assert result.exit_code == 1
    assert "opendi login" in result.output
    mock_delete.assert_called_once()


# ── Logout ────────────────────────────────────────────────────────────────────


def test_logout_success() -> None:
    """opendi logout deletes stored credentials."""
    with patch("opendi.main.credential_storage.delete", return_value=True):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Logged out" in result.output


def test_logout_not_logged_in() -> None:
    """opendi logout prints 'Not logged in' when no credentials stored."""
    with patch("opendi.main.credential_storage.delete", return_value=False):
        result = runner.invoke(app, ["logout"])
    assert result.exit_code == 0
    assert "Not logged in" in result.output


# ── Pull / Push: require login ───────────────────────────────────────────────


def test_pull_requires_login() -> None:
    """opendi pull exits 1 with message when not logged in."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["pull", "some-model"])
    assert result.exit_code == 1
    assert "Please run `opendi login` first" in result.output


def test_pull_with_credentials() -> None:
    """opendi pull runs (mock) when logged in."""
    creds = MagicMock()
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
    ):
        result = runner.invoke(app, ["pull", "some-model"])
    assert result.exit_code == 0
    assert "Pull not yet implemented for: some-model" in result.output


def test_pull_refreshes_expired_token() -> None:
    """opendi pull refreshes expired credentials and then runs."""
    creds = MagicMock()
    creds.expired = True
    creds.refresh_token = "refresh_tok"
    creds.id_token = _JWT_WITH_EMAIL  # id_token after refresh
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value="old.jwt.token"),
        patch("opendi.main.credential_storage.store_all") as mock_store,
        patch("opendi.main.Request"),
    ):
        result = runner.invoke(app, ["pull", "some-model"])
    assert result.exit_code == 0
    assert "Pull not yet implemented for: some-model" in result.output
    creds.refresh.assert_called_once()
    mock_store.assert_called_once_with(creds)


def test_push_requires_login() -> None:
    """opendi push exits 1 with message when not logged in."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["push", "/path/to/model"])
    assert result.exit_code == 1
    assert "Please run `opendi login` first" in result.output


def test_push_with_credentials() -> None:
    """opendi push runs (mock) when logged in."""
    creds = MagicMock()
    creds.expired = False
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=creds),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
    ):
        result = runner.invoke(app, ["push", "/path/to/model"])
    assert result.exit_code == 0
    assert "Push not yet implemented for: /path/to/model" in result.output


# ── Create repo ───────────────────────────────────────────────────────────────


def _logged_in_creds() -> MagicMock:
    creds = MagicMock()
    creds.expired = False
    return creds


def test_create_repo_requires_login() -> None:
    """opendi create-repo exits 1 with message when not logged in."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "Please run `opendi login` first" in result.output


def test_create_repo_success_private() -> None:
    """opendi create-repo posts to the API and prints success (private by default)."""
    response = MagicMock()
    response.status_code = 201
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", return_value=response) as mock_post,
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 0
    assert "my-repo" in result.output
    mock_post.assert_called_once()
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["slug"] == "my-repo"
    assert kwargs["json"]["visibility"] == "private"


def test_create_repo_success_public() -> None:
    """opendi create-repo creates a public repo when --public is passed."""
    response = MagicMock()
    response.status_code = 201
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
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
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
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
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", return_value=response),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "already exists" in result.output


def test_create_repo_unauthorized() -> None:
    """opendi create-repo exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", return_value=response),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_create_repo_unexpected_status() -> None:
    """opendi create-repo exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", return_value=response),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_create_repo_connection_error() -> None:
    """opendi create-repo exits 1 when the hub is unreachable."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_create_repo_timeout() -> None:
    """opendi create-repo exits 1 with timeout message when the request times out."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.post", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["create-repo", "my-repo"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── List repos ────────────────────────────────────────────────────────────────


def test_list_repos_requires_login() -> None:
    """opendi list-repos exits 1 with message when not logged in."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "Please run `opendi login` first" in result.output


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
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", return_value=response) as mock_get,
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 0
    assert "repo-a" in result.output
    assert "repo-b" in result.output
    _, kwargs = mock_get.call_args
    assert "owner" not in kwargs.get("params", {})


def test_list_repos_by_owner() -> None:
    """opendi list-repos filters by owner when an owner argument is given."""
    response = MagicMock()
    response.status_code = 200
    response.json.return_value = {"repositories": [{"slug": "repo-a", "visibility": "public", "description": ""}]}
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
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
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", return_value=response),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 0
    assert "No repositories found" in result.output


def test_list_repos_owner_not_found() -> None:
    """opendi list-repos exits 1 when the owner does not exist (HTTP 404)."""
    response = MagicMock()
    response.status_code = 404
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", return_value=response),
    ):
        result = runner.invoke(app, ["list-repos", "alice"])
    assert result.exit_code == 1
    assert "alice" in result.output


def test_list_repos_unauthorized() -> None:
    """opendi list-repos exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", return_value=response),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_list_repos_unexpected_status() -> None:
    """opendi list-repos exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", return_value=response),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_list_repos_connection_error() -> None:
    """opendi list-repos exits 1 when the hub is unreachable."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_list_repos_timeout() -> None:
    """opendi list-repos exits 1 with timeout message when the request times out."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.get", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["list-repos"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── Delete repo ───────────────────────────────────────────────────────────────


def test_delete_repo_requires_login() -> None:
    """opendi delete-repo exits 1 with message when not logged in."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "Please run `opendi login` first" in result.output


def test_delete_repo_invalid_format() -> None:
    """opendi delete-repo exits 1 when repo is not in owner/slug format."""
    with patch("opendi.main.credential_storage.load_creds", return_value=None):
        result = runner.invoke(app, ["delete-repo", "my-repo", "--yes"])
    assert result.exit_code == 1
    assert "owner/slug" in result.output


def test_delete_repo_success() -> None:
    """opendi delete-repo deletes the repo and prints success (HTTP 204)."""
    response = MagicMock()
    response.status_code = 204
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", return_value=response) as mock_delete,
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output
    mock_delete.assert_called_once()
    args, _ = mock_delete.call_args
    assert "alice/my-repo" in args[0]


def test_delete_repo_prompts_without_yes_flag() -> None:
    """opendi delete-repo prompts for confirmation when --yes is not passed."""
    response = MagicMock()
    response.status_code = 204
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", return_value=response),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo"], input="y\n")
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output


def test_delete_repo_aborts_on_prompt_decline() -> None:
    """opendi delete-repo aborts when the user declines the confirmation prompt."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
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
        with (
            patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
            patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
            patch("opendi.main.requests.delete", return_value=response),
        ):
            result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
        assert result.exit_code == 1
        assert "does not exist or you are not the owner" in result.output


def test_delete_repo_unauthorized() -> None:
    """opendi delete-repo exits 1 with auth message on HTTP 401."""
    response = MagicMock()
    response.status_code = 401
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", return_value=response),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_repo_unexpected_status() -> None:
    """opendi delete-repo exits 1 with HTTP status on unexpected response."""
    response = MagicMock()
    response.status_code = 500
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", return_value=response),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_delete_repo_connection_error() -> None:
    """opendi delete-repo exits 1 when the hub is unreachable."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", side_effect=__import__("requests").ConnectionError()),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_delete_repo_timeout() -> None:
    """opendi delete-repo exits 1 with timeout message when the request times out."""
    with (
        patch("opendi.main.credential_storage.load_creds", return_value=_logged_in_creds()),
        patch("opendi.main.credential_storage.load_id_token", return_value=_JWT_WITH_EMAIL),
        patch("opendi.main.requests.delete", side_effect=__import__("requests").Timeout()),
    ):
        result = runner.invoke(app, ["delete-repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "timed out" in result.output
