import os, json, sqlite3, datetime
from dataclasses import dataclass
from typing import Optional

DB_PATH = os.path.join(os.path.expanduser("~"), ".opendi_cli", "local.db")

@dataclass
class LocalRecord:
    local_tag: str
    remote_uuid: Optional[str]
    model_json: dict
    version: int
    created_at: str
    updated_at: str

class Store:
    def __init__(self, conn: sqlite3.Connection):
        self.conn = conn

    @staticmethod
    def open():
        os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)
        conn = sqlite3.connect(DB_PATH)
        conn.execute("PRAGMA foreign_keys = ON")
        s = Store(conn)
        s._init_schema()
        return s

    def _init_schema(self):
        self.conn.executescript("""
        CREATE TABLE IF NOT EXISTS local_models (
            local_tag   TEXT PRIMARY KEY,
            remote_uuid TEXT,
            model_json  TEXT NOT NULL,
            version     INTEGER DEFAULT 0,
            created_at  TEXT NOT NULL,
            updated_at  TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS commits (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            local_tag  TEXT NOT NULL,
            created_at TEXT NOT NULL,
            message    TEXT,
            before_json TEXT,
            after_json  TEXT,
            FOREIGN KEY(local_tag) REFERENCES local_models(local_tag) ON DELETE CASCADE
        );
        """)
        self.conn.commit()

    def has_local_tag(self, local_tag:str) -> bool:
        cur = self.conn.execute("SELECT 1 FROM local_models WHERE local_tag=?", (local_tag,))
        return cur.fetchone() is not None

    def put_local_model(self, local_tag: str, model_obj: dict, remote_uuid: str|None=None):
        now = datetime.datetime.utcnow().isoformat()
        js = json.dumps(model_obj, separators=(",",":"))
        if self.has_local_tag(local_tag):
            self.conn.execute("""
                UPDATE local_models
                   SET model_json=?, remote_uuid=COALESCE(?, remote_uuid), updated_at=?, version=version+1
                 WHERE local_tag=?
            """, (js, remote_uuid, now, local_tag))
        else:
            self.conn.execute("""
                INSERT INTO local_models(local_tag, remote_uuid, model_json, version, created_at, updated_at)
                VALUES(?,?,?,?,?,?)
            """, (local_tag, remote_uuid, js, 0, now, now))
        self.conn.commit()

    def set_remote_tag(self, local_tag:str, remote_uuid:str):
        if not self.has_local_tag(local_tag):
            raise KeyError("Local tag not found")
        # local conflict detection: ensure no *other* local tag already bound to this remote uuid
        cur = self.conn.execute("""
            SELECT local_tag FROM local_models WHERE remote_uuid=? AND local_tag<>?
        """, (remote_uuid, local_tag))
        row = cur.fetchone()
        if row:
            raise SystemExit(f"Invalid Remote Tag: already associated to local tag '{row[0]}' (UC9).")
        self.conn.execute("UPDATE local_models SET remote_uuid=?, updated_at=? WHERE local_tag=?",
                          (remote_uuid, datetime.datetime.utcnow().isoformat(), local_tag))
        self.conn.commit()

    def get_local_model_record(self, local_tag:str) -> LocalRecord|None:
        cur = self.conn.execute("""
            SELECT local_tag, remote_uuid, model_json, version, created_at, updated_at
              FROM local_models WHERE local_tag=?
        """, (local_tag,))
        row = cur.fetchone()
        if not row: return None
        return LocalRecord(
            local_tag=row[0],
            remote_uuid=row[1],
            model_json=json.loads(row[2]),
            version=int(row[3]),
            created_at=row[4],
            updated_at=row[5],
        )

    def get_local_model_json(self, local_tag:str) -> dict:
        rec = self.get_local_model_record(local_tag)
        if not rec: raise KeyError("Local tag not found")
        return rec.model_json

    def add_commit(self, local_tag:str, before:dict|None, after:dict, message:str=""):
        now = datetime.datetime.utcnow().isoformat()
        self.conn.execute("""
            INSERT INTO commits(local_tag, created_at, message, before_json, after_json)
            VALUES (?,?,?,?,?)
        """, (local_tag, now, message, json.dumps(before or {}), json.dumps(after)))
        self.conn.commit()

    def list_commits(self, local_tag:str):
        cur = self.conn.execute("""
            SELECT id, created_at, message FROM commits WHERE local_tag=? ORDER BY id DESC
        """, (local_tag,))
        return cur.fetchall()