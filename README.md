# CoreDeck

Wersja: **2.3.1** · Licencja: [MIT](LICENSE) · Zmiany: [CHANGELOG.md](CHANGELOG.md)

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
├── KnowledgeBase/     (opcjonalnie: własna strona Bazy Wiedzy, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (profil WebView2: motyw, język, localStorage)
├── logs/
└── state/
```

### Dane prywatne i ten sam EXE

Wszystkie dane użytkownika są czytane z `data\` obok `CoreDeck.exe`.
Pliki osadzone w EXE (`cmd/launcher/web/data/*.json`) to tylko puste
szablony, używane wyłącznie wtedy, gdy plik w `data\` nie istnieje. Dzięki
temu publiczny `CoreDeck.exe` z Releases i prywatna instalacja to ten sam
plik: prywatny katalog, Baza Wiedzy (`knowledge.json`) i strona
`data\KnowledgeBase\` leżą tylko w `data\`. Uszkodzony plik w `data\` jest
zgłaszany jako błąd i nie jest zastępowany pustym szablonem (kolejny zapis
nie nadpisze danych). Testy pilnują, żeby do osadzonych szablonów ani do
`cmd/launcher/web/` nie trafiły prywatne dane.

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

## Bezpieczeństwo

- Lokalny serwer interfejsu nasłuchuje tylko na `127.0.0.1`, odpowiada tylko
  na nagłówek `Host` `localhost:<port>` / `127.0.0.1:<port>` i nie wysyła
  nagłówków CORS (wszystko jest same-origin).
- Tweaki pochodzą z `data\tweaks.json` (import, edycja) i działają
  z uprawnieniami administratora. Przed pierwszym uruchomieniem tweaka
  i po każdej zmianie jego polecenia konfigurator pokazuje dokładny skrypt
  i wymaga potwierdzenia. Zatwierdzenia (SHA-256 typu i polecenia) są
  w `data\state\tweak_approvals.json`; program uruchamia polecenia
  z katalogu, nie z tego, co przyśle strona.

## Profil WebView2

Od wersji 2.3.1 profil WebView2 (motyw, język, `localStorage`) jest
w `data\webview\` obok EXE. Używana jest lokalna, załatana kopia
`webview_go` (`third_party/webview_go`, opis w `PATCHED.md`). Przy pierwszym
uruchomieniu istniejący profil z `%APPDATA%\CoreDeck.exe` jest kopiowany
(bez cache); oryginał zostaje na miejscu i można go później usunąć ręcznie.

## Licencja

CoreDeck jest udostępniany na licencji MIT, zob. [LICENSE](LICENSE).
Copyright (c) 2026 snakex21.
