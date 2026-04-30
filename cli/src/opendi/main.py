"""CLI entry point. Called when the user runs the `opendi` command."""

from __future__ import annotations

import difflib
import json
import logging
import urllib.parse
from pathlib import Path
from typing import Optional

import requests
import typer
from rich.console import Console
from rich.text import Text

from opendi import auth, credential_storage, local_cache, log
from opendi.cmds import repo_cmds, shared, tag_cmds

logger = logging.getLogger(__name__)


def _print_login_browser_prompt(absolute_url: str) -> None:
    """Prompt for browser sign-in using a terminal hyperlink (no raw URL on screen)."""
    console = Console(highlight=False)
    console.print(
        Text.assemble(
            "Please sign in using your browser. ",
            "If the login page did not open automatically, open ",
            ("this link", f"link {absolute_url}"),
            " to continue.",
        )
    )


# ── Typer apps ────────────────────────────────────────────────────────────────

app = typer.Typer(
    name="opendi",
    no_args_is_help=True,
    help="OpenDI Model Hub CLI for discovering and managing CDM models.",
    rich_markup_mode="markdown",
)


# ── Shared helpers (also used by domain modules via shared.py) ─────────────────


def _api_base_url() -> str:
    return shared.api_base_url()


def _auth_headers(content_type: str | None = None) -> dict[str, str]:
    return shared.auth_headers(content_type)


def _response_error(response: requests.Response) -> str:
    return shared.response_error(response)


def _require_access_token() -> str:
    return shared.require_access_token()


# ── App callback ───────────────────────────────────────────────────────────────


@app.callback(invoke_without_command=True)
def main(
    _ctx: typer.Context,
    debug: bool = typer.Option(
        False,
        "--debug/--no-debug",
        help="Enable debug logging",
        is_eager=True,
        hidden=True,
    ),
) -> None:
    """[bold]OpenDI Model Hub CLI[/bold] — cross-platform client for the OpenDI hub."""
    log.configure_logging(logging.DEBUG if debug else logging.WARNING)
    shared.current_token = credential_storage.load_access_token()


# ── Auth commands (registered first so help lists Authentication before other panels) ─


@app.command(rich_help_panel="Authentication")
def login() -> None:
    """Log in to the OpenDI hub (browser-assisted device flow)."""
    api_base = _api_base_url()
    try:
        if shared.current_token:
            try:
                me = auth.get_current_user(api_base, shared.current_token)
                username = me.get("username")
                typer.echo(
                    f"Already logged in as "
                    f"{typer.style(username or 'you', fg=typer.colors.GREEN, bold=True)}."
                )
                return
            except auth.AuthAPIError:
                credential_storage.delete_access_token()
                shared.current_token = None

        code, login_url, expires_in = auth.start_cli_login(api_base)
        absolute_url = auth.open_login_url(api_base, login_url)
        _print_login_browser_prompt(absolute_url)
        token = auth.poll_cli_token(api_base, code, expires_in)
        credential_storage.store_access_token(token)
        shared.current_token = token
        me = auth.get_current_user(api_base, token)
        username = me.get("username")
        if username:
            typer.echo(
                f"Login successful. Logged in as "
                f"{typer.style(username, fg=typer.colors.GREEN, bold=True)}."
            )
        else:
            typer.echo("Login successful.")
    except KeyboardInterrupt:
        typer.echo("Login cancelled.", err=True)
        raise typer.Exit(1)
    except TimeoutError:
        typer.echo("Login timed out. Please try again.", err=True)
        raise typer.Exit(1)
    except (auth.AuthAPIError, RuntimeError) as e:
        typer.echo(f"Login failed: {e}", err=True)
        raise typer.Exit(1)


@app.command(rich_help_panel="Authentication")
def whoami() -> None:
    """Show the currently logged-in account (verifies token with the hub)."""
    api_base = _api_base_url()
    token = _require_access_token()
    try:
        me = auth.get_current_user(api_base, token)
    except auth.AuthAPIError:
        credential_storage.delete_access_token()
        shared.current_token = None
        shared.echo_opendi_login_hint("Session expired. Run ", " to sign in again.")
        raise typer.Exit(1)
    except RuntimeError as e:
        typer.echo(f"Failed to verify session: {e}", err=True)
        raise typer.Exit(1)
    username = me.get("username")
    email = me.get("email")
    if username:
        typer.echo(typer.style(username, fg=typer.colors.GREEN, bold=True))
    if email:
        typer.echo(email)
    if not username and not email:
        typer.echo("Logged in.")


@app.command(rich_help_panel="Authentication")
def logout() -> None:
    """Log out from the OpenDI hub (clears stored credentials)."""
    if credential_storage.delete_access_token():
        shared.current_token = None
        typer.echo("Logged out.")
    else:
        typer.echo("Not logged in.")


# Verb subgroups — assembled from domain module functions
_create_app = typer.Typer(
    no_args_is_help=True,
    help="Create resources (`create repo`).",
)
_create_app.command("repo")(repo_cmds.create_repo)

_delete_app = typer.Typer(
    no_args_is_help=True,
    help="Delete resources (`delete repo`, `delete tag`, `delete local`).",
)
_delete_app.command("repo")(repo_cmds.delete_repo)
_delete_app.command("tag")(tag_cmds.delete_tag)
_delete_app.command("local")(tag_cmds.delete_local)

_list_app = typer.Typer(
    no_args_is_help=True,
    help="List resources (`list repos`, `list local`).",
)
_list_app.command("repos")(repo_cmds.list_repos)
_list_app.command("local")(tag_cmds.list_local)

_add_app = typer.Typer(
    no_args_is_help=True,
    help="Add resources (`add tag`).",
)
_add_app.command("tag")(tag_cmds.add_tag)

app.add_typer(_create_app, name="create", rich_help_panel="Resource Commands")
app.add_typer(_delete_app, name="delete", rich_help_panel="Resource Commands")
app.add_typer(_list_app, name="list", rich_help_panel="Resource Commands")
app.add_typer(_add_app, name="add", rich_help_panel="Resource Commands")

app.command(rich_help_panel="Resource Commands")(repo_cmds.inspect)


# ── Table helpers (kept for diff/validate output) ──────────────────────────────


def _fmt_size(size_bytes: int) -> str:
    return shared.fmt_size(size_bytes)


# ── Core pull helper (used by pull + save + push implicit pull) ────────────────


def _pull_to_cache(
    api_base: str,
    owner: str,
    repo_slug: str,
    tag: str,
    *,
    silent: bool = False,
) -> str:
    """Resolve repo, refresh metadata, and fetch model only when needed.

    Returns the JSON content string.
    Raises typer.Exit(1) on any network or HTTP error.
    """
    # Step 1 — resolve repo to get canonical owner/slug and repo_id
    owner_enc = urllib.parse.quote(owner, safe="")
    slug_enc = urllib.parse.quote(repo_slug, safe="")
    repo_url = f"{api_base}/v0/repositories/{owner_enc}/{slug_enc}"
    logger.debug("Resolving repo %s/%s", owner, repo_slug)
    try:
        repo_resp = requests.get(repo_url, headers=_auth_headers(), timeout=10)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if repo_resp.status_code == 404:
        typer.echo(f"Repository {owner}/{repo_slug} not found.", err=True)
        raise typer.Exit(1)
    if repo_resp.status_code in (401, 403):
        shared.echo_opendi_login_hint(
            "Access denied. Run ",
            " if this is a private repository.",
        )
        raise typer.Exit(1)
    if not repo_resp.ok:
        typer.echo(f"Request failed (HTTP {repo_resp.status_code}).", err=True)
        raise typer.Exit(1)

    repo_data = repo_resp.json()
    repo_id: int | None = repo_data.get("id")
    canonical_owner: str = repo_data.get("owner", owner)
    canonical_slug: str = repo_data.get("slug", repo_slug)

    # Step 2 — compare metadata digest (when available)
    cached = local_cache.get_model_info(canonical_owner, canonical_slug, tag, repo_id=repo_id)
    cached_digest = cached["digest"] if cached else None
    previous_owner = str(cached.get("owner", "")) if cached else None
    previous_slug = str(cached.get("repo", "")) if cached else None
    remote_digest: str | None = None
    tags = repo_data.get("tags")
    if isinstance(tags, list):
        for item in tags:
            if isinstance(item, dict) and item.get("name") == tag:
                maybe_digest = item.get("digest")
                if isinstance(maybe_digest, str) and maybe_digest.strip():
                    remote_digest = maybe_digest.strip()
                break

    ref = f"{canonical_owner}/{canonical_slug}:{tag}"
    if (
        cached is not None
        and remote_digest is not None
        and cached_digest is not None
        and cached_digest == remote_digest
    ):
        local_cache.save_model(
            canonical_owner,
            canonical_slug,
            tag,
            cached["content"],
            digest=remote_digest,
            repo_id=repo_id,
            stale=0,
        )
        if not silent:
            typer.echo(typer.style(f"{ref} already up to date.", dim=True))
            # Metadata can still change even when digest is unchanged.
            if previous_slug and previous_slug != canonical_slug:
                typer.echo(
                    typer.style(
                        f"  metadata updated: slug {previous_slug} -> {canonical_slug}",
                        fg=typer.colors.YELLOW,
                    )
                )
            if previous_owner and previous_owner != canonical_owner:
                typer.echo(
                    typer.style(
                        f"  metadata updated: owner {previous_owner} -> {canonical_owner}",
                        fg=typer.colors.RED,
                        bold=True,
                    )
                )
        return cached["content"]

    # Step 3 — fetch model
    tag_enc = urllib.parse.quote(tag, safe="")
    model_url = (
        f"{api_base}/v0/repositories/"
        f"{urllib.parse.quote(canonical_owner, safe='')}/{urllib.parse.quote(canonical_slug, safe='')}"
        f"/tags/{tag_enc}/model"
    )
    fetch_headers = _auth_headers()

    logger.debug("Fetching model %s/%s:%s", canonical_owner, canonical_slug, tag)
    try:
        model_resp = requests.get(model_url, headers=fetch_headers, timeout=30)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if model_resp.status_code == 404:
        typer.echo(f"Model not found: {ref}", err=True)
        raise typer.Exit(1)
    if model_resp.status_code in (401, 403):
        shared.echo_opendi_login_hint(
            "Access denied. Run ",
            " if this is a private repository.",
        )
        raise typer.Exit(1)
    if not model_resp.ok:
        typer.echo(f"Request failed (HTTP {model_resp.status_code}).", err=True)
        raise typer.Exit(1)

    content = model_resp.text
    new_digest = remote_digest

    local_cache.save_model(
        canonical_owner,
        canonical_slug,
        tag,
        content,
        digest=new_digest,
        repo_id=repo_id,
    )

    if not silent:
        size_str = _fmt_size(len(content.encode()))
        digest_str = typer.style((new_digest or "")[:12], fg=typer.colors.CYAN)
        typer.echo(
            typer.style("Pulled", fg=typer.colors.GREEN, bold=True)
            + f" {ref}  {size_str}  digest: {digest_str}"
        )
        if cached is not None and cached_digest and new_digest and cached_digest != new_digest:
            typer.echo(typer.style("  metadata updated: digest changed", dim=True))
        if previous_slug and previous_slug != canonical_slug:
            typer.echo(
                typer.style(
                    f"  metadata updated: slug {previous_slug} -> {canonical_slug}",
                    fg=typer.colors.YELLOW,
                )
            )
        if previous_owner and previous_owner != canonical_owner:
            typer.echo(
                typer.style(
                    f"  metadata updated: owner {previous_owner} -> {canonical_owner}",
                    fg=typer.colors.RED,
                    bold=True,
                )
            )

    return content


def _fetch_remote_model_json(api_base: str, owner: str, slug: str, tag: str) -> dict:
    """GET /v0/repositories/{owner}/{slug}/tags/{tag}/model and return JSON object."""
    o = urllib.parse.quote(owner, safe="")
    sl = urllib.parse.quote(slug, safe="")
    t = urllib.parse.quote(tag, safe="")
    url = f"{api_base}/v0/repositories/{o}/{sl}/tags/{t}/model"
    try:
        response = requests.get(url, headers=_auth_headers(), timeout=60)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)
    if response.status_code == 200:
        return response.json()
    err = _response_error(response)
    el = err.lower()
    if response.status_code == 401:
        shared.echo_opendi_login_hint("Not authorized. Run ", " to sign in again.")
        raise typer.Exit(1)
    if response.status_code == 404:
        if "tag not found" in el:
            typer.echo(f"Tag not found: {owner}/{slug}:{tag}", err=True)
        elif "repository not found" in el:
            typer.echo(f"Repository not found: {owner}/{slug}", err=True)
        else:
            typer.echo("Repository not found or you do not have access.", err=True)
        raise typer.Exit(1)
    if response.status_code == 403:
        typer.echo("Not authorized to read this model.", err=True)
        raise typer.Exit(1)
    typer.echo(f"Failed to load model (HTTP {response.status_code}).", err=True)
    raise typer.Exit(1)


def _looks_like_local_path(s: str) -> bool:
    """True if ``s`` is more likely a filesystem path than a hub ``repo:tag`` ref."""
    if "\\" in s:
        return True
    if s.startswith("/"):
        return True
    # Windows drive letter path (``C:\...`` or ``C:/...``)
    return len(s) >= 3 and s[0].isalpha() and s[1] == ":" and s[2] in "\\/"


def _load_model_side(spec: str, api_base: str) -> dict:
    """Load a CDM model from a remote ref or a local JSON file."""
    s = spec.strip()
    path = Path(s).expanduser()
    if path.is_file():
        try:
            return json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as e:
            typer.echo(f"Invalid JSON in file: {e}", err=True)
            raise typer.Exit(1)

    if _looks_like_local_path(s):
        typer.echo(f"File not found: {spec}", err=True)
        raise typer.Exit(1)

    # No colon → not a tag-bearing hub ref; avoid mis-reading filenames.
    if ":" not in s and not s.startswith("id@"):
        typer.echo(
            f"Expected {shared.REF_HELP_TAG}, or an existing JSON file, got: {spec}",
            err=True,
        )
        raise typer.Exit(1)

    parsed = shared.parse_ref(s)
    if not parsed.tag:
        typer.echo(
            f"Remote side needs a tag ({shared.REF_HELP_TAG}).",
            err=True,
        )
        raise typer.Exit(1)
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    return _fetch_remote_model_json(api_base, owner, slug, parsed.tag)


# ── Hub discovery ──────────────────────────────────────────────────────────────


@app.command(rich_help_panel="Discovery")
def search(
    query: str = typer.Argument(..., help="Search query"),
    owner: Optional[str] = typer.Option(None, "--owner", help="Filter by owner username"),
    sort: str = typer.Option("updated", "--sort", help="Sort field: name | updated | created"),
    order: str = typer.Option("desc", "--order", help="Sort order: asc | desc"),
    limit: int = typer.Option(20, "--limit", help="Max results to display"),
) -> None:
    """Search public repositories on the hub. Shows more results when logged in."""
    api_base = _api_base_url()
    params: dict[str, str] = {"q": query, "sortBy": sort, "sortOrder": order}
    if owner:
        params["owner"] = owner

    try:
        response = requests.get(
            f"{api_base}/v0/search",
            params=params,
            headers=_auth_headers(),
            timeout=10,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if not response.ok:
        typer.echo(f"Search failed (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)

    repos: list[dict] = response.json().get("repositories", [])
    repos = repos[:limit]

    if not repos:
        typer.echo(f'No repositories found matching "{query}".')
        return

    me = shared.extract_username_from_token(shared.current_token or "")
    headers_row = ["REPOSITORY", "VISIBILITY", "DESCRIPTION"]
    rich_rows: list[list[Text]] = []

    for repo in repos:
        repo_owner = repo.get("owner", "")
        slug = repo.get("slug", "")
        vis = repo.get("visibility", "")
        desc = repo.get("description", "") or ""
        full_name = f"{repo_owner}/{slug}"
        name_style = shared.repo_table_name_style(me, repo_owner, vis)
        name_cell = Text(full_name, style=name_style) if name_style else Text(full_name)
        vis_cell = Text(vis, style="yellow" if vis == "public" else "magenta")
        rich_rows.append([name_cell, vis_cell, Text(desc)])

    shared.print_repo_table(headers_row, rich_rows)


# ── Core model operations ──────────────────────────────────────────────────────


@app.command(rich_help_panel="Model Operations")
def pull(
    name: str = typer.Argument(..., help=shared.HELP_ARG_MODEL_REF),
) -> None:
    """Pull a model from the hub into the local cache.

    Ref must include a tag. Forms: owner/repo:tag, repo:tag (when logged in),
    or id@<repo_id>:<tag> (e.g. id@42:v1, id@[42]:v1).
    """
    parsed = shared.parse_ref(name)

    if not parsed.slug and parsed.repo_id is None:
        typer.echo(
            f"Invalid format. Use {shared.REF_HELP_TAG}.",
            err=True,
        )
        raise typer.Exit(1)
    if not parsed.tag:
        typer.echo(
            f"Tag is required for pull. Use {shared.REF_HELP_TAG}.",
            err=True,
        )
        raise typer.Exit(1)

    api_base = _api_base_url()
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    _pull_to_cache(api_base, owner, slug, parsed.tag)


@app.command(rich_help_panel="Model Operations")
def push(
    ref: str = typer.Argument(..., help=shared.HELP_ARG_PUSH_REF),
    file: str = typer.Argument(..., help="Local CDM JSON file to upload"),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip all confirmation prompts"),
) -> None:
    """Push a local CDM JSON file to the hub (auth required).

    Destination ref uses the same tag-bearing forms as pull (including id@<id>:<tag>).

    \b
    Example:
      opendi push alice/my-model:v1 ./model.json
    """
    _require_access_token()
    parsed = shared.require_full_ref(ref, label="REF")
    api_base = _api_base_url()
    owner, repo_slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    tag = parsed.tag
    assert tag is not None

    file_path = Path(file)
    if not file_path.is_file():
        typer.echo(f"File not found: {file}", err=True)
        raise typer.Exit(1)

    # Pre-validate JSON
    try:
        raw = file_path.read_bytes()
        json.loads(raw)
    except json.JSONDecodeError as e:
        typer.echo(f"Invalid JSON in file: {e}", err=True)
        raise typer.Exit(1)
    except OSError as e:
        typer.echo(f"Could not read file: {e}", err=True)
        raise typer.Exit(1)

    headers = _auth_headers("application/json")
    url = (
        f"{api_base}/v0/repositories/"
        f"{urllib.parse.quote(owner, safe='')}/{urllib.parse.quote(repo_slug, safe='')}"
        f"/tags/{urllib.parse.quote(tag, safe='')}"
    )

    def _do_put(overwrite: bool = False) -> requests.Response:
        endpoint = url + ("?overwrite=true" if overwrite else "")
        try:
            return requests.put(endpoint, data=raw, headers=headers, timeout=60)
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {api_base}.", err=True)
            raise typer.Exit(1)
        except requests.Timeout:
            typer.echo("Request timed out. Please try again.", err=True)
            raise typer.Exit(1)

    response = _do_put()

    # 409 — tag already exists
    if response.status_code == 409:
        if not yes:
            confirmed = typer.confirm(
                f"Tag {tag} already exists in {owner}/{repo_slug}. Overwrite?",
                default=False,
            )
            if not confirmed:
                typer.echo("Operation canceled.")
                raise typer.Exit(0)
        response = _do_put(overwrite=True)

    # 404 — repo doesn't exist; offer to create it (owner must be self)
    if response.status_code == 404:
        me = shared.extract_username_from_token(shared.current_token or "")
        if me and me != owner:
            typer.echo(f"Repository {owner}/{repo_slug} not found.", err=True)
            raise typer.Exit(1)

        if yes:
            choice = "private"
        else:
            choice = typer.prompt(
                f"Repository {owner}/{repo_slug} not found. Create it?",
                default="N",
                prompt_suffix=" [public/private/N]: ",
            ).strip().lower()

        if choice not in ("public", "private"):
            typer.echo("Operation canceled.")
            raise typer.Exit(1)

        try:
            create_resp = requests.post(
                f"{api_base}/v0/repositories",
                json={"slug": repo_slug, "visibility": choice},
                headers=headers,
                timeout=30,
            )
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {api_base}.", err=True)
            raise typer.Exit(1)

        if not create_resp.ok:
            typer.echo(
                f"Failed to create repository (HTTP {create_resp.status_code}).",
                err=True,
            )
            raise typer.Exit(1)

        typer.echo(
            typer.style("Created", fg=typer.colors.GREEN, bold=True)
            + f" {owner}/{repo_slug} [{choice}]."
        )
        response = _do_put()

    if response.status_code == 400:
        typer.echo(
            typer.style("Validation error: ", fg=typer.colors.RED, bold=True)
            + _response_error(response),
            err=True,
        )
        raise typer.Exit(1)
    if response.status_code in (401, 403):
        typer.echo("Not authorized. You need write access to this repository.", err=True)
        raise typer.Exit(1)
    if not response.ok:
        typer.echo(f"Request failed (HTTP {response.status_code}): {response.text}", err=True)
        raise typer.Exit(1)

    # Success
    try:
        result = response.json()
        digest = result.get("digest", "")
        size = result.get("size", 0)
    except Exception:
        digest = ""
        size = 0

    typer.echo(
        typer.style("Pushed", fg=typer.colors.GREEN, bold=True)
        + f" {file_path.name} → {owner}/{repo_slug}:{tag}"
    )
    if digest:
        typer.echo(f"  digest: {typer.style(digest[:12], fg=typer.colors.CYAN)}")
    if size:
        typer.echo(f"  size:   {_fmt_size(size)}")

    # Implicit pull to populate cache with the hub's authoritative version
    typer.echo(typer.style("  Updating local cache...", dim=True))
    try:
        _pull_to_cache(api_base, owner, repo_slug, tag, silent=True)
    except typer.Exit:
        typer.echo(typer.style("  (cache update failed — run `opendi pull` manually)", dim=True))


@app.command(rich_help_panel="Model Operations")
def save(
    name: str = typer.Argument(..., help=shared.HELP_ARG_MODEL_REF),
    output_dir: Optional[str] = typer.Argument(
        None,
        help="Directory to write <tag>.json into. Defaults to current directory.",
    ),
    output: Optional[str] = typer.Option(
        None,
        "--output", "-o",
        help="Exact output file path (e.g. ./out/model.json). Overrides OUTPUT_DIR.",
    ),
) -> None:
    """Save a model from the hub to a local JSON file.

    Ref with tag: owner/repo:tag, repo:tag when logged in, or id@<repo_id>:<tag>.

    Uses the local cache when possible; otherwise pulls from the hub (same path as ``pull``).

    \b
    Examples:
      opendi save alice/my-model:v1                  # saves v1.json in current dir
      opendi save alice/my-model:v1 ./models/        # saves v1.json inside ./models/
      opendi save alice/my-model:v1 -o out/cdm.json  # saves to a specific path
    """
    if output and output_dir:
        typer.echo("Specify either OUTPUT_DIR or --output, not both.", err=True)
        raise typer.Exit(1)

    parsed = shared.require_full_ref(name)
    api_base = _api_base_url()
    owner, repo_slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    tag = parsed.tag
    assert tag is not None

    if output:
        out_path = Path(output)
    elif output_dir:
        out_path = Path(output_dir) / f"{tag}.json"
        Path(output_dir).mkdir(parents=True, exist_ok=True)
    else:
        out_path = Path(f"{tag}.json")

    content = local_cache.get_model(owner, repo_slug, tag)
    if content is None:
        typer.echo(typer.style(f"{name} not in local cache. Pulling from hub...", dim=True))
        content = _pull_to_cache(api_base, owner, repo_slug, tag, silent=True)

    out_path.write_text(content, encoding="utf-8")
    typer.echo(
        typer.style("Saved", fg=typer.colors.GREEN, bold=True)
        + f" {name} → {out_path}"
    )


@app.command(rich_help_panel="Model Operations")
def diff(
    left: str = typer.Argument(
        ...,
        help=shared.HELP_ARG_DIFF_SIDE,
    ),
    right: str = typer.Argument(
        ...,
        help=shared.HELP_ARG_DIFF_SIDE,
    ),
) -> None:
    """Compare two CDM models: each side is a tag-bearing hub ref (same forms as pull) or a JSON file path."""
    api_base = _api_base_url()
    left_obj = _load_model_side(left, api_base)
    right_obj = _load_model_side(right, api_base)

    if json.dumps(left_obj, sort_keys=True) == json.dumps(right_obj, sort_keys=True):
        typer.echo("No differences (normalized JSON is identical).")
        return

    a = json.dumps(left_obj, sort_keys=True, indent=2, ensure_ascii=False) + "\n"
    b = json.dumps(right_obj, sort_keys=True, indent=2, ensure_ascii=False) + "\n"
    # Default lineterm="\n" is required: lineterm="" glues ---/+++ header lines with no
    # newline between them, which breaks layout (especially noticeable in PowerShell).
    out = "".join(
        difflib.unified_diff(
            a.splitlines(keepends=True),
            b.splitlines(keepends=True),
            fromfile=left,
            tofile=right,
        )
    )
    typer.echo(out, nl=not out.endswith("\n"))


def _format_validation_issue(issue: dict) -> str:
    path = str(issue.get("instancePath") or "/")
    msg = str(issue.get("message") or "Validation error")
    phase = str(issue.get("phase") or "")
    line = issue.get("line")
    col = issue.get("column")
    loc = ""
    if isinstance(line, int) and line > 0:
        if isinstance(col, int) and col > 0:
            loc = f"line {line}, col {col} — "
        else:
            loc = f"line {line} — "
    phase_prefix = f"[{phase}] " if phase else ""
    return f"- {phase_prefix}{loc}{path}: {msg}"


@app.command(rich_help_panel="Model Operations")
def validate(
    path: str = typer.Argument(
        ...,
        help="Path to a local CDM JSON file (e.g. model.json)",
    ),
) -> None:
    """Validate a local CDM JSON file against the OpenDI schema.

    \b
    Example:
      opendi validate ./model.json
    """
    file_path = Path(path).expanduser()
    if not file_path.is_file():
        typer.echo(f"File not found: {path}", err=True)
        raise typer.Exit(1)
    try:
        raw = file_path.read_bytes()
    except OSError as e:
        typer.echo(f"Could not read file: {e}", err=True)
        raise typer.Exit(1)

    try:
        json.loads(raw)
    except json.JSONDecodeError as e:
        typer.echo(f"Invalid JSON: {e}", err=True)
        raise typer.Exit(1)

    api_base = _api_base_url()
    try:
        response = requests.post(
            f"{api_base}/v0/validate",
            data=raw,
            headers=_auth_headers("application/json"),
            timeout=30,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 200:
        typer.echo(typer.style("Valid CDM.", fg=typer.colors.GREEN, bold=True))
        return

    payload: dict = {}
    try:
        maybe = response.json()
        if isinstance(maybe, dict):
            payload = maybe
    except ValueError:
        payload = {}

    error = str(payload.get("error") or response.text or f"HTTP {response.status_code}")
    details = payload.get("details")
    typer.echo(
        typer.style("Validation failed:", fg=typer.colors.RED, bold=True) + f" {error}",
        err=True,
    )
    if isinstance(details, list) and details:
        for issue in details[:50]:
            if isinstance(issue, dict):
                typer.echo(_format_validation_issue(issue), err=True)
        if len(details) > 50:
            typer.echo(f"... and {len(details) - 50} more issue(s).", err=True)
    raise typer.Exit(1)


if __name__ == "__main__":
    app()
