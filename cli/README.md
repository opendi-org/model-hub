# OpenDI CLI

Cross-platform CLI for the OpenDI Model Hub. Distributed via [PyPI](https://pypi.org/); developers use [uv](https://docs.astral.sh/uv/) for a fast, reproducible setup.

## For users (install from PyPI)

The package is published under the name **opendi**.

### Requirements

- Python 3.10+

### Install with pipx (Windows)

```bash
py -m pip install --user pipx
py -m pipx install opendi
```

### Install with pipx (macOS / Linux)

```bash
python3 -m pip install --user pipx
python3 -m pipx install opendi
```

Then run the CLI from anywhere:

```bash
opendi --help
```

Uninstall:

- Windows: `py -m pipx uninstall opendi`
- macOS / Linux: `python3 -m pipx uninstall opendi`

## For developers (run from source with uv)

We use **uv** for virtual environments and dependency management. Install uv if needed: <https://docs.astral.sh/uv/getting-started/installation/>.
On Windows, the PowerShell installation option is recommended.

From the `cli/` directory:

```bash
# Create a virtual environment and install the package in editable mode (and optional dev deps)
uv venv
uv sync --extra dev

# Activate the venv (optional; uv run works without it)
# Windows:
.venv\Scripts\activate
# macOS/Linux:
source .venv/bin/activate

# Run the CLI
uv run opendi --help
```

- **`uv venv`** – Creates `.venv` in the project. uv’s venvs are compatible with pip and the rest of the ecosystem.
- **`uv sync`** – Reads `pyproject.toml`, resolves dependencies, and installs the current project in editable mode. Use `--extra dev` to install dev dependencies (e.g. pytest). uv creates a `uv.lock` for reproducible installs — **commit `uv.lock`** so all devs and CI use the same versions.
- **`uv run opendi`** – Runs the `opendi` console script from the project environment (no need to activate the venv). Editable install means your code changes are used immediately.
- **`uv run pytest`** – Run tests (after `uv sync --extra dev`).

To add new dependencies, edit `pyproject.toml` and run `uv sync` again.

### Login setup (for `opendi login`)

The CLI now uses the backend browser-assisted flow:

1. CLI calls `POST /v0/auth/cli/login`
2. CLI opens the returned login URL in your browser
3. After browser approval, CLI polls `POST /v0/auth/cli/poll`
4. CLI stores the returned OpenDI access token in OS keyring

No local Google client secrets are needed in the CLI.

By default, CLI talks to `http://localhost:8080`. Override with:

```bash
export OPENDI_API_URL=http://localhost:8080
```

If `OPENDI_API_URL` is set to `http://` with a non-local host, the CLI prints a security warning and continues. Use HTTPS for production endpoints.

### Credential storage

After `opendi login`, the OpenDI access token is stored in the **OS credential manager** via the [keyring](https://pypi.org/project/keyring/) library:

| OS      | Backend                          |
|---------|----------------------------------|
| Windows | Windows Credential Manager       |
| macOS   | Keychain                         |
| Linux   | Secret Service (GNOME Keyring / KWallet) |

- `opendi login` — stores access token; subsequent runs reuse it.
- `opendi logout` — deletes stored token from the OS credential manager.

No tokens are written to disk files.

## Publishing to PyPI

When ready to publish:

1. Bump version in `pyproject.toml`.
2. Build the package:
   ```bash
   uv build
   ```
3. Upload to PyPI (use a token from pypi.org):
   ```bash
   uv publish
   ```

Post-publish smoke test (recommended via `pipx`):

- Windows:
  ```bash
  py -m pip install --user pipx
  py -m pipx install opendi
  opendi --help
  ```
- macOS / Linux:
  ```bash
  python3 -m pip install --user pipx
  python3 -m pipx install opendi
  opendi --help
  ```

## Project layout

- `src/opendi/main.py` — CLI entry point (`opendi` command).
- `src/opendi/auth.py` — browser-assisted login flow.
- `src/opendi/credential_storage.py` — OS keyring token storage.
- `src/opendi/local_cache.py` — local model cache storage/retrieval.
- `src/opendi/cmds/repo_cmds.py` — repository command handlers.
- `src/opendi/cmds/tag_cmds.py` — tag/local command handlers.
- `src/opendi/cmds/shared.py` — shared parsing, API helpers, and output helpers.
- `src/opendi/log.py` — CLI logging setup.
- `tests/` — pytest suite for auth, commands, cache, and credential storage.
- `pyproject.toml` + `uv.lock` — package metadata and locked dependencies.
