"""CLI entry point. Called when the user runs the `opendi` command."""

import os

import typer

from opendi import auth, credential_storage, log

app = typer.Typer(
    name="opendi",
    no_args_is_help=True,
)


def _api_base_url() -> str:
    return os.environ.get("OPENDI_API_URL", "http://localhost:8080").rstrip("/")


def _require_access_token() -> str:
    token = credential_storage.load_access_token()
    if not token:
        typer.echo("Not logged in. Please run `opendi login` first.", err=True)
        raise typer.Exit(1)
    return token


@app.callback(invoke_without_command=True)
def main(_ctx: typer.Context) -> None:
    """[bold]OpenDI Model Hub CLI[/bold] — cross-platform client for the OpenDI hub."""
    log.configure_logging()


@app.command()
def login() -> None:
    """Log in to the OpenDI hub (browser-assisted device flow)."""
    api_base = _api_base_url()
    try:
        token = credential_storage.load_access_token()
        if token:
            try:
                me = auth.get_current_user(api_base, token)
                username = me.get("username")
                if username:
                    typer.echo(f"Already logged in as {typer.style(username, fg=typer.colors.GREEN, bold=True)}.")
                else:
                    typer.echo("Already logged in.")
                return
            except Exception:
                credential_storage.delete()

        code, login_url, expires_in = auth.start_cli_login(api_base)
        absolute_url = auth.open_login_url(api_base, login_url)
        typer.echo(f"Approve login in your browser:\n{absolute_url}")
        token = auth.poll_cli_token(api_base, code, expires_in)
        credential_storage.store_access_token(token)
        me = auth.get_current_user(api_base, token)
        username = me.get("username")
        if username:
            typer.echo(f"Login successful. Logged in as {typer.style(username, fg=typer.colors.GREEN, bold=True)}.")
        else:
            typer.echo("Login successful.")
    except KeyboardInterrupt:
        typer.echo("Login cancelled.", err=True)
        raise typer.Exit(1)
    except TimeoutError:
        typer.echo("Login timed out. Please try again.", err=True)
        raise typer.Exit(1)
    except Exception as e:
        typer.echo(f"Login failed: {e}", err=True)
        raise typer.Exit(1)


@app.command()
def whoami() -> None:
    """Show the currently logged-in account."""
    api_base = _api_base_url()
    token = _require_access_token()
    try:
        me = auth.get_current_user(api_base, token)
    except Exception:
        credential_storage.delete()
        typer.echo("Session expired. Run `opendi login` to sign in again.", err=True)
        raise typer.Exit(1)
    username = me.get("username")
    email = me.get("email")
    if username:
        typer.echo(typer.style(username, fg=typer.colors.GREEN, bold=True))
    elif email:
        typer.echo(typer.style(email, fg=typer.colors.GREEN, bold=True))
    else:
        typer.echo("Logged in.")


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
    _require_access_token()
    typer.echo(f"Pull not yet implemented for: {name}")


@app.command()
def push(
    path: str = typer.Argument(..., help="Local path to push"),
) -> None:
    """Push a model or resource to the hub."""
    _require_access_token()
    typer.echo(f"Push not yet implemented for: {path}")


if __name__ == "__main__":
    app()
