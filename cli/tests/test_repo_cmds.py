"""Tests for repo subgroup commands: inspect, create repo, delete repo, list repos."""

from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from opendi.main import app

runner = CliRunner()

_TOKEN = "fake-access-token"


def _logged_in():
    return patch("opendi.main.credential_storage.load_access_token", return_value=_TOKEN)


def _logged_out():
    return patch("opendi.main.credential_storage.load_access_token", return_value=None)


def _ok(data: dict, status: int = 200):
    m = MagicMock()
    m.status_code = status
    m.ok = status < 400
    m.json.return_value = data
    m.text = ""
    return m


def _err(status: int, error: str = ""):
    m = MagicMock()
    m.status_code = status
    m.ok = False
    m.json.return_value = {"error": error} if error else {}
    m.text = error
    return m


_REPO_DATA = {
    "owner": "alice",
    "slug": "my-repo",
    "visibility": "public",
    "description": "A repo",
    "updatedAt": "2026-01-01T00:00:00Z",
    "tags": [
        {
            "name": "v1",
            "digest": "abc123def456",
            "size": 4096,
            "updatedAt": "2026-01-01T00:00:00Z",
            "createdBy": "alice",
        }
    ],
}


# ── inspect (unified repo / tag) ──────────────────────────────────────────────


def test_inspect_repo_success() -> None:
    """opendi inspect REF prints repo name, visibility, and tags when REF has no tag."""
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(_REPO_DATA)),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-repo"])
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output
    assert "public" in result.output
    assert "v1" in result.output


def test_inspect_repo_verbose() -> None:
    """opendi inspect REF --verbose shows tag digest table for a repository ref."""
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(_REPO_DATA)),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-repo", "--verbose"])
    assert result.exit_code == 0
    assert "abc123" in result.output


def test_inspect_repo_not_found() -> None:
    """opendi inspect exits 1 on 404 for a repository ref."""
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(404)),
    ):
        result = runner.invoke(app, ["inspect", "alice/no-repo"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_inspect_repo_access_denied() -> None:
    """opendi inspect exits 1 on 401/403."""
    for status in (401, 403):
        with (
            _logged_out(),
            patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(status)),
        ):
            result = runner.invoke(app, ["inspect", "alice/my-repo"])
        assert result.exit_code == 1
        assert "Access denied" in result.output


def test_inspect_repo_needs_owner_when_slug_only() -> None:
    """opendi inspect my-repo without login cannot infer owner."""
    with _logged_out():
        result = runner.invoke(app, ["inspect", "my-repo"])
    assert result.exit_code == 1
    assert "owner/repo" in result.output.lower() or "log in" in result.output.lower()


def test_inspect_repo_connection_error() -> None:
    """opendi inspect exits 1 when hub is unreachable."""
    with (
        _logged_out(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-repo"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_inspect_repo_timeout() -> None:
    """opendi inspect exits 1 when request times out."""
    with (
        _logged_out(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").Timeout(),
        ),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-repo"])
    assert result.exit_code == 1
    assert "timed out" in result.output


_TAG_REPO_DATA = {
    "id": 42,
    "owner": "alice",
    "slug": "my-model",
    "visibility": "public",
    "tags": [
        {
            "name": "v1",
            "digest": "abc123def456",
            "size": 4096,
            "updatedAt": "2026-01-01T00:00:00Z",
            "createdBy": "alice",
        }
    ],
}


def test_inspect_tag_success() -> None:
    """opendi inspect REF prints tag metadata when REF includes a tag."""
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(_TAG_REPO_DATA)),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-model:v1"])
    assert result.exit_code == 0
    assert "alice/my-model" in result.output
    assert "v1" in result.output
    assert "abc123" in result.output
    assert "alice" in result.output


def test_inspect_tag_not_found() -> None:
    """opendi inspect exits 1 when the tag does not exist in the repo."""
    data = dict(_TAG_REPO_DATA)
    data["tags"] = []
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(data)),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-model:missing"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_inspect_tag_repo_not_found() -> None:
    """opendi inspect exits 1 on 404 when resolving the repository for a tag ref."""
    with (
        _logged_out(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(404)),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-model:v1"])
    assert result.exit_code == 1
    assert "not found" in result.output.lower()


def test_inspect_tag_access_denied() -> None:
    """opendi inspect exits 1 on 401/403 for a tag ref."""
    for status in (401, 403):
        with (
            _logged_out(),
            patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(status)),
        ):
            result = runner.invoke(app, ["inspect", "alice/my-model:v1"])
        assert result.exit_code == 1
        assert "Access denied" in result.output


def test_inspect_tag_connection_error() -> None:
    """opendi inspect exits 1 when hub is unreachable (tag ref)."""
    with (
        _logged_out(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-model:v1"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_inspect_tag_timeout() -> None:
    """opendi inspect exits 1 when request times out (tag ref)."""
    with (
        _logged_out(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").Timeout(),
        ),
    ):
        result = runner.invoke(app, ["inspect", "alice/my-model:v1"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── create repo ───────────────────────────────────────────────────────────────


def test_create_repo_success_private() -> None:
    """opendi create repo posts to the API and prints success (private by default)."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_ok({}, 201)) as mock_post,
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 0
    assert "my-repo" in result.output
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["slug"] == "my-repo"
    assert kwargs["json"]["visibility"] == "private"


def test_create_repo_success_public() -> None:
    """opendi create repo creates a public repo when --public is passed."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_ok({}, 201)) as mock_post,
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo", "--public"])
    assert result.exit_code == 0
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["visibility"] == "public"


def test_create_repo_success_with_description() -> None:
    """opendi create repo forwards --description to the API."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_ok({}, 201)) as mock_post,
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo", "--description", "A test repo"])
    assert result.exit_code == 0
    _, kwargs = mock_post.call_args
    assert kwargs["json"]["description"] == "A test repo"


def test_create_repo_requires_login() -> None:
    """opendi create repo exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_create_repo_conflict() -> None:
    """opendi create repo exits 1 when the repo name already exists (409)."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_err(409)),
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "already exists" in result.output


def test_create_repo_unauthorized() -> None:
    """opendi create repo exits 1 with auth message on 401."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_err(401)),
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_create_repo_unexpected_status() -> None:
    """opendi create repo exits 1 with HTTP status on unexpected response."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.post", return_value=_err(500)),
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_create_repo_connection_error() -> None:
    """opendi create repo exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.post",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_create_repo_timeout() -> None:
    """opendi create repo exits 1 when the request times out."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.post",
            side_effect=__import__("requests").Timeout(),
        ),
    ):
        result = runner.invoke(app, ["create", "repo", "my-repo"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── delete repo ───────────────────────────────────────────────────────────────


def test_delete_repo_success() -> None:
    """opendi delete repo deletes the repo and prints success (HTTP 204)."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.delete", return_value=_ok({}, 204)) as mock_del,
        patch("opendi.cmds.repo_cmds.local_cache.mark_stale_by_repo", return_value=0),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output
    args, _ = mock_del.call_args
    assert "alice/my-repo" in args[0]


def test_delete_repo_prompts_without_yes() -> None:
    """opendi delete repo prompts for confirmation when --yes is not passed."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.delete", return_value=_ok({}, 204)),
        patch("opendi.cmds.repo_cmds.local_cache.mark_stale_by_repo", return_value=0),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo"], input="y\n")
    assert result.exit_code == 0
    assert "alice/my-repo" in result.output


def test_delete_repo_aborts_on_decline() -> None:
    """opendi delete repo aborts when user declines the confirmation prompt."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.delete") as mock_del,
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo"], input="n\n")
    assert result.exit_code != 0
    mock_del.assert_not_called()


def test_delete_repo_requires_login() -> None:
    """opendi delete repo exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_repo_invalid_format() -> None:
    """opendi delete repo exits 1 when ref cannot identify owner and slug (or id@)."""
    with _logged_in():
        result = runner.invoke(app, ["delete", "repo", "alice/", "--yes"])
    assert result.exit_code == 1
    assert "Invalid ref" in result.output


def test_delete_repo_resolves_id_at_ref() -> None:
    """opendi delete repo resolves id@repo_id:tag to owner/slug before DELETE."""
    get_resp = _ok({"owner": "alice", "slug": "resolved-slug", "id": 9, "tags": []}, 200)
    with (
        _logged_in(),
        patch("opendi.cmds.shared.requests.get", return_value=get_resp),
        patch("opendi.cmds.repo_cmds.requests.delete", return_value=_ok({}, 204)) as mock_del,
        patch("opendi.cmds.repo_cmds.local_cache.mark_stale_by_repo", return_value=0),
    ):
        result = runner.invoke(app, ["delete", "repo", "id@9:v1", "--yes"])
    assert result.exit_code == 0
    assert "alice/resolved-slug" in result.output
    del_url = mock_del.call_args[0][0]
    assert "alice" in del_url and "resolved-slug" in del_url


def test_delete_repo_not_found_or_forbidden() -> None:
    """opendi delete repo exits 1 when repo does not exist or user is not owner."""
    for status in (403, 404):
        with (
            _logged_in(),
            patch("opendi.cmds.repo_cmds.requests.delete", return_value=_err(status)),
        ):
            result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
        assert result.exit_code == 1
        assert "does not exist or you are not the owner" in result.output


def test_delete_repo_unauthorized() -> None:
    """opendi delete repo exits 1 with auth message on 401."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.delete", return_value=_err(401)),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_delete_repo_unexpected_status() -> None:
    """opendi delete repo exits 1 with HTTP status on unexpected response."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.delete", return_value=_err(500)),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_delete_repo_connection_error() -> None:
    """opendi delete repo exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.delete",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_delete_repo_timeout() -> None:
    """opendi delete repo exits 1 when the request times out."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.delete",
            side_effect=__import__("requests").Timeout(),
        ),
    ):
        result = runner.invoke(app, ["delete", "repo", "alice/my-repo", "--yes"])
    assert result.exit_code == 1
    assert "timed out" in result.output


# ── list repos ────────────────────────────────────────────────────────────────

_LIST_RESPONSE = {
    "repositories": [
        {"owner": "alice", "slug": "repo-a", "visibility": "public", "description": "First"},
        {"owner": "alice", "slug": "repo-b", "visibility": "private", "description": ""},
    ]
}


def test_list_repos_requires_login() -> None:
    """opendi list repos exits 1 when not logged in."""
    with _logged_out():
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_list_repos_success() -> None:
    """opendi list repos prints all repos."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(_LIST_RESPONSE)),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 0
    assert "repo-a" in result.output
    assert "repo-b" in result.output


def test_list_repos_by_owner() -> None:
    """opendi list repos sends owner param when given."""
    resp = _ok({"repositories": [{"owner": "alice", "slug": "repo-a", "visibility": "public", "description": ""}]})
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=resp) as mock_get,
    ):
        result = runner.invoke(app, ["list", "repos", "alice"])
    assert result.exit_code == 0
    _, kwargs = mock_get.call_args
    assert kwargs["params"].get("owner") == "alice"


def test_list_repos_all_flag() -> None:
    """opendi list repos --all passes scope=all to API."""
    resp = _ok(_LIST_RESPONSE)
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=resp) as mock_get,
    ):
        result = runner.invoke(app, ["list", "repos", "--all"])
    assert result.exit_code == 0
    _, kwargs = mock_get.call_args
    assert kwargs["params"].get("scope") == "all"


def test_list_repos_limit() -> None:
    """opendi list repos --limit N shows at most N repos."""
    many = {"repositories": [
        {"owner": "alice", "slug": f"repo-{i}", "visibility": "public", "description": ""}
        for i in range(10)
    ]}
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok(many)),
    ):
        result = runner.invoke(app, ["list", "repos", "--limit", "3"])
    assert result.exit_code == 0
    assert result.output.count("alice/repo-") == 3


def test_list_repos_empty() -> None:
    """opendi list repos prints 'No repositories found' when API returns empty list."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_ok({"repositories": []})),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 0
    assert "No repositories found" in result.output


def test_list_repos_owner_not_found() -> None:
    """opendi list repos exits 1 when the owner does not exist (404)."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(404)),
    ):
        result = runner.invoke(app, ["list", "repos", "alice"])
    assert result.exit_code == 1
    assert "alice" in result.output


def test_list_repos_unauthorized() -> None:
    """opendi list repos exits 1 with auth message on 401."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(401)),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 1
    assert "opendi login" in result.output


def test_list_repos_unexpected_status() -> None:
    """opendi list repos exits 1 with HTTP status on unexpected response."""
    with (
        _logged_in(),
        patch("opendi.cmds.repo_cmds.requests.get", return_value=_err(500)),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 1
    assert "HTTP 500" in result.output


def test_list_repos_connection_error() -> None:
    """opendi list repos exits 1 when the hub is unreachable."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").ConnectionError(),
        ),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 1
    assert "Could not connect" in result.output


def test_list_repos_timeout() -> None:
    """opendi list repos exits 1 when the request times out."""
    with (
        _logged_in(),
        patch(
            "opendi.cmds.repo_cmds.requests.get",
            side_effect=__import__("requests").Timeout(),
        ),
    ):
        result = runner.invoke(app, ["list", "repos"])
    assert result.exit_code == 1
    assert "timed out" in result.output
