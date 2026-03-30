"""Stored CLI auth state via the OS credential manager (keyring)."""

import logging

import keyring

logger = logging.getLogger(__name__)

_SERVICE = "opendi"
_ACCOUNT_ACCESS_TOKEN = "access_token"


def store_access_token(token: str) -> None:
    """Persist API bearer token to OS credential manager."""
    keyring.set_password(_SERVICE, _ACCOUNT_ACCESS_TOKEN, token)
    logger.info("Stored access token in OS credential manager")


def load_access_token() -> str | None:
    """Load API bearer token from OS credential manager."""
    return keyring.get_password(_SERVICE, _ACCOUNT_ACCESS_TOKEN)

def delete() -> bool:
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
