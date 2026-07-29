# CoreDeck

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
- ciemny i jasny motyw,
- wielojęzyczny interfejs.

## Budowanie

Wymagania:

- Windows,
- Go,
- Node.js,
- Python.

```bat
build.bat
```

Polecenie sprawdza składnię JavaScript, uruchamia testy Go i tworzy
`CoreDeck.exe`. Wariant z konsolą diagnostyczną:

```bat
build.bat debug
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
