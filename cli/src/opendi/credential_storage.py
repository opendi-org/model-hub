"""Stored CLI auth state via the OS credential manager (keyring)."""

import logging

import keyring
import keyring.errors

logger = logging.getLogger(__name__)

_SERVICE = "opendi"
_ACCOUNT_ACCESS_TOKEN = "access_token"


def store_access_token(token: str) -> None:
    """Persist API bearer token to OS credential manager."""
    try:
        keyring.set_password(_SERVICE, _ACCOUNT_ACCESS_TOKEN, token)
    except keyring.errors.KeyringError as e:
        raise RuntimeError(f"Failed to store access token in OS credential manager: {e}") from e
    logger.info("Stored access token in OS credential manager")


def load_access_token() -> str | None:
    """Load API bearer token from OS credential manager."""
    try:
        token = keyring.get_password(_SERVICE, _ACCOUNT_ACCESS_TOKEN)
    except keyring.errors.KeyringError as e:
        raise RuntimeError(f"Failed to load access token from OS credential manager: {e}") from e
    logger.debug("Access token %s in OS credential manager", "found" if token else "not found")
    return token


def delete_access_token() -> bool:
    """Delete stored access token from the OS credential manager.

    Returns:
        True if token was deleted, False if it was not stored.
    """
    try:
        keyring.delete_password(_SERVICE, _ACCOUNT_ACCESS_TOKEN)
        logger.info("Stored access token deleted from OS credential manager")
        return True
    except keyring.errors.PasswordDeleteError:
        return False
