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
