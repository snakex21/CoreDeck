**Språk:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#bygging)

En portabel verktøykasse for Windows for å organisere programmer, drivere, installasjonsprogrammer, systemjusteringer, oppsettsprofiler og lokal dokumentasjon. Versjon **2.4.0** · endringer: [CHANGELOG.md](../CHANGELOG.md).

Repositoriet er et rent utgangspunkt. Det inneholder ingen programmer, installasjonsprogrammer, private lenker eller forfatterdata. Ved første oppstart oppretter programmet tomme filer i mappen `data`; innhold legges til på skjermen **Administrasjon**.

## Innhold

- [Funksjoner](#funksjoner)
- [Hurtigstart](#hurtigstart)
- [Portabel struktur](#portabel-struktur)
- [Datamappe](#datamappe)
- [Oppdateringer](#oppdateringer)
- [Grensesnittspråk](#grensesnittspråk)
- [Bygging](#bygging)
- [Tester](#tester)
- [Prosjektstruktur](#prosjektstruktur)
- [Personvern](#personvern)
- [Sikkerhet](#sikkerhet)
- [WebView2-profil](#webview2-profil)
- [Lisens](#lisens)

## Funksjoner

- **Katalog** — din egen katalog over programmer og installasjonsprogrammer, programvarianter og oppsettsprofiler, egne sider for drivere og programinnstillinger, dine egne HTML-sider.
- **Kjøring** — EXE, MSI, BAT, CMD og PowerShell. Portable programmer og verktøy har en **▶ Kjør**-knapp og startes direkte fra minnepinnen (som et dobbeltklikk i Filutforsker, UAC inkludert): filen som er angitt i katalogen, ellers den eneste passende `.exe` i mappen (avinstallasjonsprogrammer, installasjonsprogrammer og hjelpeprogrammer hoppes over, 64-biters foretrekkes). Med flere kandidater spør CoreDeck hvilken som skal kjøres; med bare et arkiv eller et installasjonsprogram åpnes mappen.
- **Resultater fra installasjonsprogrammer** — avslutningskodene 1602 (avbrutt i installasjonsprogrammet) og 1223 (UAC avvist) viser en kort melding "Installasjonen ble avbrutt"; vanlige MSI-koder (1603, 1618, 1619, 1638, 1639 m.fl.) får lettleste meldinger; 3010/1641 regnes som vellykket med en merknad om omstart.
- **Dialoger i appen** — ingen `alert/confirm/prompt`-bokser fra nettleseren; dialoger og varsler kan brukes med tastaturet (Enter, Esc, Tab, gjenoppretting av fokus). Kjøring av et installasjonsprogram, et skript eller et program krever ingen bekreftelse; bekreftelse beholdes der noe slettes eller systemet endres (slettinger, justeringer, installasjon av mange programmer samtidig).
- **KnowledgeBase** med lenkekontroll, og et valgfritt eget KnowledgeBase-nettsted som serveres fra `data\KnowledgeBase\`.
- **Eksport og import** av data.
- **Windows-konfigurator** — justeringer og debloat, med forhåndsvisning av skript og godkjenning (se [Sikkerhet](#sikkerhet)).
- **Revisjonslogg** over administrative operasjoner i `data/logs/audit.jsonl`.
- **Oppdateringer** fra GitHub-utgivelser med stabil kanal og testkanal, verifisering av kontrollsummer og automatisk tilbakerulling (se [Oppdateringer](#oppdateringer)).
- **Grensesnitt** — vindusstørrelse og -modus huskes, mørkt og lyst tema, 30 grensesnittspråk.

## Hurtigstart

Last ned `CoreDeck-Portable.zip` fra **Releases** og pakk den ut hvor som helst. Programmet er portabelt og krever ingen installasjon: kjør `CoreDeck.exe`. Ved første oppstart opprettes mappen `data\` ved siden av exe-filen (se [Datamappe](#datamappe)).

Slik oppdaterer du manuelt i stedet for med den innebygde oppdateringen: lukk CoreDeck og erstatt `CoreDeck.exe` (og `data\tools\disk-helper.exe`). Mappen `data\` forblir uendret.

## Portabel struktur

Den tiltenkte strukturen på en minnepinne er `<stasjon>\_Launcher\CoreDeck.exe`, med programbiblioteket i mapper ved siden av `_Launcher`. Stier i katalogen lagres relativt til exe-filen (`../…`), og standardstiene for import/eksport, programmer og drivere utledes fra stasjonen CoreDeck kjører fra, så prosjektet trenger ingen fast sti og kan flyttes til en annen stasjon eller datamaskin.

Hvis minnepinnen får en annen bokstav, slås en sti som er lagret med den gamle bokstaven (f.eks. `J:\…` etter at stasjonen har blitt `L:`), opp på stasjonen CoreDeck kjører fra. Ved oppstart skrives slike stier i `programs.json`, `drivers.json` og `settings.json` om én gang til relativ form, og en kopi `*.drive-<dato>.bak` beholdes.

## Datamappe

Ved første oppstart opprettes følgende ved siden av exe-filen:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (valgfritt: ditt eget KnowledgeBase-nettsted, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-profil: tema, språk, localStorage)
├── logs/
└── state/
```

**Private data og én exe-fil.** Alle brukerdata leses fra `data\` ved siden av `CoreDeck.exe`. Filene som er innebygd i exe-filen (`cmd/launcher/web/data/*.json`), er bare tomme maler som brukes når filen i `data\` ikke finnes. Dermed er den offentlige `CoreDeck.exe` fra Releases og en privat installasjon den samme filen: den private katalogen, KnowledgeBase (`knowledge.json`) og nettstedet `data\KnowledgeBase\` finnes bare i `data\`. En skadet fil i `data\` rapporteres som en feil og erstattes ikke av den tomme malen (neste lagring overskriver ikke dataene). Tester sørger for at ingen private data havner i de innebygde malene eller i `cmd/launcher/web/`.

## Oppdateringer

Innstillinger → **Oppdateringer**.

- **Kanaler:** Stabil (standard) og Test (forhåndsversjoner).
- **Søk:** valgfritt søk ved oppstart (høyst én gang om dagen) og en manuell knapp **Se etter oppdateringer**. Når det finnes en nyere versjon, viser et vindu i appen endringsloggen med **Oppdater nå**, **Senere** og **Hopp over denne versjonen**.
- **Nedlasting og verifisering:** `CoreDeck-Portable.zip` lastes ned fra GitHub-utgivelsen av `snakex21/CoreDeck`; størrelsen og SHA-256 kontrolleres mot `SHA256SUMS.txt`. Ved avvik avbrytes oppdateringen, og ingenting endres.
- **Installasjon:** bare `CoreDeck.exe` og `data\tools\disk-helper.exe` pakkes ut, til mellomlagringsmappen `.coredeck-update\` ved siden av exe-filen. De forrige binærfilene sikkerhetskopieres i `.coredeck-update\backup\<version>\`, filene byttes ved at de gamle får nytt navn til `.old`, og CoreDeck starter på nytt.
- **Helsesjekk:** hvis den nye versjonen ikke melder en korrekt oppstart innen 60 sekunder (eller avslutter), gjenopprettes forrige versjon automatisk og startes igjen. **Gjenopprett forrige versjon** finnes også i Innstillinger.
- Fungerer i den portable USB-strukturen (`<stasjon>\_Launcher\CoreDeck.exe`), krever ikke administratorrettigheter for selve oppdateringen og endrer aldri noe annet i `data\`.
- Ingen telemetri: bare GitHub API og nedlastingen av utgivelsen kontaktes.

## Grensesnittspråk

Grensesnittet er tilgjengelig på 30 språk: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 og 한국어.

Visningsspråket i Windows oppdages automatisk. Språkvelgeren ligger i den øverste linjen og har et **Automatisk**-valg. Tekst som mangler i en oversettelse, vises på engelsk. Utgivelsesnotater og [CHANGELOG](../CHANGELOG.md) er på polsk og engelsk.

## Bygging

Krav:

- Windows,
- Go 1.26 eller nyere (`go.mod` låser verktøykjeden `go1.26.11`, som `go` laster ned automatisk),
- en C-kompilator (MinGW gcc, kreves av WebView),
- valgfritt `go-winres` (ikon og versjonsinformasjon i exe-filen),
- Node.js,
- Python.

```bat
build.bat
```

Dette kontrollerer JavaScript-syntaksen, kjører Go-testene og lager en lokal `CoreDeck.exe` og `data`. Release-bygg er strippet (`-trimpath -ldflags "-s -w"`), og `build.bat` skriver versjon, produktnavn og opphavsrett inn i Windows-ressursene. Variant med diagnostikkonsoll:

```bat
build.bat debug
```

Utviklerverktøyene i WebView2 kan slås på uten et eget bygg:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` bygger utgivelsesarkivet `dist\CoreDeck-Portable.zip`.

## Tester

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Nettlesertester: kjør `python -m http.server 8765` i repositoriets mappe og åpne `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoger og varsler, "PASS n/n" i tittelen) eller `test/web/launcher-smoke.html` (launcheren med et stubbet API).

CI (GitHub Actions) kjører kontrollen av JS-syntaks, Python-testene, `go vet`, `go test -race`, `govulncheck` og bygget.

## Prosjektstruktur

```text
cmd/launcher/         program og innebygd grensesnitt
cmd/disk-helper/      hjelper for oppslag av stasjoner
internal/             programlogikk
scripts/              bygging og pakking
test/                 tester
```

De tomme standarddataene som er innebygd i exe-filen, ligger i `cmd/launcher/web/data/`. Data som opprettes når programmet brukes, er lokale og ignoreres av Git.

## Personvern

Programmet jobber lokalt. Repositoriet inneholder ingen brukerkatalog, ingen historikk, ingen private sider, ingen AI-modeller og ikke noe bibliotek med installasjonsprogrammer. CoreDecks eneste egne nettverkstilgang er søk etter og nedlasting av oppdateringer, som bare kontakter GitHub API og prosjektets offisielle GitHub-utgivelser.

## Sikkerhet

- Den lokale grensesnittserveren lytter bare på `127.0.0.1`, svarer bare på `Host`-headeren `localhost:<port>` / `127.0.0.1:<port>` og sender ingen CORS-headere (alt er same-origin).
- Justeringer kommer fra `data\tweaks.json` (import, redigering) og kjøres med administratorrettigheter. Før en justering kjøres første gang, og etter hver endring av kommandoen, viser konfiguratoren det nøyaktige skriptet og ber om bekreftelse. Godkjenninger (SHA-256 av typen og kommandoen) lagres i `data\state\tweak_approvals.json`; programmet kjører kommandoer fra katalogen, ikke hva som helst siden sender.
- Oppdateringer installeres først når størrelsen og SHA-256 stemmer med `SHA256SUMS.txt` (se [Oppdateringer](#oppdateringer)).

## WebView2-profil

Fra versjon 2.3.1 ligger WebView2-profilen (tema, språk, `localStorage`) i `data\webview\` ved siden av exe-filen. Det brukes en lokal, patchet kopi av `webview_go` (`third_party/webview_go`, beskrevet i `PATCHED.md`). Ved første oppstart kopieres en eksisterende profil fra `%APPDATA%\CoreDeck.exe` (uten hurtigbufferen); originalen blir liggende og kan slettes manuelt senere.

## Lisens

CoreDeck er utgitt under MIT-lisensen, se [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
