# Changelog

## 2.3.0 (2026-11-14)

### PL

- **Licencja MIT** (`LICENSE`), informacja o licencji w README i w zasobach EXE.
- **Import/eksport**: pominięcie lub zduplikowanie importu nie usuwa już plików,
  a migawki eksportu są oddzielone od aktualnych katalogów (poprawki z gałęzi
  głównej, pierwszy raz w wydaniu).
- **Przenośność**: domyślne ścieżki importu/eksportu, programów i sterowników
  są liczone z dysku, z którego działa `CoreDeck.exe` (wcześniej na sztywno
  `L:\`). Nowe API `get_portable_paths`, `generate_disk_map.py` też liczy dysk
  z położenia launchera. Testy zmiany litery dysku (Go i Python).
- **Build ze świeżego klonu**: `.gitignore` ignoruje już tylko `/data/` obok
  EXE, więc puste katalogi startowe w `cmd/launcher/web/data/` są w repozytorium;
  błędy seedowania danych trafiają do logu.
- **Bezpieczeństwo i rozmiar**: toolchain Go 1.26.11 (22 podatności stdlib
  z 1.26.2 naprawione), build z `-trimpath -ldflags "-s -w"` (mniejszy EXE).
- **Wersja w EXE**: `build.bat` wpisuje wersję, nazwę produktu i copyright do
  zasobów Windows (Właściwości pliku → Szczegóły).
- **CI**: GitHub Actions (składnia JS, testy Python, `go vet`, `go test -race`,
  `govulncheck`, build).
- Znane ograniczenie: profil WebView2 nadal jest w `%APPDATA%\CoreDeck.exe`
  (zob. README, „Znane ograniczenia”).

### EN

- **MIT license** (`LICENSE`), mentioned in the README and in the EXE resources.
- **Import/export**: skipping or duplicating an import no longer deletes files,
  and export snapshots are kept separate from the live catalogues (fixes from
  main, released for the first time).
- **Portability**: default import/export, programs and drivers paths are derived
  from the drive `CoreDeck.exe` runs from (previously hardcoded `L:\`). New
  `get_portable_paths` API; `generate_disk_map.py` also derives the drive from
  the launcher location. Drive-letter change tests (Go and Python).
- **Fresh-clone build**: `.gitignore` now ignores only `/data/` next to the EXE,
  so the empty seed catalogues in `cmd/launcher/web/data/` are tracked; seeding
  errors are logged.
- **Security and size**: Go 1.26.11 toolchain (fixes 22 reachable stdlib
  vulnerabilities present in 1.26.2), `-trimpath -ldflags "-s -w"` builds
  (smaller EXE).
- **Version in EXE**: `build.bat` writes version, product name and copyright into
  the Windows resources (File properties → Details).
- **CI**: GitHub Actions (JS syntax, Python tests, `go vet`, `go test -race`,
  `govulncheck`, build).
- Known limitation: the WebView2 profile still lives in `%APPDATA%\CoreDeck.exe`
  (see README, "Znane ograniczenia").
