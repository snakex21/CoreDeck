"""Tests for generate_disk_map.py (run: python -m unittest discover -s scripts)."""

import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path, PureWindowsPath

HERE = Path(__file__).resolve().parent
_spec = importlib.util.spec_from_file_location("generate_disk_map", HERE / "generate_disk_map.py")
gdm = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(gdm)


class DeriveDefaultsTest(unittest.TestCase):
    def test_drive_letter_change_moves_every_default(self):
        before = gdm.derive_defaults(PureWindowsPath(r"L:\_Launcher"))
        after = gdm.derive_defaults(PureWindowsPath(r"J:\_Launcher"))

        self.assertEqual(before, (PureWindowsPath("L:\\"), PureWindowsPath(r"L:\Centrum Narzędzi i Oprogramowania")))
        self.assertEqual(after, (PureWindowsPath("J:\\"), PureWindowsPath(r"J:\Centrum Narzędzi i Oprogramowania")))
        for p in after:
            self.assertNotIn("L:", str(p))

    def test_no_hardcoded_drive_letter_in_defaults(self):
        source = (HERE / "generate_disk_map.py").read_text(encoding="utf-8")
        self.assertNotRegex(source, r'Path\("[A-Za-z]:')


class GenerateRunTest(unittest.TestCase):
    def test_output_root_follows_launcher_drive(self):
        with tempfile.TemporaryDirectory() as tmp:
            launcher = Path(tmp) / "_Launcher"
            launcher.mkdir()
            out = Path(tmp) / "disk_map.json"
            subprocess.run(
                [sys.executable, str(HERE / "generate_disk_map.py"), "--launcher-dir", str(launcher), "--output", str(out)],
                check=True, capture_output=True,
            )
            data = json.loads(out.read_text(encoding="utf-8"))
            drive = launcher.resolve().anchor
            self.assertEqual(data["root"], drive)
            self.assertEqual(data["launcher_dir"], str(launcher.resolve()))
            self.assertTrue(data["programs_root"].startswith(drive))


if __name__ == "__main__":
    unittest.main()
