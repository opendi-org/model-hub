"""Repository-scoped commands: inspect, create, delete, list."""

from __future__ import annotations

import logging
from typing import Optional

import requests
import typer
from rich.text import Text

from opendi import local_cache
from opendi.cmds import shared

logger = logging.getLogger(__name__)


# ── inspect (repo vs tag from REF) ─────────────────────────────────────────────


def _repo_url(api_base: str, parsed: shared.ParsedRef) -> str:
    """Build GET URL for repository metadata (owner/slug or numeric id)."""
    if parsed.repo_id is not None:
        return f"{api_base}/v0/repo/{parsed.repo_id}"
    assert parsed.owner is not None and parsed.slug is not None
    return f"{api_base}/v0/repositories/{parsed.owner}/{parsed.slug}"


def _fetch_repo(api_base: str, parsed: shared.ParsedRef) -> dict:
    """GET repository JSON or exit with an error."""
    if parsed.repo_id is None and (not parsed.owner or not parsed.slug):
        typer.echo(
            "Invalid ref for inspect. Use owner/repo, owner/repo:tag, repo:tag (when logged in), "
            "id@<repo_id>, or id@<repo_id>:<tag>.",
            err=True,
        )
        raise typer.Exit(1)

    url = _repo_url(api_base, parsed)
    try:
        response = requests.get(url, headers=shared.auth_headers(), timeout=10)
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code in (401, 403):
        shared.echo_opendi_login_hint(
            "Access denied. Run ",
            " if this is a private repository.",
        )
        raise typer.Exit(1)
    if response.status_code == 404:
        ref_hint = (
            f"{parsed.owner}/{parsed.slug}"
            if parsed.owner and parsed.slug
            else f"repo id {parsed.repo_id}"
        )
        typer.echo(f"Repository {ref_hint} not found.", err=True)
        raise typer.Exit(1)
    if not response.ok:
        typer.echo(f"Request failed (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)

    return response.json()


def _fmt_repo_timestamp(value: object) -> str:
    """Format an API ISO-8601 timestamp for display."""
    if value is None:
        return ""
    s = str(value).strip()
    if not s:
        return ""
    if len(s) >= 19 and s[10] == "T":
        return s[:19].replace("T", " ") + " UTC"
    return s[:10] if len(s) >= 10 else s


def _lineage_ref_str(ref: dict) -> str:
    """Single-line owner/slug plus hub id for a lineage ref."""
    owner = ref.get("owner") or ""
    slug = ref.get("slug") or ""
    rid = ref.get("id")
    label = f"{owner}/{slug}".strip("/")
    if rid is not None:
        return f"{label}  id@{rid}" if label else f"id@{rid}"
    return label or "—"


def _print_repo_overview(repo: dict, *, verbose: bool) -> None:
    """Print repository summary, optional collaborators/lineage (verbose), and tags."""
    tags = repo.get("tags") or []
    visibility = repo.get("visibility", "")
    description = repo.get("description", "")
    updated_short = (repo.get("updatedAt") or "")[:10]

    owner_display = repo.get("owner", "")
    slug_display = repo.get("slug", "")
    if verbose:
        rid = repo.get("id")
        repo_ref = typer.style(f"{owner_display}/{slug_display}", fg=typer.colors.GREEN, bold=True)
        if rid is not None:
            repo_ref = f"{repo_ref}  id@{rid}"
        vis_label = typer.style(f"[{visibility or 'unknown'}]", fg=typer.colors.YELLOW)
        created = _fmt_repo_timestamp(repo.get("createdAt"))
        updated_full = _fmt_repo_timestamp(repo.get("updatedAt"))
        typer.echo(f"{'Repository':<12}: {repo_ref}")
        typer.echo(f"{'Visibility':<12}: {vis_label}")
        if created:
            typer.echo(f"{'Created':<12}: {created}")
        if updated_full:
            typer.echo(f"{'Updated':<12}: {updated_full}")
        typer.echo("")
    else:
        repo_label = typer.style(f"{owner_display}/{slug_display}", fg=typer.colors.GREEN, bold=True)
        vis_label = typer.style(f"[{visibility}]", fg=typer.colors.YELLOW)
        header = f"{repo_label}  {vis_label}"
        if updated_short:
            header += f"  Updated {updated_short}"
        typer.echo(header)

    if description:
        typer.echo(typer.style("Description", bold=True) + ":")
        desc_w = shared.terminal_columns() - 2
        for line in shared.wrap_text_block(description, desc_w):
            typer.echo(f"  {line}")
        if verbose:
            typer.echo("")

    has_private = "collaborators" in repo or "lineage" in repo
    if verbose:
        if not has_private:
            typer.echo(
                typer.style(
                    "Collaborators and fork lineage are not included for your current access "
                    "(owner or collaborator access is required).",
                    dim=True,
                )
            )
            typer.echo("")
        else:
            collabs = repo.get("collaborators")
            if collabs is not None:
                typer.echo(typer.style("Collaborators", bold=True) + ":")
                if not collabs:
                    typer.echo(typer.style("  (none)", dim=True))
                else:
                    rows = [[str(c.get("username", "")), str(c.get("role", ""))] for c in collabs]
                    shared.print_table(["USERNAME", "ROLE"], rows)
                typer.echo("")

            lineage = repo.get("lineage") or {}
            typer.echo(typer.style("Fork lineage", bold=True) + ":")
            parent = lineage.get("parent")
            ancestors = lineage.get("ancestors") or []
            children = lineage.get("children") or []
            if not parent and not ancestors and not children:
                typer.echo(typer.style("  (no fork relationships)", dim=True))
            else:
                if parent:
                    typer.echo(f"  Parent:     {_lineage_ref_str(parent)}")
                if ancestors:
                    typer.echo("  Ancestors:")
                    for a in ancestors:
                        typer.echo(f"    {_lineage_ref_str(a)}")
                if children:
                    typer.echo("  Forks:")
                    for ch in children:
                        typer.echo(f"    {_lineage_ref_str(ch)}")
            typer.echo("")

    if not tags:
        typer.echo(typer.style("Tags", bold=True) + ": " + typer.style("(none)", dim=True))
    elif not verbose:
        tag_names = "  ".join(t.get("name", "") for t in tags)
        typer.echo(typer.style("Tags", bold=True) + f": {tag_names}")
    else:
        typer.echo(typer.style("Tags", bold=True) + ":")
        tag_headers = ["TAG", "DIGEST", "SIZE", "PUSHED BY", "UPDATED"]
        tag_rows = []
        for t in tags:
            digest = t.get("digest", "")
            short_digest = typer.style(digest[:12], fg=typer.colors.CYAN) if digest else ""
            size = shared.fmt_size(t.get("size", 0))
            tag_rows.append([
                typer.style(t.get("name", ""), fg=typer.colors.GREEN, bold=True),
                short_digest,
                size,
                t.get("createdBy", ""),
                (t.get("updatedAt") or "")[:10],
            ])
        shared.print_table(tag_headers, tag_rows)


def _print_tag_detail(repo: dict, tag_name: str) -> None:
    """Print one tag's metadata."""
    tags = repo.get("tags", [])
    tag_info = next((t for t in tags if t.get("name") == tag_name), None)

    if tag_info is None:
        owner_display = repo.get("owner", "")
        slug_display = repo.get("slug", "")
        typer.echo(f"Tag '{tag_name}' not found in {owner_display}/{slug_display}.", err=True)
        raise typer.Exit(1)

    canonical_owner = repo.get("owner", "")
    canonical_slug = repo.get("slug", "")
    digest = tag_info.get("digest", "")

    typer.echo(f"Repository:  {canonical_owner}/{canonical_slug}")
    typer.echo(f"Tag:         {typer.style(tag_name, fg=typer.colors.GREEN, bold=True)}")
    typer.echo(f"Digest:      {typer.style(digest, fg=typer.colors.CYAN)}")
    typer.echo(f"Size:        {shared.fmt_size(tag_info.get('size', 0))}")
    typer.echo(f"Updated:     {tag_info.get('updatedAt', 'unknown')}")
    typer.echo(f"Pushed by:   {tag_info.get('createdBy', 'unknown')}")


def inspect(
    ref: str = typer.Argument(
        ...,
        help=shared.HELP_ARG_INSPECT_REF,
    ),
    verbose: bool = typer.Option(
        False,
        "--verbose",
        "-v",
        help=(
            "Repository ref: full timestamps, collaborators and fork lineage (when the API "
            "includes them), and a per-tag digest table."
        ),
    ),
) -> None:
    """Show repository metadata, or tag metadata if the ref includes a tag."""
    parsed = shared.parse_ref(ref)
    api_base = shared.api_base_url()

    if parsed.tag is not None:
        repo = _fetch_repo(api_base, parsed)
        _print_tag_detail(repo, parsed.tag)
        return

    if parsed.repo_id is None and (not parsed.owner or not parsed.slug):
        typer.echo(
            "Invalid ref. Use owner/repo, owner/repo:tag, repo:tag when logged in, "
            "id@<repo_id>, or id@<repo_id>:<tag>.",
            err=True,
        )
        raise typer.Exit(1)

    repo = _fetch_repo(api_base, parsed)
    _print_repo_overview(repo, verbose=verbose)


# ── create repo ───────────────────────────────────────────────────────────────


def create_repo(
    name: str = typer.Argument(..., help="Repository slug (e.g. my-model)"),
    description: str = typer.Option("", "--description", "-d", help="Short description"),
    private: bool = typer.Option(True, "--private/--public", help="Visibility (default: private)"),
) -> None:
    """Create a new repository on the hub (auth required)."""
    shared.require_access_token()
    api_base = shared.api_base_url()
    visibility = "private" if private else "public"
    try:
        response = requests.post(
            f"{api_base}/v0/repositories",
            json={"slug": name, "description": description, "visibility": visibility},
            headers=shared.auth_headers(),
            timeout=10,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 201:
        me = shared.extract_username_from_token(shared.current_token or "")
        full_name = f"{me}/{name}" if me else name
        typer.echo(
            typer.style("Created", fg=typer.colors.GREEN, bold=True)
            + f" {full_name} [{visibility}]"
        )
    elif response.status_code == 409:
        typer.echo(f"Repository '{name}' already exists.", err=True)
        raise typer.Exit(1)
    elif response.status_code == 401:
        shared.echo_opendi_login_hint("Not authorized. Run ", " to sign in again.")
        raise typer.Exit(1)
    else:
        typer.echo(f"Failed to create repository (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)


# ── delete repo ───────────────────────────────────────────────────────────────


def delete_repo(
    repo: str = typer.Argument(..., help=shared.HELP_ARG_DELETE_REPO),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation prompt"),
) -> None:
    """Permanently delete a repository and all its tags (auth required)."""
    shared.require_access_token()
    parsed = shared.parse_ref(repo.strip())
    if parsed.repo_id is None and (not parsed.owner or not parsed.slug):
        typer.echo(
            "Invalid ref. Use owner/repo, owner/repo:tag, repo or repo:tag when logged in, "
            "id@<repo_id>, or id@<repo_id>:<tag>.",
            err=True,
        )
        raise typer.Exit(1)

    api_base = shared.api_base_url()
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    display = f"{owner}/{slug}"

    if not yes:
        typer.confirm(
            f"Delete {display} and all its tags? This cannot be undone.",
            abort=True,
        )

    try:
        response = requests.delete(
            f"{api_base}/v0/repositories/{owner}/{slug}",
            headers=shared.auth_headers(),
            timeout=10,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 204:
        count = local_cache.mark_stale_by_repo(owner, slug)
        typer.echo(typer.style(f"Deleted {display}.", fg=typer.colors.GREEN, bold=True))
        if count:
            typer.echo(typer.style(f"  {count} local cache entry/entries marked stale.", dim=True))
    elif response.status_code in (403, 404):
        typer.echo(f"Repository '{display}' does not exist or you are not the owner.", err=True)
        raise typer.Exit(1)
    elif response.status_code == 401:
        shared.echo_opendi_login_hint("Not authorized. Run ", " to sign in again.")
        raise typer.Exit(1)
    else:
        typer.echo(f"Failed to delete repository (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)


# ── list repos ────────────────────────────────────────────────────────────────


def list_repos(
    owner: Optional[str] = typer.Argument(None, help="Filter by owner username"),
    scope: str = typer.Option(
        "mine",
        "--scope",
        help="Scope: mine | shared | all",
    ),
    all_repos: bool = typer.Option(False, "--all", help="Show all accessible repos (shorthand for --scope all)"),
    visibility: Optional[str] = typer.Option(None, "--visibility", help="Filter: public | private"),
    sort: str = typer.Option("updated", "--sort", help="Sort field: name | updated | created"),
    order: str = typer.Option("desc", "--order", help="Sort order: asc | desc"),
    limit: int = typer.Option(20, "--limit", help="Max results to display"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Full detail per repo with tag table"),
) -> None:
    """List repositories on the hub (auth required)."""
    shared.require_access_token()
    api_base = shared.api_base_url()

    effective_scope = "all" if all_repos else scope.strip().lower()
    if effective_scope == "shared":
        # API canonical value
        effective_scope = "shared-with-me"
    allowed_scopes = {"mine", "shared-with-me", "all"}
    if effective_scope not in allowed_scopes:
        typer.echo("Invalid --scope. Use one of: mine, shared, all.", err=True)
        raise typer.Exit(1)

    params: dict[str, str] = {}
    if owner:
        params["owner"] = owner
    params["scope"] = effective_scope
    if visibility:
        params["visibility"] = visibility
    if sort:
        params["sortBy"] = sort
    if order:
        params["sortOrder"] = order

    try:
        response = requests.get(
            f"{api_base}/v0/repositories",
            params=params,
            headers=shared.auth_headers(),
            timeout=10,
        )
    except requests.ConnectionError:
        typer.echo(f"Could not connect to {api_base}.", err=True)
        raise typer.Exit(1)
    except requests.Timeout:
        typer.echo("Request timed out. Please try again.", err=True)
        raise typer.Exit(1)

    if response.status_code == 401:
        shared.echo_opendi_login_hint("Not authorized. Run ", " to sign in again.")
        raise typer.Exit(1)
    if not response.ok:
        typer.echo(f"Failed to list repositories (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)

    repos: list[dict] = response.json().get("repositories", [])

    # Client-side limit slice
    repos = repos[:limit]

    if not repos:
        typer.echo("No repositories found.")
        return

    # Determine the current username for ownership coloring
    me = shared.extract_username_from_token(shared.current_token or "")

    if not verbose:
        headers_row = ["REPOSITORY", "VISIBILITY", "UPDATED", "DESCRIPTION"]
        rich_rows: list[list[Text]] = []
        for repo in repos:
            repo_owner = repo.get("owner", "")
            slug = repo.get("slug", "")
            full_name = f"{repo_owner}/{slug}"
            updated = (repo.get("updatedAt") or "")[:10]
            vis = repo.get("visibility", "")
            description = repo.get("description", "") or ""
            name_style = shared.repo_table_name_style(me, repo_owner, vis)
            name_cell = Text(full_name, style=name_style) if name_style else Text(full_name)
            vis_cell = Text(vis, style="yellow" if vis == "public" else "magenta")
            rich_rows.append(
                [
                    name_cell,
                    vis_cell,
                    Text(updated),
                    Text(description),
                ]
            )
        shared.print_repo_table(headers_row, rich_rows)
    else:
        for i, repo in enumerate(repos):
            repo_owner = repo.get("owner", "")
            slug = repo.get("slug", "")
            vis = repo.get("visibility", "")
            desc = repo.get("description", "")
            updated = (repo.get("updatedAt") or "")[:10]

            name_color = typer.colors.GREEN if (me and repo_owner == me) else typer.colors.WHITE
            repo_label = typer.style(f"{repo_owner}/{slug}", fg=name_color, bold=True)
            vis_label = typer.style(
                f"[{vis}]",
                fg=typer.colors.YELLOW if vis == "public" else typer.colors.MAGENTA,
            )
            typer.echo(f"{repo_label}  {vis_label}")
            if updated:
                typer.echo(f"  Updated {updated}")
            if desc:
                typer.echo("  Description:")
                desc_w = shared.terminal_columns() - 4
                for line in shared.wrap_text_block(desc, desc_w):
                    typer.echo(f"    {line}")

            try:
                detail_resp = requests.get(
                    f"{api_base}/v0/repositories/{repo_owner}/{slug}",
                    headers=shared.auth_headers(),
                    timeout=10,
                )
                if detail_resp.status_code == 200:
                    tags = detail_resp.json().get("tags", [])
                    if not tags:
                        typer.echo("  (no tags)")
                    else:
                        tag_headers = ["  TAG", "DIGEST", "SIZE", "PUSHED BY", "UPDATED"]
                        tag_rows = [
                            [
                                f"  {t.get('name', '')}",
                                (t.get("digest") or "")[:12],
                                shared.fmt_size(t.get("size", 0)),
                                t.get("createdBy", ""),
                                (t.get("updatedAt") or "")[:10],
                            ]
                            for t in tags
                        ]
                        shared.print_table(tag_headers, tag_rows)
                else:
                    typer.echo("  (could not fetch tag details)")
            except (requests.ConnectionError, requests.Timeout):
                typer.echo("  (could not fetch tag details)")

            if i < len(repos) - 1:
                typer.echo("")
