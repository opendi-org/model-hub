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

def save_mappings(maps: list) -> None:
  """Persist mappings JSON to disk (inside repo)."""
  CONFIG_DIR.mkdir(parents=True, exist_ok=True)
  MAPPINGS_PATH.write_text(json.dumps(maps, indent=2), encoding="utf-8")

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
def cmd_set_url(args, cfg, maps):
  """Save/override the remote API URL."""
  url = args.url.strip()
  parsed = urlparse(url)
  if parsed.scheme not in ("http", "https") or not parsed.netloc:
    sys.exit("Error: invalid URL. Example: http://localhost:8080")
  cfg["remote_url"] = url
  save_config(cfg)
  print(f"Remote URL set to {url}")
  print(f"(Saved at {CONFIG_PATH})")

def cmd_set_token(args, cfg, maps):
  """Save/replace the auth token (login)."""
  token = args.token.strip()
  if not token:
    sys.exit("Error: token cannot be empty.")
  cfg["token"] = token
  save_config(cfg)
  print("Token saved")
  print(f"(Saved at {CONFIG_PATH})")

def cmd_clear_token(args, cfg, maps):
  """Clear the auth token (logout)."""
  cfg["token"] = ""
  save_config(cfg)
  print("Token cleared")
  print(f"(Cleared from: {CONFIG_PATH})")

def cmd_pull(args, cfg, maps):
  # need token and url for pull
  token = require_token(cfg)
  remote_url = require_remote_url(cfg)

  # get the model from the remote
  try: 
    response = requests.get(f"{remote_url}/v0/models/tag/{args.tag}", headers={"Authorization": f"Bearer {token}"})
    if response.status_code != 200:
      sys.exit(f"Error: {response.json().get("error")}")
    model_data = response.json()
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is remote running at {remote_url}?")

  # check if we already have a mapping for this model
  # if we do, make sure the file path is valid, otherwise create a new path
  mapping_index = next((i for i, m in enumerate(maps) if m.get("remote") == model_data.get("meta").get("UUID")), None)
  if mapping_index is None or not Path(maps[mapping_index].get("path")).exists():
    model_path = Path(__file__).resolve().parent / f"{args.tag.split(":")[0]}.json"
  else:
    model_path = Path(maps[mapping_index].get("path"))

  # save for use later
  retrieved_model_uuid = model_data.get("meta").get("UUID")
  if mapping_index is None:
    del model_data["meta"]["UUID"]
  else:
    model_data["meta"]["UUID"] = maps[mapping_index].get("local")
  del model_data["addons"]

  # post the model to the engine, if we have no mapping, this should create a new model
  # if we do have a mapping, this should create a new commit on an existing model
  try:
    response = requests.post(f"{DEFAULT_ENGINE_URL}/v0/models", json=model_data)
    if response.status_code != 200:
      sys.exit(f"Error: {response.json().get("error")}")
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")

  # try writing the model we retrieved to disk, new file if the mapping didn't already exist
  try:
    model_path.write_text(json.dumps(model_data, indent=2), encoding="utf-8")
  except Exception:
    sys.exit(f"Error: failed to write to model file: ${model_path}")

  print("Model written to disk successfully")
  print(f"(Saved at: ${model_path})")

  # finally, update the mapping so we keep track of everything properly
  mapping = {
    "tag": response.json().get("addons").get("tag"), 
    "path": str(model_path),
    "local": response.json().get("meta").get("UUID"),
    "remote": retrieved_model_uuid
  }
  if mapping_index is None:
    maps.append(mapping)
  else:
    maps[mapping_index] = mapping

  save_mappings(maps)

  print("Model pulled successfully")
  print(f"(Tag: {mapping.get("tag")})")

def cmd_push(args, cfg, maps):
  # need token and url for push
  token = require_token(cfg)
  remote_url = require_remote_url(cfg)

  # find the mapping for the tag
  mapping_index = next((i for i, m in enumerate(maps) if m.get("tag") == args.tag), None)
  if mapping_index is None:
    sys.exit(f"Error: no mapping found for tag: {args.tag}")

  # find the local model with the specified tag
  try:
    response = requests.get(f"{DEFAULT_ENGINE_URL}/v0/models/tag/{args.tag}")
    if response.status_code != 200:
      sys.exit(f"Error: {response.json().get("error")}")
    model_data = response.json()
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")

  # if this local model is already connected to a remote, update the uuid so we make updates and don't create an entirely new model
  if maps[mapping_index].get("remote"):
    model_data["meta"]["UUID"] = maps[mapping_index].get("remote")

  # post the local model to the remote
  try:
    response = requests.post(f"{remote_url}/v0/models", json=model_data, headers={"Authorization": f"Bearer {token}"})
    if response.status_code == 200:
      maps[mapping_index]["remote"] = response.json().get("meta").get("UUID")
      save_mappings(maps)
      print("Model pushed to remote successfully")
    else:
      sys.exit(f"Error: {response.json().get("error")}")
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is remote running at {remote_url}?")
  

def cmd_commit(args, cfg, maps):
  # find the mapping for the tag
  mapping_index = next((i for i, m in enumerate(maps) if m.get("tag") == args.tag), None)
  if mapping_index is None:
    sys.exit(f"Error: no mapping found for tag: {args.tag}")

  # make sure that the model path is valid
  model_path = Path(maps[mapping_index].get("path"))
  if not model_path.exists():
    sys.exit(f"Error: no model found at: {model_path}")

  # load the model from the stored path and post it to the engine
  try:
    model_json = json.loads(model_path.read_text(encoding="utf-8"))
    if maps[mapping_index].get("local"): # we should have a local uuid stored, use it
      model_json["meta"]["UUID"] = maps[mapping_index].get("local")
    response = requests.post(f"{DEFAULT_ENGINE_URL}/v0/models", json=model_json)
    if response.status_code == 200:
      model_tag = response.json().get("addons").get("tag")
      maps[mapping_index]["tag"] = model_tag
      save_mappings(maps)
      print("Commit created successfully")
      print(f"(Tag: {model_tag})")
    else:
      sys.exit(f"Error: {response.json().get("error")}")
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")
  except Exception:
    sys.exit(f"Error: invalid file: {model_path}")

def cmd_init(args, cfg, maps):
  path = Path(args.path)
  
  # load the provided file, post it to the engine, and then create a mapping for it
  try:
    model_json = json.loads(path.read_text(encoding="utf-8"))
    response = requests.post(f"{DEFAULT_ENGINE_URL}/v0/models", json=model_json)
    if response.status_code == 200:
      model_tag = response.json().get("addons").get("tag")
      maps = [m for m in maps if m.get("tag") != model_tag] 
      maps.append({
        "tag": model_tag, 
        "path": str(Path(__file__).resolve().parent / args.path),
        "local": response.json().get("meta").get("UUID"),
        "remote": ""
      })
      save_mappings(maps)
      print("Model initialized successfully")
      print(f"(Tag: {model_tag})")
    else:
      sys.exit(f"Error: {response.json().get("error")}")
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is engine running at {DEFAULT_ENGINE_URL}?")
  except Exception:
    sys.exit(f"Error: invalid file: {path}")

def cmd_get_commits(args, cfg, maps):
  url = require_remote_url(cfg) if args.remote else DEFAULT_ENGINE_URL
  
  # get the commits from either the local or remote instance
  try:
    response = requests.get(f"{url}/v0/models/commits/{args.tag}", headers={"Authorization": f"Bearer {require_token(cfg)}"} if args.remote else None)
    if response.status_code == 200:
      commits = response.json()
      for commit in commits:
        print(f"\n- Version: {commit.get("version")}")
        print(f"  Created At: {commit.get("createdAt")}")
    else:
      sys.exit(f"No commits found for tag: {args.tag}")
  except requests.exceptions.ConnectionError:
      sys.exit(f"Error: failed to connect to server. Is {"remote" if args.remote else "engine"} running at {url}?")

def cmd_get_lineage(args, cfg, maps):
  url = require_remote_url(cfg) if args.remote else DEFAULT_ENGINE_URL
  
  # get the lineage from either the local or remote instance
  try:
    response = requests.get(f"{url}/v0/models/lineage/{args.tag}", headers={"Authorization": f"Bearer {require_token(cfg)}"} if args.remote else None)
    if response.status_code == 200:
      lineage = response.json()
      for model in lineage:
        print(f"\n- Name: {model.get("meta").get("name")}")
        print(f"  Tag: {model.get("addons").get("tag")}")
        print(f"  Summary: {model.get("meta").get("summary")}")
        print(f"  Version: {model.get("meta").get("version")}")
        print(f"  Last Updated: {model.get("meta").get("updatedDate")}")
    else:
      sys.exit(f"Error: no model found for tag: {args.tag}")
  except requests.exceptions.ConnectionError:
    sys.exit(f"Error: failed to connect to server. Is {"remote" if args.remote else "engine"} running at {url}?")

def cmd_get_models(args, cfg, maps):
  # get the models from the engine
  try:
    response = requests.get(f"{DEFAULT_ENGINE_URL}/v0/models")
    if response.status_code == 200:
      models = response.json()
      for model in models:
        print(f"\n- Name: {model.get("meta").get("name")}")
        print(f"  Tag: {model.get("addons").get("tag")}")
        print(f"  Summary: {model.get("meta").get("summary")}")
        print(f"  Version: {model.get("meta").get("version")}")
        print(f"  Last Updated: {model.get("meta").get("updatedDate")}")
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
  sp = sub.add_parser("pull", help="Pull a model using a remote tag")
  sp.add_argument("tag", help="Remote tag name")
  sp.set_defaults(func=cmd_pull)

  # push
  sp = sub.add_parser("push", help="Push a model using a local tag")
  sp.add_argument("tag", help="Local tag name")
  sp.set_defaults(func=cmd_push)

  # commit
  sp = sub.add_parser("commit", help="Create a local commit for a tag")
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
  sp = sub.add_parser("get-models", help="List local models")
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
  maps = load_mappings()
  args.func(args, cfg, maps)

if __name__ == "__main__":
  main()