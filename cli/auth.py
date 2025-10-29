import os, json, base64

CONFIG_DIR = os.path.join(os.path.expanduser("~"), ".opendi_cli")
CONFIG_PATH = os.path.join(CONFIG_DIR, "config.json")

class Config:
    def __init__(self, base_url:str="", token:str=""):
        self.base_url = base_url
        self.token = token

    @staticmethod
    def load():
        if not os.path.exists(CONFIG_PATH):
            return Config()
        with open(CONFIG_PATH, "r", encoding="utf-8") as f:
            d = json.load(f)
        return Config(d.get("base_url",""), d.get("token",""))

    def save(self):
        os.makedirs(CONFIG_DIR, exist_ok=True)
        with open(CONFIG_PATH, "w", encoding="utf-8") as f:
            json.dump({"base_url": self.base_url, "token": self.token}, f, indent=2)

    @staticmethod
    def peek_claims(jwt_token: str):
        try:
            parts = jwt_token.split(".")
            if len(parts) != 3: return {}
            padded = parts[1] + "==="  # URL-safe b64
            payload = base64.urlsafe_b64decode(padded.encode("utf-8"))
            return json.loads(payload.decode("utf-8"))
        except Exception:
            return {}