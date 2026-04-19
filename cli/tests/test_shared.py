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
