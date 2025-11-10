"""
OpenDI Model Hub CLI 

Commands:
  cli set-url "url"
  cli set-token "token"
  cli pull "tag"
  cli push "tag"
  cli commit "tag"
  cli init "path"
  cli get-commits -r "tag"
  cli get-lineage -r "tag"
  cli get-models
  cli clear-token

Config location:
  <repo_root>/.opendi_cli/config.json

Run examples (from repo root):
  py -m cli -h
  py -m cli set-url http://localhost:8080
  py -m cli set-token "<JWT>"
  py -m cli clear-token
"""

import argparse
import json
import sys
from pathlib import Path
from urllib.parse import urlparse

# 3rd-party HTTP lib (used later for API calls)
try:
    import requests  # noqa: F401
except Exception:
    sys.exit("Error: this CLI needs 'requests'. Install with: py -m pip install requests")

# -----------------------------------------------------------------------------
# Config management
# -----------------------------------------------------------------------------
CONFIG_DIR = Path(__file__).resolve().parent / ".opendi_cli"
CONFIG_PATH = CONFIG_DIR / "config.json"
MAPPINGS_PATH = CONFIG_DIR / "mapping.json"
DEFAULT_REMOTE_URL = "http://opendi-modelhub.org" # temporary
DEFAULT_ENGINE_URL = "http://localhost:7070"

def load_config() -> dict:
    """Load config JSON (or return {} if missing/corrupt)."""
    if not CONFIG_PATH.exists():
        return {}
    try:
        return json.loads(CONFIG_PATH.read_text(encoding="utf-8"))
    except Exception:
        return {}

def save_config(cfg: dict) -> None:
    """Persist config JSON to disk (inside repo)."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    CONFIG_PATH.write_text(json.dumps(cfg, indent=2), encoding="utf-8")

def load_mappings() -> list:
    """Load mappings JSON (or return [] if missing/corrupt)."""
    if not MAPPINGS_PATH.exists():
        return []
    try:
        return json.loads(MAPPINGS_PATH.read_text(encoding="utf-8"))
    except Exception:
        return []

def save_mappings(map: list) -> None:
    """Persist mappings JSON to disk (inside repo)."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    MAPPINGS_PATH.write_text(json.dumps(map, indent=2), encoding="utf-8")

def require_remote_url(cfg: dict) -> str:
    return (cfg.get("remote_url") or "").strip() or DEFAULT_REMOTE_URL

def require_token(cfg: dict) -> str:
    tok = (cfg.get("token") or "").strip()
    if not tok:
        sys.exit("Error: not logged in. Run: py -m cli set-token <JWT>")
    return tok

# -----------------------------------------------------------------------------
# Implemented commands
# -----------------------------------------------------------------------------
def cmd_set_url(args, cfg):
    """Save/override the remote API URL."""
    url = args.url.strip()
    parsed = urlparse(url)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        sys.exit("Error: invalid URL. Example: http://localhost:8080")
    cfg["remote_url"] = url
    save_config(cfg)
    print(f"Remote URL set to {url}")
    print(f"(Saved at {CONFIG_PATH})")

def cmd_set_token(args, cfg):
    """Save/replace the auth token (login)."""
    token = args.token.strip()
    if not token:
        sys.exit("Error: token cannot be empty.")
    cfg["token"] = token
    save_config(cfg)
    print("Token saved")
    print(f"(Saved at {CONFIG_PATH})")

def cmd_clear_token(args, cfg):
    """Clear the auth token (logout)."""
    cfg["token"] = ""
    save_config(cfg)
    print("Token cleared")
    print(f"(Cleared from: {CONFIG_PATH})")

# -----------------------------------------------------------------------------
# Stubs (help pages exist; implementation still needs to be added)
# -----------------------------------------------------------------------------
def cmd_pull(args, cfg):
    print(f"[TODO] pull tag={args.tag}")

def cmd_push(args, cfg):
    require_token(cfg)
    print(f"[TODO] push tag={args.tag}")

def cmd_commit(args, cfg):
    print(f"[TODO] commit tag={args.tag}")

def cmd_init(args, cfg):
    path = Path(args.path)
    try:
      model_json = json.loads(path.read_text(encoding="utf-8"))
      response = requests.post(f"{DEFAULT_ENGINE_URL}/v0/models", json=model_json)
      if response.status_code == 200:
          model = response.json()
          model_tag = model.get("addons").get("tag") # will not work until tag is returned from API
          mappings = [m for m in load_mappings() if m.get("tag") != model_tag] 
          mappings.append({"tag": model_tag, "path": Path(__file__).resolve().parent, "remote": ""})
          print("Model initialized successfully")
      else:
          sys.exit(f"Error: {response.json().get("error")}")
    except requests.exceptions.ConnectionError:
      sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")
    except Exception:
        sys.exit(f"Error: Invalid file: {path}")

def cmd_get_commits(args, cfg):
    url = require_remote_url(cfg) if args.remote else DEFAULT_ENGINE_URL
    try:
        response = requests.get(f"{url}/v0/models/commits/{args.tag}")
        commits = response.json()
        print(commits) # once commits endpoint is updated to use tags, we can fix this
    except requests.exceptions.ConnectionError:
        sys.exit(f"Error: failed to connect to server. Is {"remote" if args.remote else "engine"} running at {url}?")

def cmd_get_lineage(args, cfg):
    url = require_remote_url(cfg) if args.remote else DEFAULT_ENGINE_URL
    try:
        response = requests.get(f"{url}/v0/models/lineage/{args.tag}")
        lineage = response.json()
        print(lineage) # once lineage endpoint is updated to use tags, we can fix this
    except requests.exceptions.ConnectionError:
        sys.exit(f"Error: failed to connect to server. Is {"remote" if args.remote else "engine"} running at {url}?")

def cmd_get_models(args, cfg):
    try:
        response = requests.get(f"{DEFAULT_ENGINE_URL}/v0/models")
        models = response.json()
        for model in models:
            print(f"- Name: {model.get("meta").get("name")}")
            print(f"  Summary: {model.get("meta").get("summary")}")
            print(f"  Version: {model.get("meta").get("version")}")
            print(f"  Last Updated: {model.get("meta").get("updatedDate")}\n")
    except requests.exceptions.ConnectionError:
        sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")

# -----------------------------------------------------------------------------
# CLI wiring (argparse)
# -----------------------------------------------------------------------------
EPILOG = """Examples:
  py -m cli set-url http://localhost:8080
  py -m cli set-token <JWT>
  py -m cli clear-token
  py -m cli pull my-tag
  py -m cli push my-tag
  py -m cli commit my-tag
  py -m cli init ./model.json
  py -m cli get-commits -r test-model:1.0
  py -m cli get-lineage -r test-model:1.0
  py -m cli get-models
"""

def build_parser():
    p = argparse.ArgumentParser(
        prog="cli",
        description="OpenDI Model Hub CLI",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=EPILOG,
    )
    sub = p.add_subparsers(dest="command", required=True)

    # set-url
    sp = sub.add_parser("set-url", help="Set the remote API URL, e.g., http://localhost:8080")
    sp.add_argument("url", help="API remote URL")
    sp.set_defaults(func=cmd_set_url)

    # set-token (login)
    sp = sub.add_parser("set-token", help="Save the JWT token used for Authorization (login)")
    sp.add_argument("token", help="JWT token string")
    sp.set_defaults(func=cmd_set_token)

    # pull
    sp = sub.add_parser("pull", help="Pull a model using a local tag (stub)")
    sp.add_argument("tag", help="Local tag name")
    sp.set_defaults(func=cmd_pull)

    # push
    sp = sub.add_parser("push", help="Push a model using a local tag (stub)")
    sp.add_argument("tag", help="Local tag name")
    sp.set_defaults(func=cmd_push)

    # commit
    sp = sub.add_parser("commit", help="Create a local commit for a tag (stub)")
    sp.add_argument("tag", help="Local tag name")
    sp.set_defaults(func=cmd_commit)

    # init
    sp = sub.add_parser("init", help="Initialize a local model from a JSON file")
    sp.add_argument("path", help="Path to model.json")
    sp.set_defaults(func=cmd_init)

    # get-commits
    sp = sub.add_parser("get-commits", help="Show commits for a model using a tag (local by default, remote with -r)")
    sp.add_argument("tag", help="Model tag")
    sp.add_argument("-r", "--remote", action="store_true", help="Get commits from remote instead of local")
    sp.set_defaults(func=cmd_get_commits)

    # lineage
    sp = sub.add_parser("get-lineage", help="Show lineage for a model using a tag (local by default, remote with -r)")
    sp.add_argument("tag", help="Model tag")
    sp.add_argument("-r", "--remote", action="store_true", help="Get lineage from remote instead of local")
    sp.set_defaults(func=cmd_get_lineage)

    # get-models
    sp = sub.add_parser("get-models", help="List models (stub)")
    sp.set_defaults(func=cmd_get_models)

    # clear-token (logout)
    sp = sub.add_parser("clear-token", help="Clear the stored JWT token (logout)")
    sp.set_defaults(func=cmd_clear_token)

    return p

def main(argv=None):
    argv = argv if argv is not None else sys.argv[1:]
    parser = build_parser()
    args = parser.parse_args(argv)
    cfg = load_config()
    args.func(args, cfg)

if __name__ == "__main__":
    main()