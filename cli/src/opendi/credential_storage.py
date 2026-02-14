"""Stored auth state via the OS credential manager (keyring).

OAuth credentials (Google Credentials object) and the id_token (JWT) are stored
as two separate keyring entries under the same service. The creds entry uses the
library's own to_json() format; the id_token entry stores the raw JWT string
(since the library omits it from to_json / from_authorized_user_info).
"""

import json
import logging

import keyring
from google.oauth2.credentials import Credentials as GoogleCredentials

logger = logging.getLogger(__name__)

_SERVICE = "opendi"
_ACCOUNT_CREDS = "creds"
_ACCOUNT_ID_TOKEN = "id_token"


def store_all(creds: GoogleCredentials) -> None:
    """Persist the full auth state (OAuth creds blob + id_token) to keyring.

    The creds blob is written as the library's to_json() output (no id_token inside).
    When creds.id_token is present, it is written to a second keyring entry.

    Args:
        creds: Google OAuth credentials to store.
    """
    keyring.set_password(_SERVICE, _ACCOUNT_CREDS, creds.to_json())
    if creds.id_token:
        keyring.set_password(_SERVICE, _ACCOUNT_ID_TOKEN, creds.id_token)
    logger.info("Stored credentials written to OS credential manager")


def load_creds() -> GoogleCredentials | None:
    """Load OAuth credentials from the OS credential manager.

    Returns only the Credentials object (no id_token); use load_id_token() for the JWT.

    Returns:
        Google Credentials if found, or None if no stored credentials.
    """
    data = keyring.get_password(_SERVICE, _ACCOUNT_CREDS)
    if data is None:
        return None
    creds = GoogleCredentials.from_authorized_user_info(json.loads(data))
    logger.info("Stored credentials loaded from OS credential manager")
    return creds


def load_id_token() -> str | None:
    """Load the id_token (JWT) from the OS credential manager.

    Returns:
        The JWT string if stored, or None if missing.
    """
    return keyring.get_password(_SERVICE, _ACCOUNT_ID_TOKEN)


def delete() -> bool:
    """Delete stored credentials and id_token from the OS credential manager.

    Returns:
        True if the credentials entry was deleted, False if it was not stored.
    """
    creds_deleted = False
    for account in (_ACCOUNT_CREDS, _ACCOUNT_ID_TOKEN):
        try:
            keyring.delete_password(_SERVICE, account)
            if account == _ACCOUNT_CREDS:
                creds_deleted = True
        except keyring.errors.PasswordDeleteError:
            pass
    if creds_deleted:
        logger.info("Stored credentials deleted from OS credential manager")
    return creds_deleted
