"""Tests for opendi.cmds.shared (ref parsing)."""

import pytest
import typer

from opendi.cmds import shared


def test_parse_ref_id_at_colon_form() -> None:
    r = shared.parse_ref("id@99:prod")
    assert r.repo_id == 99
    assert r.tag == "prod"
    assert r.owner is None and r.slug is None


def test_parse_ref_id_at_bracket_form() -> None:
    r = shared.parse_ref("id@[12]:v0")
    assert r.repo_id == 12
    assert r.tag == "v0"


def test_parse_ref_id_at_no_tag() -> None:
    r = shared.parse_ref("id@5")
    assert r.repo_id == 5
    assert r.tag is None


def test_parse_ref_digits_colon_is_not_repo_id() -> None:
    """``42:v1`` is repo slug ``42``, not hub id 42 (use ``id@42:v1``)."""
    shared.current_token = None
    r = shared.parse_ref("42:v1")
    assert r.repo_id is None
    assert r.slug == "42"
    assert r.tag == "v1"
    assert r.owner is None


def test_parse_ref_id_at_invalid_id_exits() -> None:
    with pytest.raises(typer.Exit):
        shared.parse_ref("id@abc:v1")


def test_require_full_ref_id_at_requires_tag() -> None:
    with pytest.raises(typer.Exit):
        shared.require_full_ref("id@3", label="REF")


def test_wrap_text_block_paragraph_gap() -> None:
    lines = shared.wrap_text_block("first block.\n\nsecond here.", width=40)
    assert "first" in lines[0]
    assert "" in lines
    assert any("second" in ln for ln in lines)


def test_wrap_text_block_long_token() -> None:
    long = "x" * 50
    lines = shared.wrap_text_block(long, width=20)
    assert len(lines) >= 2
    assert "".join(lines) == long


def test_repo_table_name_style() -> None:
    assert shared.repo_table_name_style("alice", "alice", "private") == "bold green"
    assert shared.repo_table_name_style(None, "bob", "public") == "white"
    assert shared.repo_table_name_style("alice", "bob", "private") == ""


def test_visible_text_width_strips_ansi() -> None:
    import typer

    styled = typer.style("abc", fg=typer.colors.GREEN, bold=True)
    assert shared._visible_text_width(styled) == 3
    assert shared._visible_text_width("") == 0


def test_col_widths_uses_visible_length_for_styled_cells() -> None:
    import typer

    styled = typer.style("longtag", fg=typer.colors.GREEN, bold=True)
    widths = shared.col_widths(
        [
            ["TAG", "DIGEST"],
            [styled, "deadbeef"],
        ]
    )
    assert widths[0] == max(len("TAG"), len("longtag"))
    assert widths[1] == len("deadbeef")
