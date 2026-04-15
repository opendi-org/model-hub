"""Tests for opendi.auth (device-flow login functions)."""

import io
import json
import urllib.error
from unittest.mock import MagicMock, patch

from opendi import auth


# ── Helpers ───────────────────────────────────────────────────────────────────


def _fake_response(body: str):
    """Mock context-manager response for urllib.request.urlopen."""
    m = MagicMock()
    m.__enter__ = lambda s: s
    m.__exit__ = MagicMock(return_value=False)
    m.read.return_value = body.encode()
    return m


def _http_error(code: int, body: str = ""):
    fp = io.BytesIO(body.encode()) if body else None
    return urllib.error.HTTPError(url="http://x", code=code, msg="err", hdrs={}, fp=fp)


# ── _json_request ─────────────────────────────────────────────────────────────


def test_json_request_success() -> None:
    with patch("urllib.request.urlopen", return_value=_fake_response('{"ok": true}')):
        result = auth._json_request("GET", "http://api/endpoint")
    assert result == {"ok": True}


def test_json_request_empty_body() -> None:
    with patch("urllib.request.urlopen", return_value=_fake_response("")):
        result = auth._json_request("GET", "http://api/endpoint")
    assert result == {}


def test_json_request_with_payload() -> None:
    with patch("urllib.request.urlopen", return_value=_fake_response("{}")) as mock_open:
        auth._json_request("POST", "http://api/endpoint", {"key": "val"})
    req = mock_open.call_args[0][0]
    assert req.get_header("Content-type") == "application/json"
    assert b'"key"' in req.data


def test_json_request_http_error_json_body() -> None:
    err = _http_error(400, '{"error": "bad input"}')
    with patch("urllib.request.urlopen", side_effect=err):
        try:
            auth._json_request("POST", "http://api/endpoint")
            assert False, "should have raised"
        except RuntimeError as e:
            assert "bad input" in str(e)


def test_json_request_http_error_non_json_body() -> None:
    err = _http_error(500, "internal server error")
    with patch("urllib.request.urlopen", side_effect=err):
        try:
            auth._json_request("POST", "http://api/endpoint")
            assert False, "should have raised"
        except RuntimeError as e:
            assert "internal server error" in str(e)


def test_json_request_http_error_no_body() -> None:
    err = _http_error(404)
    with patch("urllib.request.urlopen", side_effect=err):
        try:
            auth._json_request("GET", "http://api/endpoint")
            assert False, "should have raised"
        except RuntimeError as e:
            assert "404" in str(e)


# ── start_cli_login ───────────────────────────────────────────────────────────


def test_start_cli_login_returns_tuple() -> None:
    body = json.dumps({"code": "abc", "loginUrl": "/login", "expiresIn": 300})
    with patch("urllib.request.urlopen", return_value=_fake_response(body)):
        code, url, exp = auth.start_cli_login("http://api")
    assert code == "abc"
    assert url == "/login"
    assert exp == 300


# ── open_login_url ────────────────────────────────────────────────────────────


def test_open_login_url_relative() -> None:
    with patch("webbrowser.open", return_value=True) as mock_open:
        result = auth.open_login_url("http://api", "/login/abc")
    assert result == "http://api/login/abc"
    mock_open.assert_called_once_with("http://api/login/abc")


def test_open_login_url_absolute() -> None:
    with patch("webbrowser.open", return_value=True):
        result = auth.open_login_url("http://api", "https://other.com/login")
    assert result == "https://other.com/login"


def test_open_login_url_browser_fails(capsys) -> None:
    with patch("webbrowser.open", return_value=False):
        auth.open_login_url("http://api", "/login/abc")
    assert "http://api/login/abc" in capsys.readouterr().out


# ── poll_cli_token ────────────────────────────────────────────────────────────


def test_poll_cli_token_immediate() -> None:
    body = json.dumps({"accessToken": "tok123"})
    with patch("urllib.request.urlopen", return_value=_fake_response(body)):
        token = auth.poll_cli_token("http://api", "code", 30)
    assert token == "tok123"


def test_poll_cli_token_pending_then_success() -> None:
    pending = json.dumps({"status": "pending"})
    success = json.dumps({"accessToken": "tok456"})
    with (
        patch("urllib.request.urlopen", side_effect=[_fake_response(pending), _fake_response(success)]),
        patch("time.sleep"),
    ):
        token = auth.poll_cli_token("http://api", "code", 30)
    assert token == "tok456"


def test_poll_cli_token_timeout() -> None:
    with (
        patch("time.time", side_effect=[0, 0, 999]),
        patch("urllib.request.urlopen", return_value=_fake_response(json.dumps({"status": "pending"}))),
    ):
        try:
            auth.poll_cli_token("http://api", "code", 1)
            assert False, "should have raised"
        except TimeoutError:
            pass


def test_poll_cli_token_unknown_code_raises() -> None:
    err = _http_error(400, '{"error": "unknown code"}')
    with patch("urllib.request.urlopen", side_effect=err):
        try:
            auth.poll_cli_token("http://api", "code", 30)
            assert False, "should have raised"
        except RuntimeError as e:
            assert "unknown code" in str(e)


# ── get_current_user ──────────────────────────────────────────────────────────


def test_get_current_user_returns_dict() -> None:
    body = json.dumps({"username": "alice", "email": "alice@example.com"})
    with patch("urllib.request.urlopen", return_value=_fake_response(body)):
        user = auth.get_current_user("http://api", "tok")
    assert user["username"] == "alice"


def test_get_current_user_empty_body() -> None:
    with patch("urllib.request.urlopen", return_value=_fake_response("")):
        user = auth.get_current_user("http://api", "tok")
    assert user == {}
