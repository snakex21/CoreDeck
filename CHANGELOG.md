# Changelog

## 2.3.1 (2026-11-14)

### PL

- **Karty programów**: jeden główny przycisk (uruchom instalator/skrypt,
  a gdy nie ma czego uruchomić: Szczegóły). „Otwórz folder” i „Kopiuj
  ścieżkę” to ikony pokazywane po najechaniu i przy fokusie klawiatury.
  Równe karty: tytuł w jednej linii, opis przycięty do dwóch linii,
  przyciski na dole karty.
- **Ikony zastępcze z inicjałami** dla programów bez ikony albo z ikoną
  nieczytelną na ciemnej karcie (np. Pinokio, DirectX, Spotube).
- **Górny pasek**: szeroka wyszukiwarka na środku (bez uciętej
  podpowiedzi), jednolite przyciski nawigacji. Na szerokich ekranach treść
  ma maksymalną szerokość, lista Import/Eksport jest w kolumnach i gęstsza.
- **Te same EXE dla wszystkich**: katalog programów, tweaki, sterowniki,
  ustawienia i Baza Wiedzy są czytane tylko z `data\` obok EXE; osadzone
  pliki to puste szablony. Strona Bazy Wiedzy jest serwowana
  z `data\KnowledgeBase\`. Uszkodzony plik w `data\` nie jest już po cichu
  zastępowany pustym szablonem.
- **Przenośny profil WebView2** w `data\webview\` (motyw, język). Istniejący
  profil z `%APPDATA%\CoreDeck.exe` jest kopiowany przy pierwszym starcie,
  oryginał zostaje.
- **Bezpieczeństwo**: podgląd skryptu i wyraźne potwierdzenie przed
  uruchomieniem nowego lub zmienionego (np. zaimportowanego) tweaka; program
  uruchamia polecenia z katalogu, nie ze strony. Lokalny serwer bez
  `Access-Control-Allow-Origin: *`, tylko same-origin i nagłówek `Host`
  localhost.
- **Kod**: jedna funkcja `paths.ResolvePath` zamiast trzech, z testami.
  Zakładki konfiguratora obsługują klawiaturę.

Aktualizacja: zamknij CoreDeck i podmień `CoreDeck.exe` (oraz
`data\tools\disk-helper.exe`). Folder `data\` zostaje bez zmian.

### EN

- **Program cards**: one primary button (run installer/script, otherwise
  Details). Open folder and Copy path are icon buttons shown on hover and on
  keyboard focus. Equal cards: one-line title, two-line description, actions
  pinned to the bottom.
- **Initials fallback icons** for programs without an icon or with one that
  is unreadable on the dark card (e.g. Pinokio, DirectX, Spotube).
- **Top bar**: wide central search (no truncated placeholder), consistent
  navigation buttons. Wide screens get a maximum content width; the
  Import/Export list is a denser multi-column grid.
- **One EXE for everyone**: programs, tweaks, drivers, settings and the
  KnowledgeBase are read only from `data\` next to the exe; the embedded
  files are empty templates. The KnowledgeBase site is served from
  `data\KnowledgeBase\`. A corrupt file in `data\` is no longer silently
  replaced by the empty template.
- **Portable WebView2 profile** in `data\webview\` (theme, language). An
  existing `%APPDATA%\CoreDeck.exe` profile is copied on first start; the
  original is left in place.
- **Security**: script preview and explicit confirmation before a new or
  changed (e.g. imported) tweak runs; commands are taken from the catalogue,
  not from the page. Local server without `Access-Control-Allow-Origin: *`,
  same-origin only, localhost `Host` header only.
- **Code**: a single `paths.ResolvePath` instead of three resolvers, with
  tests. Configurator tabs are keyboard accessible.

Update: close CoreDeck and replace `CoreDeck.exe` (and
`data\tools\disk-helper.exe`). The `data\` folder is left as is.

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
