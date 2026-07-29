#!/usr/bin/env python3
"""Generate disk_map.json for AI assistants.

The output is intentionally compact and structured. It does not try to mirror
the whole disk recursively; it records useful launcher/program locations only.
"""

from __future__ import annotations

import argparse
import json
import os
from datetime import datetime
from pathlib import Path


DEFAULT_ROOT = Path("L:/")
DEFAULT_LAUNCHER_DIR = Path("L:/_Launcher")
DEFAULT_PROGRAMS_ROOT = Path("L:/Centrum Narzędzi i Oprogramowania")


def scan_exe_files(folder: Path, programs_root: Path, max_depth: int = 2, max_results: int = 5) -> list[dict[str, str]]:
    results: list[dict[str, str]] = []
    base_depth = len(folder.parts)

    for dirpath, dirnames, filenames in os.walk(folder):
        current = Path(dirpath)
        depth = len(current.parts) - base_depth
        if depth >= max_depth:
            dirnames[:] = []
            continue

        for filename in filenames:
            if not filename.lower().endswith(".exe"):
                continue
            full = current / filename
            try:
                rel = full.relative_to(programs_root)
            except ValueError:
                rel = full
            results.append({"name": full.stem, "path": str(rel)})
            if len(results) >= max_results:
                return results
    return results


def scan_categories(programs_root: Path) -> list[dict]:
    if not programs_root.is_dir():
        return []

    categories: list[dict] = []
    for cat_path in sorted((p for p in programs_root.iterdir() if p.is_dir()), key=lambda p: p.name.lower()):
        category = {"name": cat_path.name, "subcategories": []}
        for sub_path in sorted((p for p in cat_path.iterdir() if p.is_dir()), key=lambda p: p.name.lower()):
            subcategory: dict = {"name": sub_path.name}
            exes = scan_exe_files(sub_path, programs_root)
            if exes:
                subcategory["programs"] = exes
            subcategory["item_count"] = sum(1 for p in sub_path.iterdir() if p.is_dir())
            category["subcategories"].append(subcategory)
        categories.append(category)
    return categories


def scan_other_dirs(root: Path) -> list[dict]:
    important = [
        "Centrum Gier i Modyfikacji",
        "Dokumenty",
        "KnowledgeBase",
        "Multimedia",
        "Projekty",
        "Prywatne",
    ]
    locations: list[dict] = []
    for name in important:
        path = root / name
        if not path.is_dir():
            continue
        folders = sorted([p.name for p in path.iterdir() if p.is_dir()], key=str.lower)[:20]
        locations.append({"name": name, "path": str(path), "folders": folders})
    return locations


def build_map(root: Path, launcher_dir: Path, programs_root: Path) -> dict:
    return {
        "version": "1.0",
        "generated": datetime.now().strftime("%Y-%m-%d"),
        "purpose": "Structured disk map for AI assistants helping with this launcher.",
        "root": str(root),
        "launcher_dir": str(launcher_dir),
        "programs_root": str(programs_root),
        "categories": scan_categories(programs_root),
        "other_locations": scan_other_dirs(root),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate disk_map.json for Launcher AI context.")
    parser.add_argument("--root", type=Path, default=DEFAULT_ROOT)
    parser.add_argument("--launcher-dir", type=Path, default=DEFAULT_LAUNCHER_DIR)
    parser.add_argument("--programs-root", type=Path, default=DEFAULT_PROGRAMS_ROOT)
    parser.add_argument("--output", type=Path, default=None)
    args = parser.parse_args()

    output = args.output or (args.launcher_dir / "data" / "generated" / "disk_map.json")
    disk_map = build_map(args.root, args.launcher_dir, args.programs_root)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(disk_map, ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"Created {output} ({output.stat().st_size:,} bytes)")


if __name__ == "__main__":
    main()
