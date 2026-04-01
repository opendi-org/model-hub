"""Tests for credential storage (keyring)."""

from unittest.mock import patch

import keyring.errors

from opendi import credential_storage


# ── store_access_token ────────────────────────────────────────────────────────


def test_store_access_token_writes_to_keyring() -> None:
    """store_access_token() writes the token to keyring."""
    with patch("opendi.credential_storage.keyring.set_password") as mock_set:
        credential_storage.store_access_token("my-token")
    mock_set.assert_called_once_with("opendi", "access_token", "my-token")


# ── load_access_token ─────────────────────────────────────────────────────────


def test_load_access_token_returns_token_when_stored() -> None:
    """load_access_token() returns the token when stored."""
    with patch("opendi.credential_storage.keyring.get_password", return_value="my-token"):
        assert credential_storage.load_access_token() == "my-token"


def test_load_access_token_returns_none_when_empty() -> None:
    """load_access_token() returns None when nothing is stored."""
    with patch("opendi.credential_storage.keyring.get_password", return_value=None):
        assert credential_storage.load_access_token() is None


# ── delete ────────────────────────────────────────────────────────────────────


def test_delete_returns_true_when_token_exists() -> None:
    """delete() returns True when the token entry was deleted."""
    with patch("opendi.credential_storage.keyring.delete_password"):
        assert credential_storage.delete() is True


def test_delete_returns_false_when_no_token_stored() -> None:
    """delete() returns False when no token is stored."""
    with patch(
        "opendi.credential_storage.keyring.delete_password",
        side_effect=keyring.errors.PasswordDeleteError(),
    ):
        assert credential_storage.delete() is False
