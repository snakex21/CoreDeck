**Języki:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#budowanie)

Przenośny zestaw narzędzi dla Windows do porządkowania programów, sterowników, instalatorów, tweaków systemu, profili konfiguracji i lokalnej dokumentacji. Wersja **2.4.0** · zmiany: [CHANGELOG.md](../CHANGELOG.md).

Repozytorium jest czystą bazą startową. Nie zawiera programów, instalatorów, prywatnych linków ani danych autora. Po pierwszym uruchomieniu aplikacja tworzy puste pliki w katalogu `data`, a zawartość można dodać na ekranie **Zarządzanie**.

## Spis treści

- [Funkcje](#funkcje)
- [Szybki start](#szybki-start)
- [Układ portable](#układ-portable)
- [Katalog danych](#katalog-danych)
- [Aktualizacje](#aktualizacje)
- [Język interfejsu](#język-interfejsu)
- [Budowanie](#budowanie)
- [Testy](#testy)
- [Struktura projektu](#struktura-projektu)
- [Prywatność](#prywatność)
- [Bezpieczeństwo](#bezpieczeństwo)
- [Profil WebView2](#profil-webview2)
- [Licencja](#licencja)

## Funkcje

- **Katalog** — własny katalog programów i instalatorów, warianty programów i profile konfiguracji, osobne strony sterowników oraz ustawień programów, własne strony HTML.
- **Uruchamianie** — EXE, MSI, BAT, CMD i PowerShell. Programy portable i narzędzia mają przycisk **▶ Uruchom** i startują prosto z pendrive’a (jak dwuklik w Eksploratorze, z obsługą UAC): plik wskazany w katalogu, a gdy go nie ma, jedyny pasujący `.exe` w folderze (deinstalatory, instalatory i pliki pomocnicze są pomijane, 64-bit ma pierwszeństwo). Gdy pasuje kilka plików, CoreDeck pyta, który uruchomić; gdy w folderze jest tylko archiwum lub instalator, otwiera folder.
- **Wyniki instalatorów** — kody 1602 (anulowano w instalatorze) i 1223 (odmowa UAC) dają krótkie „Instalacja anulowana”; typowe kody MSI (1603, 1618, 1619, 1638, 1639 i inne) mają zrozumiałe opisy; 3010/1641 to sukces z informacją o restarcie.
- **Okna w stylu aplikacji** — bez przeglądarkowych okien `alert/confirm/prompt`; okna i powiadomienia obsługuje się klawiaturą (Enter, Esc, Tab, powrót fokusu). Uruchomienie instalatora, skryptu lub programu nie wymaga potwierdzenia; potwierdzenie zostało tam, gdzie coś się usuwa lub zmienia system (usuwanie, tweaki, instalacja wielu programów naraz).
- **Baza Wiedzy** z kontrolą linków oraz opcjonalna własna strona Bazy Wiedzy serwowana z `data\KnowledgeBase\`.
- **Eksport i import** danych.
- **Konfigurator Windows** — tweaki i debloat, z podglądem skryptu i zatwierdzaniem (zob. [Bezpieczeństwo](#bezpieczeństwo)).
- **Dziennik** operacji administracyjnych w `data/logs/audit.jsonl`.
- **Aktualizacje** z wydań GitHub z kanałem stabilnym i testowym, weryfikacją sum kontrolnych i automatycznym przywracaniem (zob. [Aktualizacje](#aktualizacje)).
- **Interfejs** — zapamiętywanie rozmiaru i trybu okna, ciemny i jasny motyw, 30 języków interfejsu.

## Szybki start

Pobierz `CoreDeck-Portable.zip` z sekcji **Releases** i rozpakuj w dowolnym miejscu. Program jest portable i nie wymaga instalacji: uruchom `CoreDeck.exe`. Przy pierwszym uruchomieniu obok EXE powstaje katalog `data\` (zob. [Katalog danych](#katalog-danych)).

Aktualizacja ręczna zamiast wbudowanej: zamknij CoreDeck i podmień `CoreDeck.exe` (oraz `data\tools\disk-helper.exe`). Folder `data\` zostaje bez zmian.

## Układ portable

Docelowy układ na pendrivie to `<dysk>\_Launcher\CoreDeck.exe`, a biblioteka programów leży w folderach obok `_Launcher`. Ścieżki w katalogu są zapisywane względem EXE (`../…`), a domyślne ścieżki importu/eksportu, programów i sterowników wynikają z dysku, z którego działa CoreDeck. Projekt nie wymaga stałej ścieżki i można go przenieść na inny dysk lub komputer.

Gdy pendrive dostanie inną literę, ścieżka zapisana na starej literze (np. `J:\…` po zmianie na `L:`) jest szukana na dysku, z którego działa CoreDeck. Przy starcie takie ścieżki w `programs.json`, `drivers.json` i `settings.json` są jednorazowo zamieniane na względne, z kopią pliku `*.drive-<data>.bak`.

## Katalog danych

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

**Dane prywatne i ten sam EXE.** Wszystkie dane użytkownika są czytane z `data\` obok `CoreDeck.exe`. Pliki osadzone w EXE (`cmd/launcher/web/data/*.json`) to tylko puste szablony, używane wyłącznie wtedy, gdy plik w `data\` nie istnieje. Dzięki temu publiczny `CoreDeck.exe` z Releases i prywatna instalacja to ten sam plik: prywatny katalog, Baza Wiedzy (`knowledge.json`) i strona `data\KnowledgeBase\` leżą tylko w `data\`. Uszkodzony plik w `data\` jest zgłaszany jako błąd i nie jest zastępowany pustym szablonem (kolejny zapis nie nadpisze danych). Testy pilnują, żeby do osadzonych szablonów ani do `cmd/launcher/web/` nie trafiły prywatne dane.

## Aktualizacje

Ustawienia → **Aktualizacje**.

- **Kanały:** Stabilny (domyślny) i Testowy (wydania przedpremierowe).
- **Sprawdzanie:** opcjonalnie przy starcie (najwyżej raz dziennie) oraz ręcznie przyciskiem **Sprawdź aktualizacje**. Gdy jest nowsza wersja, okno w aplikacji pokazuje listę zmian i przyciski **Aktualizuj teraz**, **Później** i **Pomiń tę wersję**.
- **Pobieranie i weryfikacja:** `CoreDeck-Portable.zip` jest pobierany z wydania GitHub `snakex21/CoreDeck`; rozmiar i SHA-256 są porównywane z `SHA256SUMS.txt`. Niezgodność przerywa aktualizację i niczego nie zmienia.
- **Instalacja:** rozpakowywane są tylko `CoreDeck.exe` i `data\tools\disk-helper.exe`, do folderu roboczego `.coredeck-update\` obok EXE. Kopia poprzednich plików trafia do `.coredeck-update\backup\<wersja>\`, pliki są podmieniane przez zmianę nazwy starych na `.old`, po czym CoreDeck uruchamia się ponownie.
- **Kontrola uruchomienia:** jeśli nowa wersja nie zgłosi poprawnego startu w ciągu 60 sekund (albo się zamknie), poprzednia wersja jest automatycznie przywracana i uruchamiana ponownie. W Ustawieniach jest też przycisk **Przywróć poprzednią wersję**.
- Działa w układzie portable na pendrivie (`<dysk>\_Launcher\CoreDeck.exe`), sama aktualizacja nie wymaga uprawnień administratora i nie zmienia niczego innego w `data\`.
- Bez telemetrii: program łączy się tylko z API GitHub i pobiera plik wydania.

## Język interfejsu

Interfejs jest dostępny w 30 językach: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 i 한국어.

Język systemu Windows jest wykrywany automatycznie. Wybór języka jest na górnym pasku i ma opcję **Automatycznie**. Brakujące w tłumaczeniu teksty są pokazywane po angielsku. Opisy wydań i [CHANGELOG](../CHANGELOG.md) są po polsku i angielsku.

## Budowanie

Wymagania:

- Windows,
- Go 1.26 lub nowszy (`go.mod` przypina toolchain `go1.26.11`, pobierany automatycznie przez `go`),
- kompilator C (MinGW gcc, wymagany przez WebView),
- opcjonalnie `go-winres` (ikona i informacje o wersji w EXE),
- Node.js,
- Python.

```bat
build.bat
```

Polecenie sprawdza składnię JavaScript, uruchamia testy Go oraz tworzy lokalne `CoreDeck.exe` i `data`. Build wydania jest pomniejszony (`-trimpath -ldflags "-s -w"`), a `build.bat` zapisuje wersję, nazwę produktu i copyright w zasobach Windows. Wariant z konsolą diagnostyczną:

```bat
build.bat debug
```

Narzędzia WebView2 można włączyć bez osobnego buildu:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` tworzy archiwum wydania `dist\CoreDeck-Portable.zip`.

## Testy

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testy w przeglądarce: uruchom `python -m http.server 8765` w katalogu repo i otwórz `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (okna i powiadomienia, wynik „PASS n/n” w tytule) albo `test/web/launcher-smoke.html` (launcher ze stubem API).

CI (GitHub Actions) sprawdza składnię JS, uruchamia testy Pythona, `go vet`, `go test -race`, `govulncheck` i build.

## Struktura projektu

```text
cmd/launcher/         aplikacja i osadzony interfejs
cmd/disk-helper/      pomocnicze wyszukiwanie dysku
internal/             logika programu
scripts/              budowanie i pakowanie
test/                 testy
```

Domyślne puste dane osadzane w EXE znajdują się w `cmd/launcher/web/data/`. Dane utworzone podczas korzystania z programu są lokalne i ignorowane przez Git.

## Prywatność

Program działa lokalnie. Repozytorium nie zawiera katalogu użytkownika, historii, prywatnych stron, modeli AI ani biblioteki instalatorów. Jedyny ruch sieciowy samego CoreDeck to sprawdzanie i pobieranie aktualizacji, które łączy się wyłącznie z API GitHub i oficjalnymi wydaniami projektu na GitHub.

## Bezpieczeństwo

- Lokalny serwer interfejsu nasłuchuje tylko na `127.0.0.1`, odpowiada tylko na nagłówek `Host` `localhost:<port>` / `127.0.0.1:<port>` i nie wysyła nagłówków CORS (wszystko jest same-origin).
- Tweaki pochodzą z `data\tweaks.json` (import, edycja) i działają z uprawnieniami administratora. Przed pierwszym uruchomieniem tweaka i po każdej zmianie jego polecenia konfigurator pokazuje dokładny skrypt i wymaga potwierdzenia. Zatwierdzenia (SHA-256 typu i polecenia) są w `data\state\tweak_approvals.json`; program uruchamia polecenia z katalogu, nie z tego, co przyśle strona.
- Aktualizacja jest instalowana dopiero po zgodności rozmiaru i SHA-256 z `SHA256SUMS.txt` (zob. [Aktualizacje](#aktualizacje)).

## Profil WebView2

Od wersji 2.3.1 profil WebView2 (motyw, język, `localStorage`) jest w `data\webview\` obok EXE. Używana jest lokalna, załatana kopia `webview_go` (`third_party/webview_go`, opis w `PATCHED.md`). Przy pierwszym uruchomieniu istniejący profil z `%APPDATA%\CoreDeck.exe` jest kopiowany (bez cache); oryginał zostaje na miejscu i można go później usunąć ręcznie.

## Licencja

CoreDeck jest udostępniany na licencji MIT, zob. [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
