"""package.ps1 must produce a UTF-8 release-manifest.json (PowerShell 5 safe)."""
import pathlib
import unittest

SCRIPT = pathlib.Path(__file__).with_name("package.ps1")


class PackageEncodingTest(unittest.TestCase):
    def test_script_with_non_ascii_text_has_bom(self):
        data = SCRIPT.read_bytes()
        if any(b > 127 for b in data):
            # Without a BOM Windows PowerShell 5 parses the script as ANSI and
            # Polish letters in string literals end up as mojibake.
            self.assertTrue(data.startswith(b"\xef\xbb\xbf"), "package.ps1 needs a UTF-8 BOM")

    def test_manifest_written_as_utf8_without_bom(self):
        text = SCRIPT.read_text(encoding="utf-8-sig")
        self.assertIn("UTF8Encoding($false)", text)
        self.assertIn("WriteAllText((Join-Path $packageRoot \"release-manifest.json\")", text)
        self.assertNotIn("release-manifest.json\") -Encoding", text)

    def test_sha256sums_lists_updater_assets(self):
        # The in-app updater refuses a release whose SHA256SUMS.txt lacks these.
        text = SCRIPT.read_text(encoding="utf-8-sig")
        self.assertIn('"SHA256SUMS.txt"', text)
        self.assertIn("{0}  CoreDeck.exe", text)
        self.assertIn("{0}  CoreDeck-Portable.zip", text)


if __name__ == "__main__":
    unittest.main()
