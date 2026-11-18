**Sprachen:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#erstellen)

Ein portables Windows-Toolkit zum Ordnen von Programmen, Treibern, Installern, System-Tweaks, Einrichtungsprofilen und lokaler Dokumentation. Version **2.4.0** · Änderungen: [CHANGELOG.md](../CHANGELOG.md).

Das Repository ist ein sauberer Ausgangspunkt. Es enthält keine Programme, Installer, privaten Links oder Daten des Autors. Beim ersten Start legt die Anwendung leere Dateien im Ordner `data` an; Inhalte werden auf dem Bildschirm **Verwaltung** hinzugefügt.

## Inhalt

- [Funktionen](#funktionen)
- [Schnellstart](#schnellstart)
- [Portable Ordnerstruktur](#portable-ordnerstruktur)
- [Datenordner](#datenordner)
- [Updates](#updates)
- [Sprache der Oberfläche](#sprache-der-oberfläche)
- [Erstellen](#erstellen)
- [Tests](#tests)
- [Projektstruktur](#projektstruktur)
- [Datenschutz](#datenschutz)
- [Sicherheit](#sicherheit)
- [WebView2-Profil](#webview2-profil)
- [Lizenz](#lizenz)

## Funktionen

- **Katalog** – ein eigener Katalog von Programmen und Installern, Programmvarianten und Einrichtungsprofilen, eigene Seiten für Treiber und Programmeinstellungen sowie eigene HTML-Seiten.
- **Starten** – EXE, MSI, BAT, CMD und PowerShell. Portable Programme und Werkzeuge haben eine Schaltfläche **▶ Starten** und laufen direkt vom USB-Stick (wie ein Doppelklick im Explorer, einschließlich UAC): die im Katalog festgelegte Datei, sonst die einzige passende `.exe` im Ordner (Deinstallationsprogramme, Installer und Hilfsprogramme werden übersprungen, 64-Bit wird bevorzugt). Bei mehreren Kandidaten fragt CoreDeck, welcher gestartet werden soll; gibt es nur ein Archiv oder einen Installer, wird der Ordner geöffnet.
- **Ergebnisse von Installern** – die Exit-Codes 1602 (im Installer abgebrochen) und 1223 (UAC abgelehnt) zeigen kurz „Installation abgebrochen“; gängige MSI-Codes (1603, 1618, 1619, 1638, 1639 und weitere) erhalten verständliche Meldungen; 3010/1641 gelten als Erfolg mit Hinweis auf einen Neustart.
- **Dialoge in der App** – keine Browser-Fenster `alert/confirm/prompt`; Dialoge und Benachrichtigungen lassen sich per Tastatur bedienen (Enter, Esc, Tab, Wiederherstellung des Fokus). Das Starten eines Installers, Skripts oder Programms erfordert keine Bestätigung; bestätigt werden muss nur, wenn etwas gelöscht oder das System verändert wird (Löschen, Tweaks, Installation vieler Programme auf einmal).
- **Wissensdatenbank** mit Linkprüfung und optional einer eigenen Wissensdatenbank-Website, die aus `data\KnowledgeBase\` bereitgestellt wird.
- **Export und Import** von Daten.
- **Windows-Konfigurator** – Tweaks und Debloat, mit Skriptvorschau und Freigabe (siehe [Sicherheit](#sicherheit)).
- **Audit-Protokoll** administrativer Vorgänge in `data/logs/audit.jsonl`.
- **Updates** aus GitHub-Releases mit stabilem und Testkanal, Prüfsummenverifikation und automatischem Rollback (siehe [Updates](#updates)).
- **Oberfläche** – Fenstergröße und -modus werden gespeichert, dunkles und helles Design, 30 Sprachen der Oberfläche.

## Schnellstart

Laden Sie `CoreDeck-Portable.zip` unter **Releases** herunter und entpacken Sie es an einem beliebigen Ort. Das Programm ist portabel und muss nicht installiert werden: Starten Sie `CoreDeck.exe`. Beim ersten Start wird der Ordner `data\` neben der EXE angelegt (siehe [Datenordner](#datenordner)).

Manuelles Update statt des integrierten Updaters: CoreDeck schließen und `CoreDeck.exe` (sowie `data\tools\disk-helper.exe`) ersetzen. Der Ordner `data\` bleibt unverändert.

## Portable Ordnerstruktur

Vorgesehen ist auf einem USB-Stick die Struktur `<drive>\_Launcher\CoreDeck.exe`, mit der Programmbibliothek in Ordnern neben `_Launcher`. Katalogpfade werden relativ zur EXE gespeichert (`../…`), und die Standardpfade für Import/Export, Programme und Treiber werden aus dem Laufwerk abgeleitet, von dem CoreDeck läuft. Das Projekt braucht daher keinen festen Pfad und kann auf ein anderes Laufwerk oder einen anderen Computer verschoben werden.

Bekommt der USB-Stick einen anderen Buchstaben, wird ein mit dem alten Buchstaben gespeicherter Pfad (z. B. `J:\…`, nachdem das Laufwerk zu `L:` wurde) auf dem Laufwerk gesucht, von dem CoreDeck läuft. Beim Start werden solche Pfade in `programs.json`, `drivers.json` und `settings.json` einmalig in relative Form umgeschrieben; dabei bleibt eine Kopie `*.drive-<date>.bak` erhalten.

## Datenordner

Beim ersten Start wird neben der EXE Folgendes angelegt:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (optional: eigene Wissensdatenbank-Website, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-Profil: Design, Sprache, localStorage)
├── logs/
└── state/
```

**Private Daten und eine einzige EXE.** Alle Benutzerdaten werden aus `data\` neben `CoreDeck.exe` gelesen. Die in die EXE eingebetteten Dateien (`cmd/launcher/web/data/*.json`) sind nur leere Vorlagen, die verwendet werden, wenn die Datei in `data\` nicht existiert. Dadurch sind die öffentliche `CoreDeck.exe` aus den Releases und eine private Installation dieselbe Datei: Der private Katalog, die Wissensdatenbank (`knowledge.json`) und die Website `data\KnowledgeBase\` liegen nur in `data\`. Eine beschädigte Datei in `data\` wird als Fehler gemeldet und nicht durch die leere Vorlage ersetzt (das nächste Speichern überschreibt die Daten nicht). Tests stellen sicher, dass keine privaten Daten in die eingebetteten Vorlagen oder in `cmd/launcher/web/` gelangen.

## Updates

Einstellungen → **Updates**.

- **Kanäle:** Stabil (Standard) und Test (Vorabversionen).
- **Suche:** optional eine Suche beim Start (höchstens einmal täglich) und eine Schaltfläche **Nach Updates suchen** für die manuelle Suche. Gibt es eine neuere Version, zeigt ein Fenster in der App ihr Änderungsprotokoll mit **Jetzt aktualisieren**, **Später** und **Diese Version überspringen**.
- **Download und Prüfung:** `CoreDeck-Portable.zip` wird aus dem GitHub-Release von `snakex21/CoreDeck` heruntergeladen; Größe und SHA-256 werden mit `SHA256SUMS.txt` abgeglichen. Bei einer Abweichung wird das Update abgebrochen und nichts verändert.
- **Installation:** Nur `CoreDeck.exe` und `data\tools\disk-helper.exe` werden entpackt, und zwar in den Staging-Ordner `.coredeck-update\` neben der EXE. Die bisherigen Binärdateien werden in `.coredeck-update\backup\<version>\` gesichert, die Dateien werden ausgetauscht, indem die alten in `.old` umbenannt werden, und CoreDeck startet neu.
- **Startprüfung:** Meldet die neue Version nicht innerhalb von 60 Sekunden einen korrekten Start (oder beendet sie sich), wird die vorherige Version automatisch wiederhergestellt und erneut gestartet. **Vorherige Version wiederherstellen** steht auch in den Einstellungen zur Verfügung.
- Funktioniert in der portablen USB-Struktur (`<drive>\_Launcher\CoreDeck.exe`), benötigt für das Update selbst keine Administratorrechte und ändert nie etwas anderes in `data\`.
- Keine Telemetrie: Kontaktiert werden nur die GitHub-API und der Release-Download.

## Sprache der Oberfläche

Die Oberfläche ist in 30 Sprachen verfügbar: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 und 한국어.

Die Anzeigesprache von Windows wird automatisch erkannt. Die Sprachauswahl befindet sich in der oberen Leiste und bietet die Option **Automatisch**. Texte, die in einer Übersetzung fehlen, werden auf Englisch angezeigt. Release-Notes und das [CHANGELOG](../CHANGELOG.md) sind auf Polnisch und Englisch verfasst.

## Erstellen

Voraussetzungen:

- Windows,
- Go 1.26 oder neuer (`go.mod` legt die Toolchain `go1.26.11` fest, die `go` automatisch herunterlädt),
- ein C-Compiler (MinGW gcc, von WebView benötigt),
- optional `go-winres` (Symbol und Versionsinformationen in der EXE),
- Node.js,
- Python.

```bat
build.bat
```

Dies prüft die JavaScript-Syntax, führt die Go-Tests aus und erzeugt lokal `CoreDeck.exe` und `data`. Release-Builds werden gestrippt (`-trimpath -ldflags "-s -w"`), und `build.bat` schreibt Version, Produktname und Copyright in die Windows-Ressourcen. Variante mit Diagnosekonsole:

```bat
build.bat debug
```

Die WebView2-Entwicklertools lassen sich ohne separaten Build aktivieren:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` erstellt das Release-Archiv `dist\CoreDeck-Portable.zip`.

## Tests

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Browsertests: Führen Sie im Repository-Ordner `python -m http.server 8765` aus und öffnen Sie `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (Dialoge und Benachrichtigungen, „PASS n/n“ im Titel) oder `test/web/launcher-smoke.html` (der Launcher mit einer Stub-API).

Die CI (GitHub Actions) führt die JS-Syntaxprüfung, die Python-Tests, `go vet`, `go test -race`, `govulncheck` und den Build aus.

## Projektstruktur

```text
cmd/launcher/         Anwendung und eingebettete Oberfläche
cmd/disk-helper/      Hilfsprogramm zur Laufwerkssuche
internal/             Programmlogik
scripts/              Build und Paketierung
test/                 Tests
```

Die in die EXE eingebetteten leeren Standarddaten liegen in `cmd/launcher/web/data/`. Daten, die bei der Nutzung des Programms entstehen, sind lokal und werden von Git ignoriert.

## Datenschutz

Das Programm arbeitet lokal. Das Repository enthält keinen Benutzerkatalog, keinen Verlauf, keine privaten Seiten, keine KI-Modelle und keine Installer-Bibliothek. Der einzige Netzwerkzugriff von CoreDeck selbst ist die Suche nach Updates und deren Download; dabei werden nur die GitHub-API und die offiziellen GitHub-Releases des Projekts kontaktiert.

## Sicherheit

- Der lokale Server der Oberfläche lauscht nur auf `127.0.0.1`, antwortet nur auf den `Host`-Header `localhost:<port>` / `127.0.0.1:<port>` und sendet keine CORS-Header (alles ist same-origin).
- Tweaks stammen aus `data\tweaks.json` (Import, Bearbeitung) und laufen mit Administratorrechten. Bevor ein Tweak zum ersten Mal ausgeführt wird und nach jeder Änderung seines Befehls zeigt der Konfigurator das genaue Skript an und fragt nach einer Bestätigung. Freigaben (SHA-256 von Typ und Befehl) werden in `data\state\tweak_approvals.json` gespeichert; das Programm führt Befehle aus dem Katalog aus, nicht das, was die Seite sendet.
- Updates werden erst installiert, wenn Größe und SHA-256 mit `SHA256SUMS.txt` übereinstimmen (siehe [Updates](#updates)).

## WebView2-Profil

Seit Version 2.3.1 liegt das WebView2-Profil (Design, Sprache, `localStorage`) in `data\webview\` neben der EXE. Verwendet wird eine lokale, gepatchte Kopie von `webview_go` (`third_party/webview_go`, beschrieben in `PATCHED.md`). Beim ersten Start wird ein vorhandenes Profil aus `%APPDATA%\CoreDeck.exe` kopiert (ohne Cache); das Original bleibt an seinem Platz und kann später manuell gelöscht werden.

## Lizenz

CoreDeck wird unter der MIT-Lizenz veröffentlicht, siehe [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
