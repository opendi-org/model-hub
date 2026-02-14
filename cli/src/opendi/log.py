"""CLI logging: stderr, INFO by default."""

import logging
import sys


def configure_logging(level: int = logging.INFO) -> None:
    """Configure the opendi package logger to stderr."""
    log = logging.getLogger("opendi")
    log.setLevel(level)
    if not log.handlers:
        h = logging.StreamHandler(sys.stderr)
        h.setFormatter(logging.Formatter("%(levelname)s: %(message)s"))
        log.addHandler(h)
