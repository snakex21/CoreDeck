**Languages:** [English](README.md) · [Polski](readme/README.pl.md) · [Български](readme/README.bg.md) · [Čeština](readme/README.cs.md) · [Dansk](readme/README.da.md) · [Deutsch](readme/README.de.md) · [Ελληνικά](readme/README.el.md) · [Español](readme/README.es.md) · [Eesti](readme/README.et.md) · [Suomi](readme/README.fi.md) · [Français](readme/README.fr.md) · [Hrvatski](readme/README.hr.md) · [Magyar](readme/README.hu.md) · [Italiano](readme/README.it.md) · [Lietuvių](readme/README.lt.md) · [Latviešu](readme/README.lv.md) · [Norsk bokmål](readme/README.nb.md) · [Nederlands](readme/README.nl.md) · [Português (Brasil)](readme/README.pt-BR.md) · [Română](readme/README.ro.md) · [Русский](readme/README.ru.md) · [Slovenčina](readme/README.sk.md) · [Slovenščina](readme/README.sl.md) · [Srpski](readme/README.sr-Latn.md) · [Svenska](readme/README.sv.md) · [Türkçe](readme/README.tr.md) · [Українська](readme/README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#build)

A portable Windows toolkit for organising apps, drivers, installers, system tweaks, setup profiles and local documentation. Version **2.4.0** · changes: [CHANGELOG.md](CHANGELOG.md).

The repository is a clean starting point. It contains no programs, installers, private links or author data. On first start the application creates empty files in the `data` folder; content is added on the **Management** screen.

## Contents

- [Features](#features)
- [Quick start](#quick-start)
- [Portable layout](#portable-layout)
- [Data folder](#data-folder)
- [Updates](#updates)
- [Interface language](#interface-language)
- [Build](#build)
- [Tests](#tests)
- [Project structure](#project-structure)
- [Privacy](#privacy)
- [Security](#security)
- [WebView2 profile](#webview2-profile)
- [License](#license)

## Features

- **Catalogue** — your own catalogue of programs and installers, program variants and setup profiles, separate pages for drivers and program settings, your own HTML pages.
- **Running** — EXE, MSI, BAT, CMD and PowerShell. Portable programs and tools have a **▶ Run** button and start in place from the pendrive (like a double-click in Explorer, UAC included): the file set in the catalogue, otherwise the single matching `.exe` in the folder (uninstallers, installers and helpers are skipped, 64-bit preferred). With several candidates CoreDeck asks which one to run; with only an archive or an installer it opens the folder.
- **Installer results** — exit codes 1602 (cancelled in the installer) and 1223 (UAC declined) show a short "Installation cancelled"; common MSI codes (1603, 1618, 1619, 1638, 1639 and more) get readable messages; 3010/1641 are success with a restart note.
- **In-app dialogs** — no browser `alert/confirm/prompt` boxes; dialogs and toasts work with the keyboard (Enter, Esc, Tab, focus restore). Running an installer, script or program needs no confirmation; confirmation stays where something is deleted or the system is changed (deletions, tweaks, installing many programs at once).
- **KnowledgeBase** with link checking, and an optional own KnowledgeBase site served from `data\KnowledgeBase\`.
- **Export and import** of data.
- **Windows configurator** — tweaks and debloat, with script preview and approval (see [Security](#security)).
- **Audit log** of administrative operations in `data/logs/audit.jsonl`.
- **Updates** from GitHub releases with stable and test channels, checksum verification and automatic rollback (see [Updates](#updates)).
- **Interface** — window size and mode are remembered, dark and light theme, 30 interface languages.

## Quick start

Download `CoreDeck-Portable.zip` from **Releases** and unpack it anywhere. The program is portable and needs no installation: run `CoreDeck.exe`. On first start the `data\` folder is created next to the exe (see [Data folder](#data-folder)).

To update manually instead of using the built-in updater: close CoreDeck and replace `CoreDeck.exe` (and `data\tools\disk-helper.exe`). The `data\` folder is left as is.

## Portable layout

The intended layout on a pendrive is `<drive>\_Launcher\CoreDeck.exe`, with the program library in folders next to `_Launcher`. Catalogue paths are stored relative to the exe (`../…`), and default import/export, programs and drivers paths are derived from the drive CoreDeck runs from, so the project needs no fixed path and can be moved to another drive or computer.

If the pendrive gets a different letter, a path saved on the old letter (e.g. `J:\…` after the drive became `L:`) is looked up on the drive CoreDeck runs from. On start such paths in `programs.json`, `drivers.json` and `settings.json` are rewritten once to relative form, keeping a `*.drive-<date>.bak` copy.

## Data folder

On first start the following is created next to the exe:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (optional: your own KnowledgeBase site, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2 profile: theme, language, localStorage)
├── logs/
└── state/
```

**Private data and one exe.** All user data is read from `data\` next to `CoreDeck.exe`. The files embedded in the exe (`cmd/launcher/web/data/*.json`) are only empty templates, used when the file in `data\` does not exist. This makes the public `CoreDeck.exe` from Releases and a private installation the same file: the private catalogue, the KnowledgeBase (`knowledge.json`) and the `data\KnowledgeBase\` site live only in `data\`. A corrupt file in `data\` is reported as an error and is not replaced by the empty template (the next save will not overwrite the data). Tests make sure no private data ends up in the embedded templates or in `cmd/launcher/web/`.

## Updates

Settings → **Updates**.

- **Channels:** Stable (default) and Test (pre-releases).
- **Checking:** optional check on start (at most once a day) and a manual **Check for updates** button. When a newer version exists, an in-app window shows its changelog with **Update now**, **Later** and **Skip this version**.
- **Download and verification:** `CoreDeck-Portable.zip` is downloaded from the GitHub release of `snakex21/CoreDeck`; its size and SHA-256 are checked against `SHA256SUMS.txt`. A mismatch aborts the update and nothing is changed.
- **Installation:** only `CoreDeck.exe` and `data\tools\disk-helper.exe` are unpacked, into the staging folder `.coredeck-update\` next to the exe. The previous binaries are backed up in `.coredeck-update\backup\<version>\`, the files are swapped by renaming the old ones to `.old`, and CoreDeck restarts.
- **Health check:** if the new version does not report a correct start within 60 seconds (or exits), the previous version is restored automatically and started again. **Restore the previous version** is also available in Settings.
- Works in the portable USB layout (`<drive>\_Launcher\CoreDeck.exe`), needs no administrator rights for the update itself, and never changes anything else in `data\`.
- No telemetry: only the GitHub API and the release download are contacted.

## Interface language

The interface is available in 30 languages: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 and 한국어.

The Windows display language is detected automatically. The language picker is in the top bar and has an **Automatic** option. Text that is missing in a translation falls back to English. Release notes and the [CHANGELOG](CHANGELOG.md) are in Polish and English.

## Build

Requirements:

- Windows,
- Go 1.26 or newer (`go.mod` pins the `go1.26.11` toolchain, downloaded automatically by `go`),
- a C compiler (MinGW gcc, required by WebView),
- optionally `go-winres` (icon and version information in the exe),
- Node.js,
- Python.

```bat
build.bat
```

This checks JavaScript syntax, runs the Go tests and produces a local `CoreDeck.exe` and `data`. Release builds are stripped (`-trimpath -ldflags "-s -w"`), and `build.bat` writes the version, product name and copyright into the Windows resources. Variant with a diagnostic console:

```bat
build.bat debug
```

WebView2 developer tools can be enabled without a separate build:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` builds the release archive `dist\CoreDeck-Portable.zip`.

## Tests

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Browser tests: run `python -m http.server 8765` in the repository folder and open `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialogs and toasts, "PASS n/n" in the title) or `test/web/launcher-smoke.html` (the launcher with a stubbed API).

CI (GitHub Actions) runs the JS syntax check, the Python tests, `go vet`, `go test -race`, `govulncheck` and the build.

## Project structure

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

The default empty data embedded in the exe is in `cmd/launcher/web/data/`. Data created while using the program is local and ignored by Git.

## Privacy

The program works locally. The repository contains no user catalogue, history, private pages, AI models or installer library. The only network access of CoreDeck itself is the update check and download, which contact only the GitHub API and the official GitHub releases of the project.

## Security

- The local interface server listens only on `127.0.0.1`, answers only to the `Host` header `localhost:<port>` / `127.0.0.1:<port>` and sends no CORS headers (everything is same-origin).
- Tweaks come from `data\tweaks.json` (import, editing) and run with administrator rights. Before a tweak runs for the first time, and after every change of its command, the configurator shows the exact script and asks for confirmation. Approvals (SHA-256 of the type and command) are stored in `data\state\tweak_approvals.json`; the program runs commands from the catalogue, not whatever the page sends.
- Updates are installed only after the size and SHA-256 match `SHA256SUMS.txt` (see [Updates](#updates)).

## WebView2 profile

Since version 2.3.1 the WebView2 profile (theme, language, `localStorage`) lives in `data\webview\` next to the exe. A local, patched copy of `webview_go` is used (`third_party/webview_go`, described in `PATCHED.md`). On first start an existing profile from `%APPDATA%\CoreDeck.exe` is copied (without the cache); the original stays in place and can be deleted manually later.

## License

CoreDeck is released under the MIT license, see [LICENSE](LICENSE).
Copyright (c) 2026 snakex21.
