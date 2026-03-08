"""Tests for credential storage (keyring) — two-entry model."""

import json
from unittest.mock import MagicMock, patch

import keyring.errors

from opendi import credential_storage


def _make_creds_json() -> str:
    """Return a minimal Google credentials JSON string."""
    return json.dumps(
        {
            "token": "access_tok",
            "refresh_token": "refresh_tok",
            "token_uri": "https://oauth2.googleapis.com/token",
            "client_id": "id.apps.googleusercontent.com",
            "client_secret": "secret",
            "scopes": ["openid"],
        }
    )


# ── store_all ─────────────────────────────────────────────────────────────────


def test_store_all_writes_creds_only_when_no_id_token() -> None:
    """store_all() writes only the creds entry when id_token is absent."""
    creds = MagicMock()
    creds.to_json.return_value = '{"token":"t"}'
    creds.id_token = None
    with patch("opendi.credential_storage.keyring.set_password") as mock_set:
        credential_storage.store_all(creds)
    mock_set.assert_called_once_with("opendi", "creds", '{"token":"t"}')


def test_store_all_writes_both_entries_when_id_token_present() -> None:
    """store_all() writes creds and id_token as two separate keyring entries."""
    creds = MagicMock()
    creds.to_json.return_value = '{"token":"t"}'
    creds.id_token = "fake.jwt.token"
    with patch("opendi.credential_storage.keyring.set_password") as mock_set:
        credential_storage.store_all(creds)
    assert mock_set.call_count == 2
    mock_set.assert_any_call("opendi", "creds", '{"token":"t"}')
    mock_set.assert_any_call("opendi", "id_token", "fake.jwt.token")


def test_store_all_creds_blob_has_no_id_token() -> None:
    """store_all() does not inject id_token into the creds JSON blob."""
    creds = MagicMock()
    creds.to_json.return_value = '{"token":"t"}'
    creds.id_token = "fake.jwt.token"
    with patch("opendi.credential_storage.keyring.set_password") as mock_set:
        credential_storage.store_all(creds)
    # First call is the creds entry; verify blob has no id_token.
    creds_blob = json.loads(mock_set.call_args_list[0][0][2])
    assert "id_token" not in creds_blob


# ── load_creds ────────────────────────────────────────────────────────────────


def test_load_creds_returns_credentials_when_stored() -> None:
    """load_creds() returns Credentials when data exists in keyring."""
    data = _make_creds_json()
    with patch("opendi.credential_storage.keyring.get_password", return_value=data):
        creds = credential_storage.load_creds()
    assert creds is not None
    assert creds.token == "access_tok"
    assert creds.refresh_token == "refresh_tok"


def test_load_creds_returns_none_when_empty() -> None:
    """load_creds() returns None when nothing is stored."""
    with patch("opendi.credential_storage.keyring.get_password", return_value=None):
        creds = credential_storage.load_creds()
    assert creds is None


def test_load_creds_does_not_set_id_token() -> None:
    """load_creds() does not populate id_token on the Credentials object."""
    data = _make_creds_json()
    with patch("opendi.credential_storage.keyring.get_password", return_value=data):
        creds = credential_storage.load_creds()
    assert creds is not None
    assert creds.id_token is None


# ── load_id_token ─────────────────────────────────────────────────────────────


def test_load_id_token_returns_jwt_when_stored() -> None:
    """load_id_token() returns the JWT string from keyring."""
    with patch("opendi.credential_storage.keyring.get_password", return_value="fake.jwt.token"):
        assert credential_storage.load_id_token() == "fake.jwt.token"


def test_load_id_token_returns_none_when_empty() -> None:
    """load_id_token() returns None when nothing is stored."""
    with patch("opendi.credential_storage.keyring.get_password", return_value=None):
        assert credential_storage.load_id_token() is None


# ── delete ────────────────────────────────────────────────────────────────────


def test_delete_returns_true_and_deletes_both_entries() -> None:
    """delete() returns True and deletes both keyring entries."""
    with patch("opendi.credential_storage.keyring.delete_password") as mock_del:
        assert credential_storage.delete() is True
    assert mock_del.call_count == 2
    mock_del.assert_any_call("opendi", "creds")
    mock_del.assert_any_call("opendi", "id_token")


def test_delete_returns_false_when_no_creds_stored() -> None:
    """delete() returns False when the main creds entry does not exist."""
    with patch(
        "opendi.credential_storage.keyring.delete_password",
        side_effect=keyring.errors.PasswordDeleteError(),
    ):
        assert credential_storage.delete() is False


def test_delete_returns_true_when_only_id_token_missing() -> None:
    """delete() returns True when creds exist but id_token entry is missing."""
    def side_effect(service: str, account: str) -> None:
        if account == "id_token":
            raise keyring.errors.PasswordDeleteError()

    with patch("opendi.credential_storage.keyring.delete_password", side_effect=side_effect):
        assert credential_storage.delete() is True
