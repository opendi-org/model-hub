"""Shared state and helpers for all command modules.

`current_token` is set by the main app callback before any command runs.
Domain command modules read it at call time via ``from opendi.cmds import shared``.
"""

from __future__ import annotations

import base64
import json
import logging
import os
from typing import NamedTuple

import requests
import typer

logger = logging.getLogger(__name__)

# Populated by main.callback() before every command; None if not logged in.
current_token: str | None = None


# ── API helpers ────────────────────────────────────────────────────────────────


def api_base_url() -> str:
    return os.environ.get("OPENDI_API_URL", "http://localhost:8080").rstrip("/")


def auth_headers(content_type: str | None = None) -> dict[str, str]:
    """Build request headers with optional Content-Type and bearer token."""
    headers: dict[str, str] = {}
    if content_type:
        headers["Content-Type"] = content_type
    if current_token:
        headers["Authorization"] = f"Bearer {current_token}"
    return headers


def response_error(response: requests.Response) -> str:
    """Extract a human-readable error string from an API error response."""
    try:
        body = response.json()
        if isinstance(body, dict):
            msg = str(body.get("error") or "")
            if msg:
                return msg
    except (ValueError, TypeError):
        pass
    return response.text or ""


def require_access_token() -> str:
    """Return the current access token or exit with a helpful error."""
    if not current_token:
        typer.echo("Not logged in. Please run `opendi login` first.", err=True)
        raise typer.Exit(1)
    return current_token


# ── Token / username helpers ───────────────────────────────────────────────────


def extract_username_from_token(token: str) -> str | None:
    """Extract the username claim from a JWT payload without signature verification."""
    try:
        parts = token.split(".")
        if len(parts) != 3:
            return None
        payload_b64 = parts[1]
        # Restore base64 padding
        padding = 4 - len(payload_b64) % 4
        if padding != 4:
            payload_b64 += "=" * padding
        payload = json.loads(base64.urlsafe_b64decode(payload_b64))
        return payload.get("username") or payload.get("sub") or None
    except Exception:
        return None


# ── Ref parsing ────────────────────────────────────────────────────────────────


class ParsedRef(NamedTuple):
    owner: str | None   # None when not provided and cannot be inferred
    slug: str | None    # None when repo_id form is used
    tag: str | None     # None when omitted (will be prompted by the command)
    repo_id: int | None  # Non-None only for ``id@<n>`` / ``id@[n]`` hub repo id form


def _parse_id_at_ref(ref: str) -> ParsedRef | None:
    """If ``ref`` uses the ``id@`` hub-repo-id prefix, return ParsedRef; else None.

    Supported:
      id@42           repo id 42, no tag
      id@42:v1        repo id 42, tag v1
      id@[42]:v1      same (bracketed id)
      id@[42]         repo id 42, no tag
    """
    if not ref.startswith("id@"):
        return None
    rest = ref[3:].strip()
    if not rest:
        typer.echo(
            "Invalid ref: use id@<repo_id> or id@<repo_id>:<tag> (e.g. id@42:v1).",
            err=True,
        )
        raise typer.Exit(1)

    tag: str | None = None
    inner: str

    if rest.startswith("["):
        end = rest.find("]")
        if end == -1:
            typer.echo("Invalid ref: missing ']' in id@[repo_id] form.", err=True)
            raise typer.Exit(1)
        inner = rest[1:end].strip()
        tail = rest[end + 1 :].strip()
        if tail.startswith(":"):
            tag = tail[1:].strip() or None
        elif tail:
            typer.echo("Invalid ref: expected id@[repo_id] or id@[repo_id]:<tag>.", err=True)
            raise typer.Exit(1)
    elif ":" in rest:
        id_part, _, tag_part = rest.partition(":")
        inner = id_part.strip()
        tag = tag_part.strip() or None
    else:
        inner = rest.strip()

    if not inner.isdigit():
        typer.echo(
            "Invalid ref: after id@ use digits (e.g. id@42:v1) or id@[42]:v1.",
            err=True,
        )
        raise typer.Exit(1)
    return ParsedRef(owner=None, slug=None, tag=tag, repo_id=int(inner))


def parse_ref(ref: str) -> ParsedRef:
    """Parse a model ref string into its components.

    Supported forms:
      owner/repo:tag  → full canonical form
      repo:tag        → owner defaults to logged-in username
      owner/repo      → tag omitted (command should prompt)
      repo            → both owner and tag defaulted/prompted
      id@42:v1        → hub repo id (uses /v0/repo/:id); also id@[42]:v1
    """
    ref = ref.strip()

    id_parsed = _parse_id_at_ref(ref)
    if id_parsed is not None:
        return id_parsed

    if ":" in ref:
        left, _, tag = ref.rpartition(":")
        tag = tag.strip() or None

        if "/" in left:
            owner, slug = left.split("/", 1)
            return ParsedRef(
                owner=owner.strip() or None,
                slug=slug.strip() or None,
                tag=tag,
                repo_id=None,
            )

        # repo:tag — infer owner from token
        owner = _infer_owner()
        return ParsedRef(owner=owner, slug=left.strip() or None, tag=tag, repo_id=None)

    # No colon — no tag
    if "/" in ref:
        owner, slug = ref.split("/", 1)
        return ParsedRef(
            owner=owner.strip() or None,
            slug=slug.strip() or None,
            tag=None,
            repo_id=None,
        )

    owner = _infer_owner()
    return ParsedRef(owner=owner, slug=ref.strip() or None, tag=None, repo_id=None)


def _infer_owner() -> str | None:
    """Try to get the current username from the active token."""
    if current_token:
        return extract_username_from_token(current_token)
    return None


def require_full_ref(ref: str, label: str = "REF") -> ParsedRef:
    """Parse a ref and abort if owner, slug, and tag are missing (or id@ without tag)."""
    parsed = parse_ref(ref)
    if parsed.repo_id is not None:
        if not parsed.tag:
            typer.echo(
                f"Invalid {label}: tag is required for id@ refs (e.g. id@42:v1).",
                err=True,
            )
            raise typer.Exit(1)
        return parsed
    if not parsed.owner or not parsed.slug or not parsed.tag:
        typer.echo(
            f"Invalid {label}: use owner/repo:tag (e.g. alice/my-model:v1) "
            f"or id@<repo_id>:<tag> (e.g. id@42:v1).",
            err=True,
        )
        raise typer.Exit(1)
    return parsed


def resolve_parsed_ref_to_owner_slug(parsed: ParsedRef, api_base: str) -> tuple[str, str]:
    """Resolve a parsed hub ref to ``(owner, slug)``, calling ``GET /v0/repo/:id`` when needed."""
    if parsed.repo_id is not None:
        try:
            repo_resp = requests.get(
                f"{api_base}/v0/repo/{parsed.repo_id}",
                headers=auth_headers(),
                timeout=10,
            )
        except requests.ConnectionError:
            typer.echo(f"Could not connect to {api_base}.", err=True)
            raise typer.Exit(1)
        except requests.Timeout:
            typer.echo("Request timed out.", err=True)
            raise typer.Exit(1)

        if not repo_resp.ok:
            typer.echo(f"Repository ID {parsed.repo_id} not found.", err=True)
            raise typer.Exit(1)

        repo_data = repo_resp.json()
        return str(repo_data.get("owner", "")), str(repo_data.get("slug", ""))

    if not parsed.owner:
        typer.echo("Could not determine owner. Log in or use owner/repo:tag.", err=True)
        raise typer.Exit(1)
    if not parsed.slug:
        typer.echo("Invalid ref: missing repository slug.", err=True)
        raise typer.Exit(1)
    return parsed.owner, parsed.slug


# ── Typer help strings (hub reference forms) ───────────────────────────────────
# Canonical tag-bearing ref (pull/save/diff/delete tag/delete local/add tag source).
REF_HELP_TAG = (
    "owner/repo:tag, repo:tag when logged in, or id@<repo_id>:<tag> (e.g. id@42:v1)"
)
HELP_ARG_MODEL_REF = f"Hub model ref with tag — {REF_HELP_TAG}."
HELP_ARG_PUSH_REF = f"Upload destination with tag — {REF_HELP_TAG} (same forms as pull)."
HELP_ARG_DELETE_REPO = (
    "Repository to delete — owner/repo, owner/repo:tag (tag ignored), repo or repo:tag when logged in, "
    "id@<repo_id>, or id@<repo_id>:<tag> (tag ignored); deletes the whole repository."
)
HELP_ARG_INSPECT_REF = (
    "Repository or tag — owner/repo; tag: owner/repo:tag, repo:tag when logged in, "
    "id@<repo_id>, or id@<repo_id>:<tag>."
)
HELP_ARG_DIFF_SIDE = f"Remote model ({REF_HELP_TAG}) or path to a local CDM JSON file."
HELP_ARG_ADD_TAG_SOURCE = f"Source tag — {REF_HELP_TAG}."
HELP_ARG_DELETE_TAG = HELP_ARG_MODEL_REF
HELP_ARG_DELETE_LOCAL = HELP_ARG_MODEL_REF


# ── Output helpers ─────────────────────────────────────────────────────────────


def fmt_size(size_bytes: int) -> str:
    """Format a byte count as a human-readable string."""
    for unit in ("B", "KB", "MB", "GB"):
        if size_bytes < 1024:
            return f"{size_bytes:.0f} {unit}"
        size_bytes /= 1024  # type: ignore[assignment]
    return f"{size_bytes:.1f} TB"


def col_widths(rows: list[list[str]]) -> list[int]:
    """Return max column width for each column across all rows."""
    if not rows:
        return []
    return [max(len(r[i]) for r in rows) for i in range(len(rows[0]))]


def print_table(
    headers: list[str],
    rows: list[list[str]],
    col_colors: dict[int, str] | None = None,
    dim_rows: list[bool] | None = None,
) -> None:
    """Print a simple aligned table with a header separator.

    col_colors  — maps column index to a typer color string
    dim_rows    — parallel list to rows; True means print row in dim style
    """
    all_rows = [headers] + rows
    widths = col_widths(all_rows)
    sep = "  "

    header_line = sep.join(h.ljust(widths[i]) for i, h in enumerate(headers))
    typer.echo(typer.style(header_line, bold=True))
    typer.echo(typer.style("-" * len(header_line), dim=True))

    for row_idx, row in enumerate(rows):
        parts = []
        for i, cell in enumerate(row):
            padded = cell.ljust(widths[i])
            if col_colors and i in col_colors:
                padded = typer.style(padded, fg=col_colors[i])
            parts.append(padded)
        line = sep.join(parts)
        if dim_rows and dim_rows[row_idx]:
            line = typer.style(line, dim=True)
        typer.echo(line)
