import json, urllib.request, urllib.error
from dataclasses import dataclass
from .auth import Config

def _request(url, method="GET", body=None, headers=None):
    data = None if body is None else json.dumps(body).encode("utf-8")
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if headers:
        for k, v in headers.items():
            req.add_header(k, v)
    try:
        with urllib.request.urlopen(req) as resp:
            charset = resp.headers.get_content_charset() or "utf-8"
            txt = resp.read().decode(charset)
            return json.loads(txt) if txt else None
    except urllib.error.HTTPError as e:
        try:
            err = json.loads(e.read().decode("utf-8"))
        except Exception:
            err = {"error": e.reason}
        raise SystemExit(f"HTTP {e.code}: {err.get('error', err)}")
    except urllib.error.URLError as e:
        raise SystemExit(f"Network error: {e.reason}")

@dataclass
class ApiClient:
    cfg: Config

    @property
    def base(self):
        return self.cfg.base_url.rstrip("/")

    def _auth_headers(self):
        return {"Authorization": f"Bearer {self.cfg.token}"} if self.cfg.token else {}

    # DEV helper to fetch JWT via /auth/testlogin?id=
    def test_login(self, user_id: int) -> str:
        url = f"{self.base}/auth/testlogin?id={user_id}"
        data = _request(url, "GET")
        token = data.get("token")
        if not token:
            raise SystemExit("Could not obtain token from /auth/testlogin.")
        return token

    # GET /v0/models/:uuid
    def get_model_by_uuid(self, uuid: str):
        url = f"{self.base}/v0/models/{uuid}"
        return _request(url, "GET")

    # PUT /v0/models
    def put_model(self, model_obj: dict):
        url = f"{self.base}/v0/models"
        return _request(url, "PUT", body=model_obj, headers=self._auth_headers())

    # GET /v0/models/lineage/:uuid
    def get_lineage(self, uuid: str):
        url = f"{self.base}/v0/models/lineage/{uuid}"
        return _request(url, "GET")

    # GET /v0/models/children/:uuid
    def get_children(self, uuid: str):
        url = f"{self.base}/v0/models/children/{uuid}"
        return _request(url, "GET")

    # GET /v0/commits/model/:uuid
    def get_commits(self, uuid: str):
        url = f"{self.base}/v0/commits/model/{uuid}"
        return _request(url, "GET")