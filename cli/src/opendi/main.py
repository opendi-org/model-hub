"""CLI entry point. Called when the user runs the `opendi` command."""

import difflib
import json
import os
import urllib.parse
from pathlib import Path

import requests
import typer

from opendi import auth, credential_storage, local_cache, log

# Populated by main() before every command; None if not logged in.
_current_token: str | None = None

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
    global _current_token
    _current_token = None
    log.configure_logging()

    _current_token = credential_storage.load_access_token()


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
                credential_storage.delete_access_token()

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
        credential_storage.delete_access_token()
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
    if credential_storage.delete_access_token():
        typer.echo("Logged out.")
    else:
        typer.echo("Not logged in.")


@app.command()
def diff(
    left: str = typer.Argument(
        ...,
        help="Remote owner/slug:tag or path to a local CDM JSON file",
    ),
    right: str = typer.Argument(
        ...,
        help="Remote owner/slug:tag or path to a local CDM JSON file",
    ),
) -> None:
    """Compare two CDM models (two remote tags and/or a local JSON file)."""
    api_base = _api_base_url()

    def load_side(spec: str) -> dict:
        s = spec.strip()
        if "/" in s and ":" in s:
            owner, rest = s.split("/", 1)
            if ":" in rest:
                slug, tag = rest.split(":", 1)
                owner, slug, tag = owner.strip(), slug.strip(), tag.strip()
                if owner and slug and tag:
                    owner = owner.lower()
                    o = urllib.parse.quote(owner, safe="")
                    sl = urllib.parse.quote(slug, safe="")
                    t = urllib.parse.quote(tag, safe="")
                    url = f"{api_base}/v0/repositories/{o}/{sl}/tags/{t}/model"
                    headers: dict[str, str] = {"Accept": "application/json"}
                    if _current_token:
                        headers["Authorization"] = f"Bearer {_current_token}"
                    try:
                        response = requests.get(url, headers=headers, timeout=60)
                    except requests.ConnectionError:
                        typer.echo(f"Could not connect to the hub at {api_base}.", err=True)
                        raise typer.Exit(1)
                    except requests.Timeout:
                        typer.echo("Request timed out. Please try again.", err=True)
                        raise typer.Exit(1)
                    if response.status_code == 200:
                        return response.json()
                    err = ""
                    try:
                        body = response.json()
                        if isinstance(body, dict):
                            err = str(body.get("error") or "")
                    except (ValueError, TypeError):
                        err = response.text or ""
                    el = err.lower()
                    if response.status_code == 401:
                        typer.echo(
                            "Not authorised. Run `opendi login` to sign in again.",
                            err=True,
                        )
                        raise typer.Exit(1)
                    if response.status_code == 404:
                        if "tag not found" in el:
                            typer.echo(f"Tag not found: {owner}/{slug}:{tag}", err=True)
                        elif "repository not found" in el:
                            typer.echo(f"Repository not found: {owner}/{slug}", err=True)
                        else:
                            typer.echo("Repository not found or access denied.", err=True)
                        raise typer.Exit(1)
                    if response.status_code == 403:
                        typer.echo("Not authorised to read this model.", err=True)
                        raise typer.Exit(1)
                    typer.echo(
                        f"Failed to load model (HTTP {response.status_code}).",
                        err=True,
                    )
                    raise typer.Exit(1)

        path = Path(s).expanduser()
        if not path.is_file():
            typer.echo(
                f"Expected owner/slug:tag or an existing JSON file, got: {spec}",
                err=True,
            )
            raise typer.Exit(1)
        try:
            return json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as e:
            typer.echo(f"Invalid JSON in file: {e}", err=True)
            raise typer.Exit(1)

    left_obj = load_side(left)
    right_obj = load_side(right)

    if json.dumps(left_obj, sort_keys=True) == json.dumps(right_obj, sort_keys=True):
        typer.echo("No differences (normalized JSON is identical).")
        return

    a = json.dumps(left_obj, sort_keys=True, indent=2, ensure_ascii=False) + "\n"
    b = json.dumps(right_obj, sort_keys=True, indent=2, ensure_ascii=False) + "\n"
    out = "".join(
        difflib.unified_diff(
            a.splitlines(keepends=True),
            b.splitlines(keepends=True),
            fromfile=left,
            tofile=right,
            lineterm="",
        )
    )
    typer.echo(out)


@app.command()
def inspect(
    name: str = typer.Argument(..., help="Model ref to inspect: owner/repo:tag"),
) -> None:
    """Show metadata for a specific model tag (digest, size, timestamps, creator).

    NAME format: owner/repo:tag  (e.g. alice/my-model:v1.0)
    """
    if "/" not in name or ":" not in name:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    repo_part, tag = name.rsplit(":", 1)
    owner, repo_slug = repo_part.split("/", 1)

    if not owner or not repo_slug or not tag:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    api_base = _api_base_url()
    headers: dict[str, str] = {}
    token = credential_storage.load_access_token()
    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        response = requests.get(
            f"{api_base}/v0/repositories/{owner}/{repo_slug}",
            headers=headers,
            timeout=10,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}. Is the server running?", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 401 or response.status_code == 403:
        typer.echo("Access denied. Run `opendi login` if this is a private repository.", err=True)
        raise typer.Exit(1)
    if response.status_code == 404:
        typer.echo(f"Repository {owner}/{repo_slug} not found.", err=True)
        raise typer.Exit(1)
    if not response.ok:
        typer.echo(f"Server error {response.status_code}.", err=True)
        raise typer.Exit(1)

    repo = response.json()
    tags = repo.get("tags", [])
    tag_info = next((t for t in tags if t["name"] == tag), None)

    if tag_info is None:
        typer.echo(f"Tag '{tag}' not found in {owner}/{repo_slug}.", err=True)
        raise typer.Exit(1)

    typer.echo(f"Repository:  {owner}/{repo_slug}")
    typer.echo(f"Tag:         {typer.style(tag, fg=typer.colors.GREEN, bold=True)}")
    typer.echo(f"Digest:      {tag_info.get('digest', 'unknown')}")
    size = tag_info.get("size", 0)
    typer.echo(f"Size:        {size:,} bytes")
    typer.echo(f"Updated:     {tag_info.get('updatedAt', 'unknown')}")
    typer.echo(f"Created by:  {tag_info.get('createdBy', 'unknown')}")


@app.command()
def pull(
    name: str = typer.Argument(..., help="Model ref to pull: owner/repo:tag"),
) -> None:
    """Pull a model from the hub into the local cache.

    NAME format: owner/repo:tag  (e.g. alice/my-model:v1.0)

    Use `opendi save` to write a working copy to disk.
    """
    # Parse owner/repo:tag
    if "/" not in name or ":" not in name:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    repo_part, tag = name.rsplit(":", 1)
    owner, repo_slug = repo_part.split("/", 1)

    if not owner or not repo_slug or not tag:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    # Try authenticated download first; fall back to unauthenticated for public repos.
    headers: dict[str, str] = {}
    token = credential_storage.load_access_token()
    if token:
        headers["Authorization"] = f"Bearer {token}"

    url = f"{_api_base_url()}/v0/repositories/{owner}/{repo_slug}/tags/{tag}/model"
    try:
        response = requests.get(url, headers=headers, timeout=30)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {_api_base_url()}. Is the server running?", err=True)
        raise typer.Exit(1)

    if response.status_code == 404:
        typer.echo(f"Not found: {name}", err=True)
        raise typer.Exit(1)
    if response.status_code == 403:
        typer.echo("Access denied. Run `opendi login` if this is a private repository.", err=True)
        raise typer.Exit(1)
    if not response.ok:
        typer.echo(f"Server error {response.status_code}: {response.text}", err=True)
        raise typer.Exit(1)

    digest = response.headers.get("ETag") or None

    # Replace old cache entries that have the same digest (repo was renamed)
    if digest:
        old_entries = [
            e for e in local_cache.find_by_digest(digest)
            if not (e["owner"] == owner and e["repo"] == repo_slug and e["tag"] == tag)
        ]
        for e in old_entries:
            local_cache.remove_model(e["owner"], e["repo"], e["tag"])
            old_ref = f"{e['owner']}/{e['repo']}:{e['tag']}"
            typer.echo(f"Replaced cached entry {old_ref} → {name}")

    local_cache.save_model(owner, repo_slug, tag, response.text, digest)
    typer.echo(f"Pulled {name} into local cache.")


@app.command()
def push(
    path: str = typer.Argument(..., help="Local CDM JSON file to push"),
    name: str = typer.Option(..., "--name", "-n", help="Target ref: owner/repo:tag"),
) -> None:
    """Push a local CDM JSON file to the hub.

    Example:
      opendi push model.json --name alice/my-repo:v1.0
    """
    # Parse owner/repo:tag
    if "/" not in name or ":" not in name:
        typer.echo("Invalid --name format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    repo_part, tag = name.rsplit(":", 1)
    owner, repo_slug = repo_part.split("/", 1)

    if not owner or not repo_slug or not tag:
        typer.echo("Invalid --name format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    # Read local file
    file_path = Path(path)
    if not file_path.is_file():
        typer.echo(f"File not found: {path}", err=True)
        raise typer.Exit(1)

    raw = file_path.read_bytes()

    # Auth header (optional — works without auth in dev mode for public repos)
    headers: dict[str, str] = {"Content-Type": "application/json"}
    token = credential_storage.load_access_token()
    if token:
        headers["Authorization"] = f"Bearer {token}"

    url = f"{_api_base_url()}/v0/repositories/{owner}/{repo_slug}/tags/{tag}"
    try:
        response = requests.put(url, data=raw, headers=headers, timeout=60)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {_api_base_url()}. Is the server running?", err=True)
        raise typer.Exit(1)

    if response.status_code == 400:
        typer.echo(f"Validation error: {response.json().get('error', response.text)}", err=True)
        raise typer.Exit(1)
    if response.status_code == 403:
        typer.echo("Access denied. You need write access to this repository.", err=True)
        raise typer.Exit(1)
    if response.status_code == 404:
        create = typer.confirm(
            f"Repository {owner}/{repo_slug} not found. Create it as a public repository?",
            default=False,
        )
        if not create:
            typer.echo("Aborted.", err=True)
            raise typer.Exit(1)

        create_url = f"{_api_base_url()}/v0/repositories"
        try:
            create_resp = requests.post(
                create_url,
                json={"slug": repo_slug, "visibility": "public"},
                headers=headers,
                timeout=30,
            )
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {_api_base_url()}. Is the server running?", err=True)
            raise typer.Exit(1)

        if not create_resp.ok:
            typer.echo(
                f"Failed to create repository: {create_resp.status_code} {create_resp.text}",
                err=True,
            )
            raise typer.Exit(1)

        typer.echo(f"Created repository {owner}/{repo_slug}.")

        # Retry the push now that the repo exists
        try:
            response = requests.put(url, data=raw, headers=headers, timeout=60)
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {_api_base_url()}. Is the server running?", err=True)
            raise typer.Exit(1)

        if not response.ok:
            typer.echo(f"Server error {response.status_code}: {response.text}", err=True)
            raise typer.Exit(1)

    if not response.ok:
        typer.echo(f"Server error {response.status_code}: {response.text}", err=True)
        raise typer.Exit(1)

    typer.echo(f"Pushed {file_path.name} → {name}")
    try:
        result = response.json()
        digest = result.get("digest", "unknown")
        size = result.get("size", "unknown")
        typer.echo(f"  digest: {digest}")
        typer.echo(f"  size:   {size} bytes")
    except Exception:
        typer.echo(f"  (raw response: {response.text[:200]})")


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
            f"{_api_base_url()}/v0/repositories",
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
        typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


def _fmt_size(size_bytes: int) -> str:
    """Format a byte count as a human-readable string."""
    for unit in ("B", "KB", "MB", "GB"):
        if size_bytes < 1024:
            return f"{size_bytes:.0f} {unit}"
        size_bytes /= 1024
    return f"{size_bytes:.1f} TB"


def _col_widths(rows: list[list[str]]) -> list[int]:
    """Return max width for each column across all rows."""
    if not rows:
        return []
    return [max(len(r[i]) for r in rows) for i in range(len(rows[0]))]


def _print_table(headers: list[str], rows: list[list[str]], col_colors: dict[int, str] | None = None) -> None:
    """Print a simple aligned table with a header separator."""
    all_rows = [headers] + rows
    widths = _col_widths(all_rows)
    sep = "  "

    # Header
    header_line = sep.join(h.ljust(widths[i]) for i, h in enumerate(headers))
    typer.echo(typer.style(header_line, bold=True))
    typer.echo(typer.style("-" * len(header_line), dim=True))

    for row in rows:
        parts = []
        for i, cell in enumerate(row):
            padded = cell.ljust(widths[i])
            if col_colors and i in col_colors:
                padded = typer.style(padded, fg=col_colors[i])
            parts.append(padded)
        typer.echo(sep.join(parts))


@app.command()
def list_repos(
    owner: str = typer.Argument(None, help="Owner username (defaults to all repositories)"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show tag details for each repository"),
) -> None:
    """List repositories on the hub."""
    jwt = _require_access_token()
    params = {"owner": owner} if owner else {}
    try:
        response = requests.get(
            f"{_api_base_url()}/v0/repositories",
            params=params,
            headers={"Authorization": f"Bearer {jwt}"},
            timeout=10,
        )
        if response.status_code == 200:
            repos = response.json().get("repositories", [])
            if not repos:
                typer.echo("No repositories found.")
                return

            if not verbose:
                headers = ["REPOSITORY", "VISIBILITY", "UPDATED", "DESCRIPTION"]
                rows = []
                for repo in repos:
                    updated = repo.get("updatedAt", "")[:10]
                    rows.append([
                        f"{repo.get('owner', '')}/{repo.get('slug', '')}",
                        repo.get("visibility", ""),
                        updated,
                        repo.get("description", ""),
                    ])
                _print_table(headers, rows, col_colors={0: typer.colors.GREEN, 1: typer.colors.YELLOW})
            else:
                # Fetch full details for each repo to get tag metadata
                for i, repo in enumerate(repos):
                    o = repo.get("owner", "")
                    slug = repo.get("slug", "")
                    vis = repo.get("visibility", "")
                    desc = repo.get("description", "")
                    updated = repo.get("updatedAt", "")[:10]

                    repo_label = typer.style(f"{o}/{slug}", fg=typer.colors.GREEN, bold=True)
                    vis_label = typer.style(f"[{vis}]", fg=typer.colors.YELLOW)
                    header_parts = f"{repo_label} {vis_label}"
                    if desc:
                        header_parts += f"  {desc}"
                    header_parts += f"  (updated {updated})"
                    typer.echo(header_parts)

                    detail_resp = requests.get(
                        f"{_api_base_url()}/v0/repositories/{o}/{slug}",
                        headers={"Authorization": f"Bearer {jwt}"},
                        timeout=10,
                    )
                    if detail_resp.status_code == 200:
                        tags = detail_resp.json().get("tags", [])
                        if not tags:
                            typer.echo("  (no tags)")
                        else:
                            tag_headers = ["  TAG", "DIGEST", "SIZE", "PUSHED BY", "UPDATED"]
                            tag_rows = []
                            for t in tags:
                                digest = t.get("digest", "")
                                short_digest = digest[:12] if digest else ""
                                size = _fmt_size(t.get("size", 0))
                                pushed_by = t.get("createdBy", "")
                                tag_updated = t.get("updatedAt", "")[:10]
                                tag_rows.append([
                                    f"  {t.get('name', '')}",
                                    short_digest,
                                    size,
                                    pushed_by,
                                    tag_updated,
                                ])
                            _print_table(tag_headers, tag_rows)
                    else:
                        typer.echo("  (could not fetch tag details)")

                    if i < len(repos) - 1:
                        typer.echo("")

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
        typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
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
    jwt = _require_access_token()
    visibility = "private" if private else "public"
    try:
        response = requests.post(
            f"{_api_base_url()}/v0/repositories",
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
        typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
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
    jwt = _require_access_token()
    if not yes:
        typer.confirm(f"Delete repository '{repo}'? This cannot be undone.", abort=True)
    try:
        response = requests.delete(
            f"{_api_base_url()}/v0/repositories/{owner}/{slug}",
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
        typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)


@app.command()
def save(
    name: str = typer.Argument(..., help="Model ref in owner/repo:tag format (e.g. alice/my-model:v1.0)"),
    output_dir: str = typer.Argument(None, help="Directory to write the output file. Defaults to current directory. Created automatically if it does not exist."),
    output: str = typer.Option(None, "--output", "-o", help="Full output file path (e.g. ./out/model.json). Overrides OUTPUT_DIR. Defaults to <tag>.json in the current directory."),
) -> None:
    """Save a model from the hub to a local JSON file.

    Checks the local cache first; if the model is not cached it is automatically
    pulled from the hub before saving.

    \b
    Examples:
      opendi save alice/my-model:v1.0                  # saves v1.0.json in current dir
      opendi save alice/my-model:v1.0 ./models/        # saves v1.0.json inside ./models/
      opendi save alice/my-model:v1.0 -o out/cdm.json  # saves to a specific file path
    """
    if "/" not in name or ":" not in name:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    repo_part, tag = name.rsplit(":", 1)
    owner, repo_slug = repo_part.split("/", 1)

    if not owner or not repo_slug or not tag:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    if output:
        out_path = Path(output)
    elif output_dir:
        out_path = Path(output_dir) / f"{tag}.json"
    else:
        out_path = Path(f"{tag}.json")

    # Check local cache first
    content = local_cache.get_model(owner, repo_slug, tag)

    if content is None:
        # Not in cache — implicitly pull from hub
        typer.echo(f"{name} not in local cache. Pulling from hub...", err=True)

        headers: dict[str, str] = {}
        token = credential_storage.load_access_token()
        if token:
            headers["Authorization"] = f"Bearer {token}"

        url = f"{_api_base_url()}/v0/repositories/{owner}/{repo_slug}/tags/{tag}/model"
        try:
            response = requests.get(url, headers=headers, timeout=30)
        except requests.ConnectionError:
            typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
            raise typer.Exit(1)
        except requests.Timeout:
            typer.echo("Request timed out. Please try again.", err=True)
            raise typer.Exit(1)

        if response.status_code == 404:
            typer.echo(f"Error: model not found: {name}", err=True)
            raise typer.Exit(1)
        if response.status_code in (401, 403):
            typer.echo("Access denied. Run `opendi login` if this is a private repository.", err=True)
            raise typer.Exit(1)
        if not response.ok:
            typer.echo(f"Server error {response.status_code}: {response.text}", err=True)
            raise typer.Exit(1)

        content = response.text
        digest = response.headers.get("ETag") or None
        local_cache.save_model(owner, repo_slug, tag, content, digest)

        # Verify it was actually cached
        content = local_cache.get_model(owner, repo_slug, tag)
        if content is None:
            typer.echo(f"Error: pulled {name} but could not retrieve it from local cache.", err=True)
            raise typer.Exit(1)

    if output_dir:
        Path(output_dir).mkdir(parents=True, exist_ok=True)

    out_path.write_text(content, encoding="utf-8")
    typer.echo(f"Saved {typer.style(name, fg=typer.colors.GREEN, bold=True)} → {out_path}")


@app.command()
def remove_local(
    name: str = typer.Argument(..., help="Model ref to remove: owner/repo:tag"),
) -> None:
    """Remove a model from the local cache."""
    if "/" not in name or ":" not in name:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    repo_part, tag = name.rsplit(":", 1)
    owner, repo_slug = repo_part.split("/", 1)

    if not owner or not repo_slug or not tag:
        typer.echo("Invalid format. Use: owner/repo:tag", err=True)
        raise typer.Exit(1)

    if not local_cache.remove_model(owner, repo_slug, tag):
        typer.echo(f"Tag not found in local cache: {name}", err=True)
        raise typer.Exit(1)

    typer.echo(f"Removed {typer.style(name, fg=typer.colors.GREEN, bold=True)} from local cache.")


@app.command()
def list_local() -> None:
    models = local_cache.list_models()
    if not models:
        typer.echo("No models in local cache. Use `opendi pull owner/repo:tag` to cache one.")
        return

    # Group by owner/repo
    groups: dict[str, list[dict]] = {}
    for m in models:
        key = f"{m['owner']}/{m['repo']}"
        groups.setdefault(key, []).append(m)

    for idx, (repo_key, tags) in enumerate(groups.items()):
        typer.echo(typer.style(repo_key, fg=typer.colors.GREEN, bold=True))
        tag_rows = [
            [f"  {t['tag']}", t["pulled_at"][:19].replace("T", " ") + " UTC"]
            for t in tags
        ]
        _print_table(["  TAG", "PULLED AT"], tag_rows)
        if idx < len(groups) - 1:
            typer.echo("")


@app.command()
def validate(
    path: str = typer.Argument(..., help="Path to a local CDM JSON file (e.g. v1.0.json), or a model ref (owner/repo:tag) that is already in the local cache."),
) -> None:
    """Validate a CDM model against the OpenDI schema.

    Accepts either a local file path or a cached model ref. To validate a ref
    that is not yet cached, run `opendi pull owner/repo:tag` first.

    \b
    Examples:
      opendi validate ./model.json          # validate a local file
      opendi validate alice/my-model:v1.0  # validate from local cache
    """
    raw: bytes

    if "/" in path and ":" in path:
        # Treat as model ref owner/repo:tag
        repo_part, tag = path.rsplit(":", 1)
        owner, repo_slug = repo_part.split("/", 1)

        if not owner or not repo_slug or not tag:
            typer.echo("Invalid format. Use: owner/repo:tag", err=True)
            raise typer.Exit(1)

        content = local_cache.get_model(owner, repo_slug, tag)
        if content is None:
            typer.echo(f"Model not found in local cache: {path}. Run `opendi pull {path}` first.", err=True)
            raise typer.Exit(1)

        raw = content.encode("utf-8")
    else:
        file_path = Path(path).expanduser()
        if not file_path.is_file():
            typer.echo(f"File not found: {path}", err=True)
            raise typer.Exit(1)

        try:
            raw = file_path.read_bytes()
        except OSError as e:
            typer.echo(f"Could not read file: {e}", err=True)
            raise typer.Exit(1)

    # Verify it's at least parseable JSON before sending
    try:
        json.loads(raw)
    except json.JSONDecodeError as e:
        typer.echo(f"Invalid JSON: {e}", err=True)
        raise typer.Exit(1)

    headers: dict[str, str] = {"Content-Type": "application/json"}
    if _current_token:
        headers["Authorization"] = f"Bearer {_current_token}"

    try:
        response = requests.post(
            f"{_api_base_url()}/v0/validate",
            data=raw,
            headers=headers,
            timeout=30,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to the hub at {_api_base_url()}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 200:
        typer.echo(typer.style("Valid CDM.", fg=typer.colors.GREEN, bold=True))
        return

    error = ""
    try:
        body = response.json()
        if isinstance(body, dict):
            error = str(body.get("error") or "")
    except (ValueError, TypeError):
        error = response.text or ""

    typer.echo(typer.style("Validation failed:", fg=typer.colors.RED, bold=True) + f" {error}", err=True)
    raise typer.Exit(1)


if __name__ == "__main__":
    app()
