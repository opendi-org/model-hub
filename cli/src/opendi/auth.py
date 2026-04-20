"""Browser-assisted CLI login against OpenDI API auth endpoints."""

import json
import logging
import time
import urllib.error
import urllib.request
import webbrowser

logger = logging.getLogger(__name__)

POLL_INTERVAL_SECONDS = 2


class AuthAPIError(RuntimeError):
    """HTTP error from the OpenDI API, carrying the response status code."""

    def __init__(self, message: str, status: int) -> None:
        super().__init__(message)
        self.status = status


def _json_request(
    method: str,
    url: str,
    payload: dict | None = None,
    extra_headers: dict | None = None,
) -> dict:
    data = None
    headers = {"Accept": "application/json"}
    if extra_headers:
        headers.update(extra_headers)
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url=url, method=method, data=data, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as res:
            body = res.read().decode("utf-8")
            return json.loads(body) if body else {}
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8") if e.fp else ""
        parsed: dict = {}
        if raw:
            try:
                parsed = json.loads(raw)
            except json.JSONDecodeError:
                parsed = {"error": raw}
        message = parsed.get("error") or f"HTTP {e.code}"
        raise AuthAPIError(message, e.code) from e


def start_cli_login(api_base_url: str) -> tuple[str, str, int]:
    """Create CLI login session and return (code, login_url, expires_in)."""
    logger.debug("Starting CLI login session at %s", api_base_url)
    data = _json_request("POST", f"{api_base_url}/v0/auth/cli/login")
    try:
        return data["code"], data["loginUrl"], int(data["expiresIn"])
    except (KeyError, TypeError, ValueError) as e:
        raise RuntimeError(f"Unexpected login response from server: {data}") from e


def open_login_url(api_base_url: str, login_url: str) -> str:
    """Open login URL in browser; return absolute URL used."""
    if login_url.startswith("http://") or login_url.startswith("https://"):
        absolute = login_url
    else:
        absolute = f"{api_base_url}{login_url}"
    logger.debug("Opening login URL: %s", absolute)
    opened = webbrowser.open(absolute)
    if not opened:
        logger.warning("Could not open a web browser automatically.")
    return absolute


def poll_cli_token(api_base_url: str, code: str, timeout_seconds: int) -> str:
    """Poll for approved CLI session and return access token."""
    deadline = time.time() + timeout_seconds
    attempt = 0
    while time.time() < deadline:
        attempt += 1
        logger.debug("Polling for CLI token (attempt %d)", attempt)
        data = _json_request("POST", f"{api_base_url}/v0/auth/cli/poll", {"code": code})
        if data.get("status") == "pending":
            time.sleep(POLL_INTERVAL_SECONDS)
            continue
        token = data.get("accessToken")
        if token:
            logger.info("CLI token acquired")
            return token
        raise RuntimeError("Unexpected token response from server")
    raise TimeoutError("Login timed out before approval.")


def get_current_user(api_base_url: str, token: str) -> dict:
    """Fetch current user using bearer token."""
    logger.debug("Fetching current user")
    return _json_request(
        "GET",
        f"{api_base_url}/v0/auth/me",
        extra_headers={"Authorization": f"Bearer {token}"},
    )
