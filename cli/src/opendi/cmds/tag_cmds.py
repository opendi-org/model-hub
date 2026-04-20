"""Tag-scoped commands: add tag, delete tag, delete local, list local."""

from __future__ import annotations

import logging
from typing import Optional

import requests
import typer

from opendi import local_cache
from opendi.cmds import shared

logger = logging.getLogger(__name__)


# ── add tag ───────────────────────────────────────────────────────────────────


def add_tag(
    source_ref: str = typer.Argument(
        ...,
        help=shared.HELP_ARG_ADD_TAG_SOURCE,
    ),
    target_tag: str = typer.Argument(
        ...,
        metavar="TARGET_TAG",
        help="New tag name in the same repository (required).",
    ),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation on overwrite"),
) -> None:
    """Create a new tag pointing to the same model as an existing tag (auth required).

    Both tags reside in the same repository. No file upload is required.

    \b
    Example:
      opendi add tag alice/my-model:v1 stable
      → creates alice/my-model:stable pointing to the same model as :v1
    """
    shared.require_access_token()
    parsed = shared.require_full_ref(source_ref, label="SOURCE")
    api_base = shared.api_base_url()
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    source = parsed.tag

    url = f"{api_base}/v0/repositories/{owner}/{slug}/tags/{target_tag}"
    body = {"sourceTag": source}
    headers = shared.auth_headers()

    def _do_put(overwrite: bool = False) -> requests.Response:
        endpoint = url + ("?overwrite=true" if overwrite else "")
        try:
            return requests.put(endpoint, json=body, headers=headers, timeout=30)
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {api_base}.", err=True)
            raise typer.Exit(1)
        except requests.Timeout:
            typer.echo("Request timed out. Please try again.", err=True)
            raise typer.Exit(1)

    response = _do_put()

    if response.status_code == 409:
        if not yes:
            confirmed = typer.confirm(
                f"Tag {target_tag} already exists in {owner}/{slug}. Overwrite?",
                default=False,
            )
            if not confirmed:
                typer.echo("Operation canceled.")
                raise typer.Exit(0)
        response = _do_put(overwrite=True)

    if response.status_code in (200, 201):
        typer.echo(
            typer.style("Added", fg=typer.colors.GREEN, bold=True)
            + f" {owner}/{slug}:{target_tag} → same model as :{source}"
        )
    elif response.status_code in (401, 403):
        typer.echo("Not authorized. You need write access to this repository.", err=True)
        raise typer.Exit(1)
    elif response.status_code == 404:
        typer.echo(f"Source tag {source} not found in {owner}/{slug}.", err=True)
        raise typer.Exit(1)
    else:
        err = shared.response_error(response)
        typer.echo(f"Request failed (HTTP {response.status_code}): {err}", err=True)
        raise typer.Exit(1)


# ── delete tag ────────────────────────────────────────────────────────────────


def delete_tag(
    ref: str = typer.Argument(..., help=shared.HELP_ARG_DELETE_TAG),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation prompt"),
) -> None:
    """Delete a remote tag from the hub (auth required)."""
    parsed = shared.require_full_ref(ref)
    shared.require_access_token()
    api_base = shared.api_base_url()
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    tag = parsed.tag

    if not yes:
        typer.confirm(
            f"Delete tag {tag} from {owner}/{slug}?",
            abort=True,
        )

    try:
        response = requests.delete(
            f"{api_base}/v0/repositories/{owner}/{slug}/tags/{tag}",
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
        local_cache.mark_stale(owner, slug, tag)  # type: ignore[arg-type]
        typer.echo(
            typer.style("Deleted", fg=typer.colors.GREEN, bold=True)
            + f" {owner}/{slug}:{tag}."
        )
        typer.echo(typer.style("  Local cache entry marked stale.", dim=True))
    elif response.status_code in (401, 403):
        shared.echo_opendi_login_hint("Not authorized. Run ", " to sign in again.")
        raise typer.Exit(1)
    elif response.status_code == 404:
        typer.echo(f"Tag {tag} not found in {owner}/{slug}.", err=True)
        raise typer.Exit(1)
    else:
        typer.echo(f"Request failed (HTTP {response.status_code}).", err=True)
        raise typer.Exit(1)


# ── delete local ──────────────────────────────────────────────────────────────


def delete_local(
    ref: Optional[str] = typer.Argument(None, help=shared.HELP_ARG_DELETE_LOCAL),
    all_entries: bool = typer.Option(False, "--all", help="Remove ALL entries from the local cache"),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation for --all"),
) -> None:
    """Remove one or all entries from the local cache (no network call).

    \b
    Examples:
      opendi delete local alice/my-model:v1   # remove a specific entry
      opendi delete local --all               # purge entire cache
    """
    if all_entries and ref is not None:
        typer.echo("Specify either a REF or --all, not both.", err=True)
        raise typer.Exit(1)

    if all_entries:
        if not yes:
            typer.confirm("Remove all locally cached models?", abort=True)
        count = local_cache.remove_all()
        typer.echo(
            typer.style(f"Cleared local cache. ({count} entries removed.)", fg=typer.colors.GREEN, bold=True)
        )
        return

    if ref is None:
        typer.echo(f"Provide a REF ({shared.REF_HELP_TAG}) or use --all.", err=True)
        raise typer.Exit(1)

    parsed = shared.require_full_ref(ref)
    api_base = shared.api_base_url()
    owner, slug = shared.resolve_parsed_ref_to_owner_slug(parsed, api_base)
    tag = parsed.tag

    if not local_cache.remove_model(owner, slug, tag):
        typer.echo(f"Tag not in local cache: {ref}", err=True)
        raise typer.Exit(1)
    typer.echo(
        typer.style("Removed", fg=typer.colors.GREEN, bold=True)
        + f" {owner}/{slug}:{tag} from local cache."
    )


# ── list local ────────────────────────────────────────────────────────────────


def list_local(
    filter_repo: Optional[str] = typer.Argument(
        None,
        help="Filter by owner/repo (e.g. alice/my-model)",
    ),
    clean: bool = typer.Option(False, "--clean", help="Remove all stale entries without confirmation"),
) -> None:
    """List all models in the local cache."""
    if clean:
        count = local_cache.remove_stale()
        typer.echo(f"Removed {count} stale cache entries.")
        return

    filter_owner: str | None = None
    filter_slug: str | None = None
    if filter_repo:
        if "/" not in filter_repo:
            typer.echo("Filter must be owner/repo format (e.g. alice/my-model).", err=True)
            raise typer.Exit(1)
        filter_owner, filter_slug = filter_repo.split("/", 1)

    models = local_cache.list_models(filter_owner=filter_owner, filter_repo=filter_slug)

    if not models:
        if filter_repo:
            typer.echo(f"No cached models for {filter_repo}.")
        else:
            typer.echo(
                "No models in local cache. Use "
                + typer.style("opendi pull", fg=typer.colors.CYAN, bold=True)
                + f" with a ref ({shared.REF_HELP_TAG}) to cache one."
            )
        return

    headers_row = ["REPOSITORY", "TAG", "DIGEST", "PULLED", "STATUS"]
    rows = []
    dim_flags = []

    for m in models:
        repo_label = f"{m['owner']}/{m['repo']}"
        tag = m["tag"]
        digest = (m.get("digest") or "")[:12]
        pulled = (m.get("pulled_at") or "")[:19].replace("T", " ") + " UTC"
        stale = bool(m.get("stale"))
        status = typer.style("[stale]", dim=True) if stale else "ok"
        rows.append([repo_label, tag, digest, pulled, status])
        dim_flags.append(stale)

    col_colors = {2: typer.colors.CYAN}
    shared.print_table(headers_row, rows, col_colors=col_colors, dim_rows=dim_flags)
