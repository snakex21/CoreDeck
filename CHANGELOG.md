# Changelog

## 2.4.0 (2026-11-19)

### PL

- **Aktualizacje w aplikacji**: CoreDeck sam pobiera i instaluje nowe wersje.
  Kanały: stabilny (domyślny) i testowy (wersje przedpremierowe), wybór
  w Ustawieniach → Aktualizacje, razem z opcją sprawdzania przy
  uruchomieniu (najwyżej raz dziennie). Okno nowej wersji pokazuje listę
  zmian i przyciski „Aktualizuj teraz”, „Później” i „Pomiń tę wersję”.
- **Bezpieczna podmiana**: paczka `CoreDeck-Portable.zip` z wydania na GitHubie
  jest sprawdzana rozmiarem i sumą SHA-256 z `SHA256SUMS.txt` (niezgodność
  przerywa aktualizację bez żadnych zmian). Podmieniane są tylko
  `CoreDeck.exe` i `data\tools\disk-helper.exe`; poprzednie wersje zostają
  jako `.old` i kopia w `.coredeck-update\backup\`. Jeśli nowa wersja nie
  uruchomi się poprawnie w ciągu 60 sekund, poprzednia wraca automatycznie.
  Przycisk „Przywróć poprzednią wersję” w Ustawieniach. Działa na pendrivie,
  bez uprawnień administratora; reszta folderu `data\` nie jest zmieniana.
- **30 języków interfejsu**: doszły bułgarski, duński, grecki, estoński,
  fiński, chorwacki, węgierski, litewski, łotewski, norweski, rumuński,
  słowacki, słoweński, serbski (łacinka) i ukraiński. Wszystkie języki mają
  komplet tekstów (także okna, powiadomienia, kody instalatorów, uruchamianie
  programów portable i aktualizacje) oraz poprawne formy liczby mnogiej.
  Wybór języka „Automatycznie” bierze język systemu; brakujący tekst pokazuje
  się po angielsku. Opisy tweaków po czesku działają (wcześniej pokazywały
  się po angielsku).
- **README w 27 językach** (`readme\README.<język>.md`) z sekcjami
  o aktualizacjach i językach.
- **Teksty interfejsu jako klucze tłumaczeń**: napisy w HTML i JS (także
  własne okna i powiadomienia aplikacji) są kluczami tłumaczeń, zdania nie są
  sklejane z kawałków, liczby mają formy liczby mnogiej. Komunikaty serwera
  mają stałe kody (`internal/uimsg`), które interfejs tłumaczy na wybrany
  język; polski tekst zostaje w logach.
- **Poprawka**: `release-manifest.json` w paczce ma poprawne polskie znaki
  (skrypt `scripts\package.ps1` czyta i zapisuje UTF-8 jawnie).
- **Testy**: zła suma SHA-256 i zły rozmiar przerywają aktualizację, podmiana
  i wycofanie na makiecie, kontrola zdrowia po restarcie, kanały stabilny
  i testowy, komplet kluczy i form liczby mnogiej we wszystkich językach,
  brak zaszytych tekstów w UI i w komunikatach serwera, słownik terminów
  (`docs\i18n-glossary.md`) i znaki diakrytyczne każdego języka.

Aktualizacja z 2.3.x: zamknij CoreDeck i podmień `CoreDeck.exe` (oraz
`data\tools\disk-helper.exe`). Kolejne wersje zainstalujesz już z aplikacji.

### EN

- **In-app updates**: CoreDeck downloads and installs new versions itself.
  Channels: stable (default) and test (pre-releases), chosen in Settings →
  Updates together with an optional check on start (at most once a day). The
  new-version window shows the changelog with "Update now", "Later" and
  "Skip this version".
- **Safe swap**: the `CoreDeck-Portable.zip` package from the GitHub release is
  checked against its size and the SHA-256 in `SHA256SUMS.txt` (a mismatch
  stops the update with nothing changed). Only `CoreDeck.exe` and
  `data\tools\disk-helper.exe` are replaced; the previous versions stay as
  `.old` plus a copy in `.coredeck-update\backup\`. If the new version does
  not start correctly within 60 seconds, the previous one comes back
  automatically. "Restore the previous version" in Settings. Works on a USB
  drive without administrator rights; the rest of `data\` is not touched.
- **30 interface languages**: added Bulgarian, Danish, Greek, Estonian,
  Finnish, Croatian, Hungarian, Lithuanian, Latvian, Norwegian, Romanian,
  Slovak, Slovenian, Serbian (Latin) and Ukrainian. Every language has every
  text (including dialogs, notifications, installer codes, portable launch and
  updates) and correct plural forms. "Automatic" follows the system language;
  missing text falls back to English. Czech tweak descriptions now work (they
  used to show in English).
- **README in 27 languages** (`readme\README.<lang>.md`) with sections on
  updates and languages.
- **Interface texts as translation keys**: texts in HTML and JS (including the
  app's own dialogs and notifications) are translation keys, sentences are no
  longer glued from fragments, counts use plural forms. Server messages carry
  stable codes (`internal/uimsg`) that the interface translates into the chosen
  language; the Polish text stays in the logs.
- **Fix**: `release-manifest.json` in the package has correct Polish
  characters (`scripts\package.ps1` reads and writes UTF-8 explicitly).
- **Tests**: a wrong SHA-256 or size stops the update, swap and rollback on a
  mock, the health check after restart, stable and test channels, every key and
  plural form in every language, no hard-coded texts in the UI or in server
  messages, the term glossary (`docs\i18n-glossary.md`) and the diacritics of
  every language.

Update from 2.3.x: close CoreDeck and replace `CoreDeck.exe` (and
`data\tools\disk-helper.exe`). Later versions install from the app.

## 2.3.2 (2026-11-18)

### PL

- **Programy portable da się uruchomić**: karty programów portable
  i narzędzi mają przycisk „▶ Uruchom”. Program startuje prosto z pendrive’a
  (jak dwuklik w Eksploratorze, z obsługą UAC): plik wskazany w katalogu,
  a gdy go nie ma, jedyny pasujący `.exe` w folderze (deinstalatory,
  instalatory i pliki pomocnicze są pomijane, 64-bit ma pierwszeństwo).
  Gdy pasuje kilka plików, CoreDeck pyta, który uruchomić; gdy w folderze
  jest tylko archiwum lub instalator, otwiera folder. Wcześniej karta
  portable miała tylko „Szczegóły”.
- **Bez okien przeglądarki**: wszystkie komunikaty „Komunikat z witryny
  localhost:8000” (alert/confirm/prompt) zastąpiły okna i powiadomienia
  w stylu aplikacji, obsługiwane klawiaturą (Enter, Esc, Tab, powrót fokusu).
  Uruchomienie instalatora, skryptu lub programu nie wymaga już
  potwierdzenia: kliknięcie jest decyzją. Potwierdzenie zostało tylko tam,
  gdzie coś się usuwa lub zmienia system (usuwanie, tweaki, instalacja
  wielu programów naraz).
- **Anulowanie instalatora to nie błąd**: kod 1602 (anulowano
  w instalatorze) i 1223 (odmowa UAC) dają krótkie „Instalacja anulowana”.
  Typowe kody instalatorów MSI mają zrozumiałe opisy (1618 trwa inna
  instalacja, 1638 wersja już zainstalowana, 1603, 1619, 1639 i inne);
  3010/1641 to sukces z informacją o restarcie. Bez „exit status 1602”.
- **Zmiana litery pendrive’a**: ścieżka zapisana na starej literze
  (np. `J:\…` po zmianie na `L:`) jest szukana na dysku, z którego działa
  CoreDeck. Przy starcie takie ścieżki w `programs.json`, `drivers.json`
  i `settings.json` są jednorazowo zamieniane na względne (`../…`), z kopią
  pliku `*.drive-<data>.bak`.
- **Poprawka**: domyślny folder instalacji to `C:\Programs`, a nie
  względne `C:Programs`.
- **Testy**: test pilnujący braku natywnych `alert/confirm/prompt` w UI,
  mapowanie kodów instalatora (z prawdziwym procesem zwracającym 1602),
  uruchamianie portable po symulowanej zmianie litery dysku, migracja
  ścieżek; przeglądarkowe testy okien (`test/web/dialogs-smoke.html`)
  i podgląd launchera ze stubem API (`test/web/launcher-smoke.html`).

Aktualizacja: zamknij CoreDeck i podmień `CoreDeck.exe` (oraz
`data\tools\disk-helper.exe`). Folder `data\` zostaje bez zmian.

### EN

- **Portable programs can be started**: portable and tool cards get a
  "▶ Run" button. The program starts in place from the pendrive (like a
  double-click in Explorer, UAC included): the file set in the catalogue,
  otherwise the single matching `.exe` in the folder (uninstallers,
  installers and helpers are skipped, 64-bit preferred). With several
  candidates CoreDeck asks which one to run; with only an archive or an
  installer it opens the folder. Portable cards used to offer Details only.
- **No browser dialogs**: every "Message from localhost:8000" box
  (alert/confirm/prompt) is replaced by in-app dialogs and toasts with
  keyboard support (Enter, Esc, Tab, focus restore). Running an installer,
  script or program no longer asks for confirmation: the click is the
  decision. Confirmation stays where something is deleted or the system is
  changed (deletions, tweaks, installing many programs at once).
- **Cancelling an installer is not an error**: exit code 1602 (cancelled
  in the installer) and 1223 (UAC declined) show a short "Installation
  cancelled". Common MSI codes get readable messages (1618 another
  installation in progress, 1638 already installed, 1603, 1619, 1639 and
  more); 3010/1641 are success with a restart note. No "exit status 1602".
- **Pendrive drive letter change**: a path saved on the old letter
  (e.g. `J:\…` after the drive became `L:`) is looked up on the drive
  CoreDeck runs from. On start such paths in `programs.json`,
  `drivers.json` and `settings.json` are rewritten once to relative form
  (`../…`), keeping a `*.drive-<date>.bak` copy.
- **Fix**: the default install folder is `C:\Programs`, not the
  drive-relative `C:Programs`.
- **Tests**: a check that the UI has no native `alert/confirm/prompt`,
  installer exit code mapping (with a real process returning 1602),
  launching a portable entry after a simulated drive letter change, path
  migration; browser tests for the dialogs (`test/web/dialogs-smoke.html`)
  and a launcher preview with a stubbed API (`test/web/launcher-smoke.html`).

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
