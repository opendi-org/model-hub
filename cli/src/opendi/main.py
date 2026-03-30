"""CLI entry point. Called when the user runs the `opendi` command."""

import os

import requests
import typer

from opendi import auth, credential_storage, log

_API_BASE_URL = os.environ.get("OPENDI_API_URL", "http://localhost:8080")
# Populated by main() before every command; None if not logged in.
_current_token: str | None = None
_had_creds: bool = False  # True when creds were found but a valid token couldn't be obtained.

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
    global _current_token, _had_creds
    _current_token = None
    _had_creds = False
    log.configure_logging()

    creds = credential_storage.load_creds()
    if creds is None:
        return

    _had_creds = True
    id_token = credential_storage.load_id_token()

    if creds.expired and creds.refresh_token:
        try:
            creds.refresh(Request())
            credential_storage.store_all(creds)
            id_token = creds.id_token
        except Exception:
            return  # _had_creds=True, _current_token=None → _require_token() will handle it

    if id_token:
        _current_token = id_token


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


@app.command()
def search(
    query: str = typer.Argument(..., help="Search query"),
) -> None:
    """Search repositories on the hub. Shows more results when logged in."""
    # Auth is optional: use token if available, otherwise search as unauthenticated.
    headers: dict[str, str] = {}
    if _current_token:
        headers["Authorization"] = f"Bearer {_current_token}"

    try:
        response = requests.get(
            f"{_API_BASE_URL}/v0/repositories",
            params={"q": query},
            headers=headers,
            timeout=10,
        )
        if response.status_code == 200:
            repos = response.json().get("repositories", [])
            if not repos:
                typer.echo("No repositories found.")
                return
            for repo in repos:
                visibility = typer.style(repo.get("visibility", ""), fg=typer.colors.YELLOW)
                name = typer.style(repo.get("slug", ""), fg=typer.colors.GREEN, bold=True)
                description = repo.get("description", "")
                line = f"{name} [{visibility}]"
                if description:
                    line += f"  {description}"
                typer.echo(line)
        else:
            typer.echo(f"Search failed (HTTP {response.status_code}).", err=True)
            raise typer.Exit(1)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to the hub at {_API_BASE_URL}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


@app.command()
def list_repos(
    owner: str = typer.Argument(None, help="Owner username (defaults to all repositories)"),
) -> None:
    """List repositories on the hub."""
    jwt = _require_token()
    params = {"owner": owner} if owner else {}
    try:
        response = requests.get(
            f"{_API_BASE_URL}/v0/repositories",
            params=params,
            headers={"Authorization": f"Bearer {jwt}"},
            timeout=10,
        )
        if response.status_code == 200:
            repos = response.json().get("repositories", [])
            if not repos:
                typer.echo("No repositories found.")
                return
            for repo in repos:
                visibility = typer.style(repo.get("visibility", ""), fg=typer.colors.YELLOW)
                name = typer.style(repo.get("slug", ""), fg=typer.colors.GREEN, bold=True)
                description = repo.get("description", "")
                line = f"{name} [{visibility}]"
                if description:
                    line += f"  {description}"
                typer.echo(line)
        elif response.status_code == 404:
            typer.echo(f"Owner '{owner}' not found.", err=True)
            raise typer.Exit(1)
        elif response.status_code == 401:
            typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
            raise typer.Exit(1)
        else:
            typer.echo(f"Failed to list repositories (HTTP {response.status_code}).", err=True)
            raise typer.Exit(1)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to the hub at {_API_BASE_URL}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


@app.command()
def create_repo(
    name: str = typer.Argument(..., help="Repository name (slug)"),
    description: str = typer.Option("", "--description", "-d", help="Short description"),
    private: bool = typer.Option(True, "--private/--public", help="Visibility"),
) -> None:
    """Create a new repository on the hub."""
    jwt = _require_token()
    visibility = "private" if private else "public"
    try:
        response = requests.post(
            f"{_API_BASE_URL}/v0/repositories",
            json={"slug": name, "description": description, "visibility": visibility},
            headers={"Authorization": f"Bearer {jwt}"},
            timeout=10,
        )
        if response.status_code == 201:
            typer.echo(f"Repository {typer.style(name, fg=typer.colors.GREEN, bold=True)} created.")
        elif response.status_code == 409:
            typer.echo(f"A repository named '{name}' already exists.", err=True)
            raise typer.Exit(1)
        elif response.status_code == 401:
            typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
            raise typer.Exit(1)
        else:
            typer.echo(f"Failed to create repository (HTTP {response.status_code}).", err=True)
            raise typer.Exit(1)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to the hub at {_API_BASE_URL}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


@app.command()
def delete_repo(
    repo: str = typer.Argument(..., help="Repository to delete in owner/slug format (e.g. alice/my-repo)"),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation prompt"),
) -> None:
    """Permanently delete a repository and all associated data."""
    if "/" not in repo or repo.count("/") != 1:
        typer.echo("Repository must be in owner/slug format (e.g. alice/my-repo).", err=True)
        raise typer.Exit(1)
    owner, slug = repo.split("/", 1)
    jwt = _require_token()
    if not yes:
        typer.confirm(f"Delete repository '{repo}'? This cannot be undone.", abort=True)
    try:
        response = requests.delete(
            f"{_API_BASE_URL}/v0/repositories/{owner}/{slug}",
            headers={"Authorization": f"Bearer {jwt}"},
            timeout=10,
        )
        if response.status_code == 204:
            typer.echo(f"Repository {typer.style(repo, fg=typer.colors.GREEN, bold=True)} deleted.")
        elif response.status_code in (403, 404):
            typer.echo(
                f"Repository '{repo}' does not exist or you are not the owner.", err=True
            )
            raise typer.Exit(1)
        elif response.status_code == 401:
            typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
            raise typer.Exit(1)
        else:
            typer.echo(f"Failed to delete repository (HTTP {response.status_code}).", err=True)
            raise typer.Exit(1)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to the hub at {_API_BASE_URL}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


if __name__ == "__main__":
    app()
