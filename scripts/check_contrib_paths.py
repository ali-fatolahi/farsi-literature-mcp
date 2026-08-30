#!/usr/bin/env python3

import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
COMMUNITY = ROOT / "community"
USERNAME_RE = re.compile(r"^[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?$")
TIMESTAMP_RE = re.compile(r"^\d{4}-\d{2}-\d{2}T\d{6}Z$")
ALLOWED_TOP = {"ideas", "design-proposals", "test-results"}
ALLOWED_SUFFIXES = {".md"}


def iter_submissions():
    for top in sorted(COMMUNITY.iterdir()):
        if not top.is_dir() or top.name not in ALLOWED_TOP:
            continue
        for user in sorted(top.iterdir()):
            if not user.is_dir() or not USERNAME_RE.match(user.name):
                yield (user, f"invalid username directory: {user.relative_to(ROOT)}")
                continue
            for stamp in sorted(user.iterdir()):
                if not stamp.is_dir() or not TIMESTAMP_RE.match(stamp.name):
                    yield (stamp, f"invalid timestamp directory: {stamp.relative_to(ROOT)}")
                    continue
                for file in sorted(stamp.iterdir()):
                    if file.is_dir():
                        yield (file, f"unexpected nested directory: {file.relative_to(ROOT)}")
                        continue
                    if file.suffix.lower() not in ALLOWED_SUFFIXES:
                        yield (file, f"unsupported extension: {file.relative_to(ROOT)}")


def main() -> int:
    if not COMMUNITY.exists():
        print(f"Missing community directory: {COMMUNITY}", file=sys.stderr)
        return 1

    errors = []
    for path, message in iter_submissions():
        if path is not None:
            errors.append(f"{message}")

    if errors:
        print("Invalid contribution paths detected:", file=sys.stderr)
        for err in errors:
            print(f"- {err}", file=sys.stderr)
        return 1

    print("Community contribution directories look valid.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
