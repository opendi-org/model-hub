"""Browser-assisted CLI login against OpenDI API auth endpoints."""

import json
import logging
import time
import urllib.error
import urllib.parse
import urllib.request
import webbrowser

logger = logging.getLogger(__name__)

POLL_INTERVAL_SECONDS = 2


def _json_request(method: str, url: str, payload: dict | None = None) -> dict:
    data = None
    headers = {"Accept": "application/json"}
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
        parsed["_status"] = e.code
        raise RuntimeError(parsed.get("error") or f"HTTP {e.code}") from e


def start_cli_login(api_base_url: str) -> tuple[str, str, int]:
    """Create CLI login session and return (code, login_url, expires_in)."""
    data = _json_request("POST", f"{api_base_url}/v0/auth/cli/login")
    return data["code"], data["loginUrl"], int(data["expiresIn"])


def open_login_url(api_base_url: str, login_url: str) -> str:
    """Open login URL in browser; return absolute URL used."""
    if login_url.startswith("http://") or login_url.startswith("https://"):
        absolute = login_url
    else:
        absolute = f"{api_base_url}{login_url}"
    opened = webbrowser.open(absolute)
    if not opened:
        print(f"Open this URL in your browser to continue login:\n{absolute}")
    return absolute


def poll_cli_token(api_base_url: str, code: str, timeout_seconds: int) -> str:
    """Poll for approved CLI session and return access token."""
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            data = _json_request("POST", f"{api_base_url}/v0/auth/cli/poll", {"code": code})
            if data.get("status") == "pending":
                time.sleep(POLL_INTERVAL_SECONDS)
                continue
            token = data.get("accessToken")
            if token:
                return token
            raise RuntimeError("invalid token response")
        except RuntimeError as e:
            msg = str(e).lower()
            if "pending" in msg:
                time.sleep(POLL_INTERVAL_SECONDS)
                continue
            if "unknown code" in msg or "expired" in msg:
                raise
            raise
    raise TimeoutError("Login timed out before approval.")


def get_current_user(api_base_url: str, token: str) -> dict:
    """Fetch current user using bearer token."""
    req = urllib.request.Request(
        url=f"{api_base_url}/v0/auth/me",
        method="GET",
        headers={"Authorization": f"Bearer {token}", "Accept": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=15) as res:
        body = res.read().decode("utf-8")
        return json.loads(body) if body else {}
