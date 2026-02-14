"""CLI entry point. Called when the user runs the `opendi` command."""

import base64
import json
import os
from pathlib import Path

import typer
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials as GoogleCredentials
from oauthlib.oauth2.rfc6749.errors import AccessDeniedError

from opendi import auth, credential_storage, log

app = typer.Typer(
    name="opendi",
    no_args_is_help=True,
)


def _resolve_oauth_client_secrets() -> Path:
    """Resolve path to OAuth client secrets file (bundled or env override).

    Returns:
        Path to a GCP client_secret.json file.

    Raises:
        FileNotFoundError: If the env override path or bundled file is missing.
    """
    # Override: env var points to a GCP client_secret.json file path
    env_path = os.environ.get("OPENDI_OAUTH_CLIENT_SECRETS")
    if env_path:
        path = Path(env_path)
        if not path.is_file():
            raise FileNotFoundError(
                f"OPENDI_OAUTH_CLIENT_SECRETS points to missing file: {env_path}"
            )
        return path

    # Default: bundled file next to this module
    bundled = Path(__file__).resolve().parent / "client_secret.json"
    if not bundled.is_file():
        raise FileNotFoundError(
            "OAuth client secrets not found. "
            "Contributors: set OPENDI_OAUTH_CLIENT_SECRETS to your GCP file path, "
            "or place client_secret.json in cli/src/opendi/. See README."
        )
    return bundled


def _require_credentials() -> str:
    """Load stored OAuth credentials and id_token; refresh if expired. Return the JWT.

    Exit with a message if not logged in, refresh fails, or no id_token is available.
    On refresh failure or missing id_token, stored state is cleared so the user
    can run `opendi login` again cleanly.
    """
    creds = credential_storage.load_creds()
    if creds is None:
        typer.echo("Not logged in.", err=True)
        raise typer.Exit(1)

    id_token = credential_storage.load_id_token()

    if creds.expired and creds.refresh_token:
        try:
            creds.refresh(Request())
            credential_storage.store_all(creds)
            id_token = creds.id_token
        except Exception:
            credential_storage.delete()
            typer.echo("Session expired. Run `opendi login` to sign in again.", err=True)
            raise typer.Exit(1)

    if not id_token:
        credential_storage.delete()
        typer.echo("Session expired. Run `opendi login` to sign in again.", err=True)
        raise typer.Exit(1)

    return id_token


def _get_email_from_id_token(id_token: str | None) -> str | None:
    """Decode email from a Google id_token (JWT) without verification.

    Only for display purposes — no signature check needed.
    """
    if not id_token:
        return None
    try:
        payload_b64 = id_token.split(".")[1]
        # Fix base64 padding.
        payload_b64 += "=" * (-len(payload_b64) % 4)
        payload = json.loads(base64.urlsafe_b64decode(payload_b64))
        return payload.get("email")
    except Exception:
        return None


def _styled_email(email: str) -> str:
    """Return email styled in bold green for terminal display."""
    return typer.style(email, fg=typer.colors.GREEN, bold=True)


@app.callback(invoke_without_command=True)
def main(_ctx: typer.Context) -> None:
    """[bold]OpenDI Model Hub CLI[/bold] — cross-platform client for the OpenDI hub."""
    log.configure_logging()


@app.command()
def login() -> None:
    """Log in to the OpenDI hub (Google OAuth, local loopback)."""
    try:
        creds: GoogleCredentials | None = credential_storage.load_creds()
        id_token = credential_storage.load_id_token()

        if creds is not None and id_token is not None:
            email = _get_email_from_id_token(id_token)
            if email:
                typer.echo(f"Already logged in as {_styled_email(email)}.")
            else:
                typer.echo("Already logged in.")
            return

        # Full OAuth flow (opens browser). Refresh/errors are handled by other commands via _require_credentials().
        client_secrets_path = _resolve_oauth_client_secrets()
        creds = auth.run_login_flow(client_secrets_path)
        credential_storage.store_all(creds)
        email = _get_email_from_id_token(creds.id_token)
        if email:
            typer.echo(f"Login successful. Logged in as {_styled_email(email)}.")
        else:
            typer.echo("Login successful.")
    except (KeyboardInterrupt, AccessDeniedError):
        typer.echo("Login cancelled.", err=True)
        raise typer.Exit(1)
    except FileNotFoundError as e:
        # Setup problem: missing client secrets; message is actionable (path, README).
        typer.echo(str(e), err=True)
        raise typer.Exit(1)
    except AttributeError:
        # No redirect received (timeout or closed tab); explain and suggest retry.
        typer.echo("Login timed out or cancelled. Please try again.", err=True)
        raise typer.Exit(1)
    except Exception:
        # Other OAuth or unexpected errors; generic retry.
        typer.echo("Login failed. Please try again.", err=True)
        raise typer.Exit(1)


@app.command()
def whoami() -> None:
    """Show the email for the currently logged-in account."""
    jwt = _require_credentials()
    email = _get_email_from_id_token(jwt)
    if email:
        typer.echo(_styled_email(email))
    else:
        typer.echo("Logged in (could not determine email).")


@app.command()
def logout() -> None:
    """Log out from the OpenDI hub (clears stored credentials)."""
    if credential_storage.delete():
        typer.echo("Logged out.")
    else:
        typer.echo("Not logged in.")


@app.command()
def pull(
    name: str = typer.Argument(..., help="Model or resource name to pull"),
) -> None:
    """Pull a model or resource from the hub."""
    _require_credentials()
    typer.echo(f"Pull not yet implemented for: {name}")


@app.command()
def push(
    path: str = typer.Argument(..., help="Local path to push"),
) -> None:
    """Push a model or resource to the hub."""
    typer.echo(f"Push not yet implemented for: {path}")


if __name__ == "__main__":
    app()
