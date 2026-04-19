"""Repository-scoped commands: inspect, create, delete, list."""

from __future__ import annotations

import logging
from typing import Optional

import requests
import typer

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
        typer.echo("Access denied. Run `opendi login` if this is a private repository.", err=True)
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
        typer.echo(f"Server error (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)

    return response.json()


def _print_repo_overview(repo: dict, *, verbose: bool) -> None:
    """Print repository summary and tag list."""
    tags = repo.get("tags", [])
    visibility = repo.get("visibility", "")
    description = repo.get("description", "")
    updated = (repo.get("updatedAt") or "")[:10]

    owner_display = repo.get("owner", "")
    slug_display = repo.get("slug", "")
    repo_label = typer.style(f"{owner_display}/{slug_display}", fg=typer.colors.GREEN, bold=True)
    vis_label = typer.style(f"[{visibility}]", fg=typer.colors.YELLOW)
    header = f"{repo_label}  {vis_label}"
    if updated:
        header += f"  Updated {updated}"
    typer.echo(header)
    if description:
        typer.echo(f"Description: {description}")

    if not tags:
        typer.echo("Tags: (none)")
    elif not verbose:
        tag_names = "  ".join(t.get("name", "") for t in tags)
        typer.echo(f"Tags: {tag_names}")
    else:
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
        help="When inspecting a repository: show per-tag digest table.",
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
        typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
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
        typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
        raise typer.Exit(1)
    else:
        typer.echo(f"Failed to delete repository (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)


# ── list repos ────────────────────────────────────────────────────────────────


def list_repos(
    owner: Optional[str] = typer.Argument(None, help="Filter by owner username"),
    scope: str = typer.Option("mine", "--scope", help="Scope: mine | shared | all"),
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

    effective_scope = "all" if all_repos else scope

    params: dict[str, str] = {}
    if owner:
        params["owner"] = owner
    if effective_scope and effective_scope != "mine":
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
        typer.echo("Not authorised. Run `opendi login` to sign in again.", err=True)
        raise typer.Exit(1)
    if response.status_code == 404:
        typer.echo(f"Owner '{owner}' not found.", err=True)
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
        rows = []
        col_colors: dict[int, str] = {}
        for repo in repos:
            repo_owner = repo.get("owner", "")
            slug = repo.get("slug", "")
            full_name = f"{repo_owner}/{slug}"
            updated = (repo.get("updatedAt") or "")[:10]
            vis = repo.get("visibility", "")
            rows.append([full_name, vis, updated, repo.get("description", "")])

        # Build per-row colors
        row_colors: list[dict[int, str]] = []
        for repo in repos:
            repo_owner = repo.get("owner", "")
            if me and repo_owner == me:
                row_colors.append({0: typer.colors.GREEN})
            else:
                row_colors.append({})

        # Print with manual coloring since print_table doesn't support per-row colors
        all_rows = [headers_row] + rows
        widths = shared.col_widths(all_rows)
        sep = "  "
        header_line = sep.join(h.ljust(widths[i]) for i, h in enumerate(headers_row))
        typer.echo(typer.style(header_line, bold=True))
        typer.echo(typer.style("-" * len(header_line), dim=True))
        for idx, row in enumerate(rows):
            parts = []
            for i, cell in enumerate(row):
                padded = cell.ljust(widths[i])
                if i == 0:
                    repo_owner = repos[idx].get("owner", "")
                    if me and repo_owner == me:
                        padded = typer.style(padded, fg=typer.colors.GREEN, bold=True)
                    else:
                        vis_scope = repos[idx].get("visibility", "")
                        if vis_scope == "public":
                            padded = typer.style(padded, fg=typer.colors.WHITE)
                elif i == 1:
                    vis = repos[idx].get("visibility", "")
                    padded = typer.style(
                        padded,
                        fg=typer.colors.YELLOW if vis == "public" else typer.colors.MAGENTA,
                    )
                parts.append(padded)
            typer.echo(sep.join(parts))
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
            line = f"{repo_label}  {vis_label}"
            if desc:
                line += f"  {desc}"
            if updated:
                line += f"  (updated {updated})"
            typer.echo(line)

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
