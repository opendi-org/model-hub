#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse, sys, json, textwrap
from .auth import Config
from .store import Store
from .api import ApiClient
from .util import (
    load_json_file, is_valid_local_tag, is_valid_uuid, gen_local_tag, pretty_json
)

EPILOG = """\
Examples:
  # set base URL + login with a token you already have (UC8)
  py -m cli login --base-url http://localhost:8080 --token <JWT>

  # DEV login using /auth/testlogin?id= (when DEV_MODE=true on server)
  py -m cli login --base-url http://localhost:8080 --user-id 1

  # create a local model (UC10)
  py -m cli init ./my_model.json --local-tag risk-v1

  # associate local with remote UUID tag (UC9)
  py -m cli tag risk-v1 --remote-tag 1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d

  # make a local commit from a file (UC12)
  py -m cli commit risk-v1 ./my_model_updated.json -m "adjusted priors"

  # push local → remote (UC13)
  py -m cli push risk-v1

  # pull a remote model to a file and optionally tag it locally (UC11)
  py -m cli pull --remote-tag 1a2b3c4d-... --out pulled.json --local-tag pulled-risk

  # see lineage or children (UC14)
  py -m cli lineage risk-v1
  py -m cli lineage risk-v1 --children-only

  # see commits (UC15)
  py -m cli commits risk-v1
"""

def build_parser():
    p = argparse.ArgumentParser(
        prog="opendi-cli",
        description="OpenDI Model Hub CLI (Iteration 3: CLI Utility)",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=EPILOG,
    )
    sub = p.add_subparsers(dest="cmd", required=True)

    # config / auth
    sp = sub.add_parser("login", help="Authenticate and save token (UC8).")
    sp.add_argument("--base-url", help="API base URL, e.g. http://localhost:8080")
    g = sp.add_mutually_exclusive_group(required=True)
    g.add_argument("--token", help="JWT token to store")
    g.add_argument("--user-id", type=int, help="DEV ONLY: fetch token from /auth/testlogin?id=")

    sp = sub.add_parser("logout", help="Forget stored token.")
    sp = sub.add_parser("status", help="Show config and token claims.")

    sp = sub.add_parser("set-url", help="Set/override base URL.")
    sp.add_argument("base_url")

    # UC10
    sp = sub.add_parser("init", help="Create local model from JSON file (UC10).")
    sp.add_argument("json_path")
    sp.add_argument("--local-tag", help="Custom local tag (auto-generated if omitted)")

    # UC9
    sp = sub.add_parser("tag", help="Associate a local tag with a remote UUID tag (UC9).")
    sp.add_argument("local_tag")
    sp.add_argument("--remote-tag", required=True, help="Remote tag (UUID)")

    # UC11
    sp = sub.add_parser("pull", help="Pull remote model to file and optionally tag locally (UC11).")
    sp.add_argument("--remote-tag", required=True, help="Remote tag (UUID)")
    sp.add_argument("--out", required=True, help="Output JSON path")
    sp.add_argument("--local-tag", help="Optional local tag to create/update")

    # UC12
    sp = sub.add_parser("commit", help="Create local commit from JSON (UC12).")
    sp.add_argument("local_tag")
    sp.add_argument("json_path")
    sp.add_argument("-m", "--message", default="", help="Optional commit message")

    # UC13
    sp = sub.add_parser("push", help="Push local model to remote (UC13).")
    sp.add_argument("local_tag")

    # UC14
    sp = sub.add_parser("lineage", help="Show lineage or children for a model (UC14).")
    sp.add_argument("local_tag")
    sp.add_argument("--children-only", action="store_true")

    # UC15
    sp = sub.add_parser("commits", help="Show remote commits for a model (UC15).")
    sp.add_argument("local_tag")

    return p

def require_base_url(cfg: Config):
    if not cfg.base_url:
        sys.exit("Error: no base URL set. Use: python -m cli set-url http://host:8080 OR login --base-url ...")

def require_token(cfg: Config):
    if not cfg.token:
        sys.exit("Error: not authenticated. Run: python -m cli login --token <JWT>  (or --user-id <id> in DEV)")

def main(argv=None):
    argv = argv if argv is not None else sys.argv[1:]
    cfg = Config.load()
    store = Store.open()

    parser = build_parser()
    args = parser.parse_args(argv)

    # ===== auth & config =====
    if args.cmd == "login":
        if args.base_url:
            cfg.base_url = args.base_url
        require_base_url(cfg)

        api = ApiClient(cfg)
        if args.user_id is not None:
            token = api.test_login(args.user_id)  # DEV endpoint
        else:
            token = args.token
        cfg.token = token
        cfg.save()
        print("Logged in and token saved.")
        return

    if args.cmd == "logout":
        cfg.token = ""
        cfg.save()
        print("Token cleared.")
        return

    if args.cmd == "status":
        print(f"Base URL: {cfg.base_url or '(not set)'}")
        if cfg.token:
            claims = Config.peek_claims(cfg.token)
            print("Token: present")
            print("Claims:", json.dumps(claims, indent=2, sort_keys=True))
        else:
            print("Token: (none)")
        return

    if args.cmd == "set-url":
        cfg.base_url = args.base_url
        cfg.save()
        print(f"Base URL set to {cfg.base_url}")
        return

    # ===== model ops =====
    if args.cmd == "init":
        data = load_json_file(args.json_path)
        # minimal validity: must be JSON object with "meta" dict
        if not isinstance(data, dict) or "meta" not in data or not isinstance(data["meta"], dict):
            sys.exit("Invalid Model: JSON must contain an object with a 'meta' object (UC10).")

        local_tag = args.local_tag or gen_local_tag(data.get("meta", {}).get("name"))
        if not is_valid_local_tag(local_tag):
            sys.exit("Invalid Local Tag (UC10/UC11): must be 3–40 chars of letters/digits/._-")
        store.put_local_model(local_tag, data)
        print(f"Local model initialized as tag '{local_tag}'.")
        return

    if args.cmd == "tag":
        if not is_valid_local_tag(args.local_tag):
            sys.exit("Local Tag Not Found/Invalid (UC9).")
        if not store.has_local_tag(args.local_tag):
            sys.exit("Local Tag Not Found (UC9).")
        if not is_valid_uuid(args.remote_tag):
            sys.exit("Invalid Remote Tag: must be a UUID like 8-4-4-4-12 hex (UC9).")
        # ensure no remote-tag conflicts in this local DB
        store.set_remote_tag(args.local_tag, args.remote_tag)
        print(f"Associated local '{args.local_tag}' with remote tag (UUID) {args.remote_tag}.")
        return

    if args.cmd == "pull":
        require_base_url(cfg)
        api = ApiClient(cfg)
        if not is_valid_uuid(args.remote_tag):
            sys.exit("Remote Tag Not Found/Invalid (UC11).")
        model = api.get_model_by_uuid(args.remote_tag)  # raises on 404
        # write to file
        with open(args.out, "w", encoding="utf-8") as f:
            f.write(pretty_json(model))
        print(f"Wrote model to {args.out}.")
        if args.local_tag:
            if not is_valid_local_tag(args.local_tag):
                sys.exit("Invalid Local Tag (UC11).")
            store.put_local_model(args.local_tag, model, remote_uuid=args.remote_tag)
            print(f"Stored/updated local tag '{args.local_tag}'.")
        return

    if args.cmd == "commit":
        if not store.has_local_tag(args.local_tag):
            sys.exit("Local Tag Not Found (UC12).")
        data = load_json_file(args.json_path)
        if not isinstance(data, dict) or "meta" not in data or not isinstance(data["meta"], dict):
            sys.exit("Invalid Model: must include 'meta' object (UC12).")
        before = store.get_local_model_json(args.local_tag)
        store.add_commit(args.local_tag, before, data, message=args.message)
        store.put_local_model(args.local_tag, data)  # update working copy
        print(f"Commit recorded for '{args.local_tag}'.")
        return

    if args.cmd == "push":
        require_base_url(cfg)
        require_token(cfg)
        api = ApiClient(cfg)
        rec = store.get_local_model_record(args.local_tag)
        if not rec:
            sys.exit("Local Tag Not Found (UC13).")
        if not rec.remote_uuid:
            sys.exit("Missing Remote Tag (UC13): use 'tag <local> --remote-tag <uuid>' first.")
        # ensure meta.uuid matches remote-tag
        data = rec.model_json
        data.setdefault("meta", {})
        data["meta"]["uuid"] = rec.remote_uuid
        # PUT full model to server (backend will create commit+version)
        api.put_model(data)  # raises on error (401/403/409/etc)
        print(f"Pushed '{args.local_tag}' to remote UUID {rec.remote_uuid}.")
        return

    if args.cmd == "lineage":
        require_base_url(cfg)
        api = ApiClient(cfg)
        rec = store.get_local_model_record(args.local_tag)
        if not rec or not rec.remote_uuid:
            sys.exit("Local Tag Not Found or Missing Remote Tag (UC14).")
        if args.children_only:
            items = api.get_children(rec.remote_uuid)
            header = "Children"
        else:
            items = api.get_lineage(rec.remote_uuid)
            header = "Lineage (earliest → latest ancestor)"
        print(f"{header} for {rec.remote_uuid}:\n")
        if not items:
            print("(none)")
            return
        for i, m in enumerate(items, 1):
            meta = m.get("meta", {})
            print(f"{i:>2}. name={meta.get('name')}  uuid={meta.get('uuid')}  version={meta.get('version')}")
        return

    if args.cmd == "commits":
        require_base_url(cfg)
        api = ApiClient(cfg)
        rec = store.get_local_model_record(args.local_tag)
        if not rec or not rec.remote_uuid:
            sys.exit("Local Tag Not Found or Missing Remote Tag (UC15).")
        commits = api.get_commits(rec.remote_uuid)
        if not commits:
            print("(no commits)")
            return
        print(f"Commits for {rec.remote_uuid}:")
        for c in commits:
            print(f"  v{c.get('version')}  parent={c.get('parentCommitID') or '-'}  user={c.get('userid')}  ts={c.get('CreatedAt')}")
        return

if __name__ == "__main__":
    main()