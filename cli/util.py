import json, re, random, string

UUID_RE = re.compile(r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", re.I)
LOCAL_RE = re.compile(r"^[A-Za-z0-9._-]{3,40}$")

def is_valid_uuid(s:str) -> bool:
    return bool(UUID_RE.match(s or ""))

def is_valid_local_tag(s:str) -> bool:
    return bool(LOCAL_RE.match(s or ""))

def gen_local_tag(name: str|None=None) -> str:
    base = (name or "model").lower()
    base = re.sub(r"[^A-Za-z0-9._-]+", "-", base).strip("-")
    base = base[:20] if base else "model"
    suffix = "".join(random.choices(string.ascii_lowercase + string.digits, k=6))
    return f"{base}-{suffix}"

def load_json_file(path:str) -> dict:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)

def pretty_json(obj) -> str:
    return json.dumps(obj, indent=2, sort_keys=True)