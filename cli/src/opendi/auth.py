"""OAuth login with local loopback (Google)."""

import logging
import threading
import time
from pathlib import Path

from google_auth_oauthlib.flow import InstalledAppFlow
from google.oauth2.credentials import Credentials as GoogleCredentials

logger = logging.getLogger(__name__)

SCOPES = [
    "openid",
    "https://www.googleapis.com/auth/userinfo.email",
    "https://www.googleapis.com/auth/userinfo.profile",
]

# Time allowed for user to sign in in the browser (password, 2FA, etc.). No redirect = timeout.
LOGIN_TIMEOUT_SECONDS = 120

# Poll interval so the main thread can respond to KeyboardInterrupt (Windows blocks in socket otherwise).
_POLL_INTERVAL = 0.3

# User-facing strings for the OAuth flow (library prints these).
# OSC 8 hyperlink: clickable "link" text, URL in the escape payload (no long URL in terminal).
_AUTH_PROMPT_MESSAGE = (
    "Opening your browser to sign in with Google...\n"
    "If it didn't open, visit this \033[36m\033]8;;{url}\033\\link\033]8;;\033\\\033[0m."
)
_SUCCESS_MESSAGE = "Login successful. You can close this tab."


def run_login_flow(client_secrets_path: Path) -> GoogleCredentials:
    """Run OAuth with local loopback; return Google OAuth credentials (caller persists if needed).

    The blocking server runs in a daemon thread so the main thread can respond
    to Ctrl+C (on Windows the socket block otherwise ignores KeyboardInterrupt).

    Args:
        client_secrets_path: Path to a GCP client_secret.json file.

    Returns:
        Google OAuth credentials from the completed OAuth flow.

    Raises:
        FileNotFoundError: If the path does not exist or cannot be read (e.g. from_client_secrets_file).
        AccessDeniedError: If the user denies consent (OAuth redirect with error=access_denied).
        OAuth2Error: Other OAuth errors from the provider or token exchange.
        AttributeError: If no redirect is received within the timeout (user closed tab or did not complete).

    Note:
        Closing the browser tab without clicking Allow/Cancel cannot be detected; we only get
        a request when Google redirects to our callback. We rely on a timeout (LOGIN_TIMEOUT_SECONDS).
    """
    logger.info("Starting OAuth flow (local loopback)")
    flow = InstalledAppFlow.from_client_secrets_file(
        str(client_secrets_path),
        scopes=SCOPES,
    )
    result_holder: list[GoogleCredentials | None] = []
    exc_holder: list[BaseException] = []

    def run_server() -> None:
        try:
            creds = flow.run_local_server(
                port=0,
                host="localhost",
                open_browser=True,
                success_message=_SUCCESS_MESSAGE,
                timeout_seconds=LOGIN_TIMEOUT_SECONDS,
                authorization_prompt_message=_AUTH_PROMPT_MESSAGE,
            )
            result_holder.append(creds)
        except BaseException as e:
            logger.debug("OAuth flow failed: %s", e, exc_info=True)
            exc_holder.append(e)

    thread = threading.Thread(target=run_server, daemon=True)
    thread.start()
    # Poll so main thread can receive KeyboardInterrupt; daemon thread exits when process exits.
    while True:
        if result_holder:
            logger.info("OAuth flow completed")
            return result_holder[0]
        if exc_holder:
            raise exc_holder[0]
        time.sleep(_POLL_INTERVAL)
