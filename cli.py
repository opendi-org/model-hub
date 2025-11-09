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
  cli lineage -r "tag"
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
    sys.exit("This CLI needs 'requests'. Install with: py -m pip install requests")

# -----------------------------------------------------------------------------
# Config management
# -----------------------------------------------------------------------------
CONFIG_DIR = Path(__file__).resolve().parent / ".opendi_cli"
CONFIG_PATH = CONFIG_DIR / "config.json"

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

def require_base_url(cfg: dict) -> str:
    base = (cfg.get("base_url") or "").strip()
    if not base:
        sys.exit("Error: no base URL set. Run: py -m cli set-url http://localhost:8080")
    return base.rstrip("/")

def require_token(cfg: dict) -> str:
    tok = (cfg.get("token") or "").strip()
    if not tok:
        sys.exit("Error: not logged in. Run: py -m cli set-token <JWT>")
    return tok

# -----------------------------------------------------------------------------
# Implemented commands
# -----------------------------------------------------------------------------
def cmd_set_url(args, cfg):
    """Save/override the base API URL."""
    url = args.url.strip()
    parsed = urlparse(url)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        sys.exit("Invalid URL. Example: http://localhost:8080")
    cfg["base_url"] = url
    save_config(cfg)
    print(f"Base URL set to {url}")
    print(f"(Saved at {CONFIG_PATH})")

def cmd_set_token(args, cfg):
    """Save/replace the auth token (login)."""
    token = args.token.strip()
    if not token:
        sys.exit("Token cannot be empty.")
    cfg["token"] = token
    save_config(cfg)
    print("Token saved. (Login successful)")
    print(f"(Saved at {CONFIG_PATH})")

def cmd_clear_token(args, cfg):
    """Clear the auth token (logout)."""
    cfg["token"] = ""
    save_config(cfg)
    print("Token cleared. (Logged out)")
    print(f"(Config file: {CONFIG_PATH})")

# -----------------------------------------------------------------------------
# Stubs (help pages exist; implementation still needs to be added)
# -----------------------------------------------------------------------------
def cmd_pull(args, cfg):
    require_base_url(cfg)
    print(f"[TODO] pull tag={args.tag}")

def cmd_push(args, cfg):
    require_base_url(cfg)
    require_token(cfg)
    print(f"[TODO] push tag={args.tag}")

def cmd_commit(args, cfg):
    print(f"[TODO] commit tag={args.tag}")

def cmd_init(args, cfg):
    print(f"[TODO] init path={args.path}")

def cmd_get_commits(args, cfg):
    require_base_url(cfg)
    print(f"[TODO] get-commits remote_tag={args.remote_tag}")

def cmd_lineage(args, cfg):
    require_base_url(cfg)
    print(f"[TODO] lineage remote_tag={args.remote_tag}")

def cmd_get_models(args, cfg):
    require_base_url(cfg)
    print("[TODO] get-models")

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
  py -m cli get-commits -r 1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d
  py -m cli lineage -r 1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d
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
    sp = sub.add_parser("set-url", help="Set the base API URL, e.g., http://localhost:8080")
    sp.add_argument("url", help="API base URL")
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
    sp = sub.add_parser("init", help="Initialize a local model from a JSON file (stub)")
    sp.add_argument("path", help="Path to model.json")
    sp.set_defaults(func=cmd_init)

    # get-commits
    sp = sub.add_parser("get-commits", help="Show remote commits for a model UUID (-r) (stub)")
    sp.add_argument("-r", "--remote-tag", required=True, help="Remote tag (UUID)")
    sp.set_defaults(func=cmd_get_commits)

    # lineage
    sp = sub.add_parser("lineage", help="Show lineage for a model UUID (-r) (stub)")
    sp.add_argument("-r", "--remote-tag", required=True, help="Remote tag (UUID)")
    sp.set_defaults(func=cmd_lineage)

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