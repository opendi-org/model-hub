"""Minimal tests for the CLI entry point."""

import pytest
from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()


def test_app_help_exits_zero() -> None:
    """opendi --help exits with code 0."""
    result = runner.invoke(app, ["--help"])
    assert result.exit_code == 0
    assert "opendi" in result.output.lower()
    assert "OpenDI" in result.output


def test_app_without_command_exits_zero() -> None:
    """opendi (no subcommand) exits with code 0."""
    result = runner.invoke(app, [])
    assert result.exit_code == 0


def test_login_subcommand() -> None:
    """opendi login runs (placeholder)."""
    result = runner.invoke(app, ["login"])
    assert result.exit_code == 0
    assert "not yet implemented" in result.output.lower()
