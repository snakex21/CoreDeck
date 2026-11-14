# CoreDeck

Wersja: **2.3.0** · Licencja: [MIT](LICENSE) · Zmiany: [CHANGELOG.md](CHANGELOG.md)

Portable Windows toolkit for organizing apps, drivers, installers, system tweaks,
setup profiles and local documentation.

Repozytorium jest czystą bazą startową. Nie zawiera programów, instalatorów,
prywatnych linków ani danych autora. Po pierwszym uruchomieniu aplikacja tworzy
puste pliki w katalogu `data`, a zawartość można dodać przez ekran
**Zarządzanie**.

## Pobieranie

Najprościej pobrać gotowy `CoreDeck.exe` z sekcji **Releases**. Program jest
portable i nie wymaga instalacji.

Przy pierwszym uruchomieniu obok EXE zostaną utworzone:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── user_pages/
├── logs/
└── state/
```

## Funkcje

- własny katalog programów i instalatorów,
- uruchamianie EXE, MSI, BAT, CMD i PowerShell,
- warianty programu i profile konfiguracji,
- osobne strony sterowników oraz ustawień programów,
- własne strony HTML,
- Baza Wiedzy i kontrola linków,
- eksport oraz import danych,
- konfigurator Windows, tweaki i debloat,
- ręczne sprawdzanie oficjalnych wydań GitHub,
- dziennik operacji administracyjnych w `data/logs/audit.jsonl`,
- zapamiętywanie rozmiaru i trybu okna,
- ciemny i jasny motyw,
- wielojęzyczny interfejs.

## Budowanie

Wymagania:

- Windows,
- Go 1.26 lub nowszy (go.mod przypina toolchain `go1.26.11`, pobierany
  automatycznie przez `go`),
- kompilator C (MinGW gcc, wymagany przez WebView),
- opcjonalnie `go-winres` (ikona i informacje o wersji w EXE),
- Node.js,
- Python.

```bat
build.bat
```

Polecenie sprawdza składnię JavaScript, uruchamia testy Go oraz tworzy lokalne
`CoreDeck.exe` i `data`. Projekt nie wymaga stałej ścieżki i można go przenieść
na inny dysk lub komputer. Wariant z konsolą diagnostyczną:

```bat
build.bat debug
```

Narzędzia WebView2 można włączyć bez osobnego buildu:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

Build wydania jest pomniejszony (`-trimpath -ldflags "-s -w"`). Testy:

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

## Struktura

```text
cmd/launcher/         aplikacja i osadzony interfejs
cmd/disk-helper/      pomocnicze wyszukiwanie dysku
internal/             logika programu
scripts/              budowanie i pakowanie
test/                 testy
```

Domyślne puste dane osadzane w EXE znajdują się w
`cmd/launcher/web/data/`. Dane utworzone podczas korzystania z programu są
lokalne i ignorowane przez Git.

## Prywatność

Program działa lokalnie. Repozytorium nie zawiera katalogu użytkownika,
historii, prywatnych stron, modeli AI ani biblioteki instalatorów.
Sprawdzenie aktualizacji jest uruchamiane ręcznie i odpytuje wyłącznie
oficjalne GitHub Releases projektu.

## Znane ograniczenia

- **Profil WebView2 nie jest przenośny.** Dane aplikacji (`data/`, w tym
  `data/webview/custom_profiles.json`) leżą obok EXE, ale sam profil
  przeglądarki WebView2 (cache, `localStorage`: motyw, język, historia
  uruchomień) biblioteka `webview_go` zawsze tworzy w
  `%APPDATA%\CoreDeck.exe`. Zmienna `WEBVIEW2_USER_DATA_FOLDER` nie działa,
  bo `webview_go` używa wbudowanego loadera WebView2, który jej nie czyta.
  Na innym komputerze motyw i język trzeba więc ustawić ponownie. Poprawka
  wymaga łatki w `webview_go` (parametr `userDataFolder`) albo dołączenia
  `WebView2Loader.dll`; jest planowana na kolejną wersję.

## Licencja

CoreDeck jest udostępniany na licencji MIT, zob. [LICENSE](LICENSE).
Copyright (c) 2026 snakex21.
