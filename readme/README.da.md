**Sprog:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#byg)

En portabel værktøjskasse til Windows til at organisere programmer, drivere, installationsprogrammer, systemjusteringer, opsætningsprofiler og lokal dokumentation. Version **2.4.0** · ændringer: [CHANGELOG.md](../CHANGELOG.md).

Repositoriet er et rent udgangspunkt. Det indeholder ingen programmer, installationsprogrammer, private links eller forfatterdata. Ved første start opretter programmet tomme filer i mappen `data`; indhold tilføjes på skærmen **Administration**.

## Indhold

- [Funktioner](#funktioner)
- [Hurtig start](#hurtig-start)
- [Portabelt layout](#portabelt-layout)
- [Datamappe](#datamappe)
- [Opdateringer](#opdateringer)
- [Grænsefladesprog](#grænsefladesprog)
- [Byg](#byg)
- [Test](#test)
- [Projektstruktur](#projektstruktur)
- [Privatliv](#privatliv)
- [Sikkerhed](#sikkerhed)
- [WebView2-profil](#webview2-profil)
- [Licens](#licens)

## Funktioner

- **Katalog** — dit eget katalog over programmer og installationsprogrammer, programvarianter og opsætningsprofiler, separate sider til drivere og programindstillinger, dine egne HTML-sider.
- **Kørsel** — EXE, MSI, BAT, CMD og PowerShell. Portable programmer og værktøjer har en **▶ Kør**-knap og startes direkte fra USB-nøglen (som et dobbeltklik i Stifinder, UAC inklusive): den fil, der er angivet i kataloget, ellers den eneste passende `.exe` i mappen (afinstallationsprogrammer, installationsprogrammer og hjælpeprogrammer springes over, 64-bit foretrækkes). Er der flere kandidater, spørger CoreDeck, hvilken der skal køres; er der kun et arkiv eller et installationsprogram, åbnes mappen.
- **Resultater fra installationsprogrammer** — afslutningskoderne 1602 (annulleret i installationsprogrammet) og 1223 (UAC afvist) viser en kort besked "Installationen blev annulleret"; almindelige MSI-koder (1603, 1618, 1619, 1638, 1639 m.fl.) får letlæselige meddelelser; 3010/1641 er succes med en note om genstart.
- **Dialoger i programmet** — ingen `alert/confirm/prompt`-bokse fra browseren; dialoger og toast-beskeder kan betjenes med tastaturet (Enter, Esc, Tab, gendannelse af fokus). Kørsel af et installationsprogram, et script eller et program kræver ingen bekræftelse; bekræftelse bevares, hvor noget slettes eller systemet ændres (sletninger, justeringer, installation af mange programmer på én gang).
- **KnowledgeBase** med kontrol af links og et valgfrit eget KnowledgeBase-websted, der serveres fra `data\KnowledgeBase\`.
- **Eksport og import** af data.
- **Windows-konfigurator** — justeringer og debloat med forhåndsvisning af script og godkendelse (se [Sikkerhed](#sikkerhed)).
- **Revisionslog** over administrative handlinger i `data/logs/audit.jsonl`.
- **Opdateringer** fra GitHub-udgivelser med stabil kanal og testkanal, verificering af kontrolsummer og automatisk tilbagerulning (se [Opdateringer](#opdateringer)).
- **Grænseflade** — vinduets størrelse og tilstand huskes, mørkt og lyst tema, 30 grænsefladesprog.

## Hurtig start

Download `CoreDeck-Portable.zip` fra **Releases**, og pak den ud hvor som helst. Programmet er portabelt og kræver ingen installation: kør `CoreDeck.exe`. Ved første start oprettes mappen `data\` ved siden af exe-filen (se [Datamappe](#datamappe)).

Sådan opdaterer du manuelt i stedet for med den indbyggede opdatering: luk CoreDeck, og erstat `CoreDeck.exe` (og `data\tools\disk-helper.exe`). Mappen `data\` forbliver uændret.

## Portabelt layout

Det tilsigtede layout på en USB-nøgle er `<drev>\_Launcher\CoreDeck.exe` med programbiblioteket i mapper ved siden af `_Launcher`. Stier i kataloget gemmes relativt til exe-filen (`../…`), og standardstierne til import/eksport, programmer og drivere udledes af det drev, CoreDeck kører fra, så projektet ikke kræver en fast sti og kan flyttes til et andet drev eller en anden computer.

Hvis USB-nøglen får et andet bogstav, slås en sti gemt med det gamle bogstav (f.eks. `J:\…` efter at drevet er blevet til `L:`) op på det drev, CoreDeck kører fra. Ved start omskrives sådanne stier i `programs.json`, `drivers.json` og `settings.json` én gang til relativ form, og der gemmes en kopi `*.drive-<dato>.bak`.

## Datamappe

Ved første start oprettes følgende ved siden af exe-filen:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (valgfrit: dit eget KnowledgeBase-websted, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-profil: tema, sprog, localStorage)
├── logs/
└── state/
```

**Private data og én exe-fil.** Alle brugerdata læses fra `data\` ved siden af `CoreDeck.exe`. Filerne, der er indlejret i exe-filen (`cmd/launcher/web/data/*.json`), er kun tomme skabeloner, som bruges, når filen i `data\` ikke findes. Derfor er den offentlige `CoreDeck.exe` fra Releases og en privat installation den samme fil: det private katalog, KnowledgeBase (`knowledge.json`) og webstedet `data\KnowledgeBase\` findes kun i `data\`. En beskadiget fil i `data\` rapporteres som en fejl og erstattes ikke af den tomme skabelon (den næste lagring overskriver ikke dataene). Test sikrer, at ingen private data havner i de indlejrede skabeloner eller i `cmd/launcher/web/`.

## Opdateringer

Indstillinger → **Opdateringer**.

- **Kanaler:** Stabil (standard) og Test (prøveversioner).
- **Søgning:** valgfri søgning ved opstart (højst én gang om dagen) og en manuel knap **Søg efter opdateringer**. Når der findes en nyere version, viser et vindue i programmet dens ændringslog med **Opdater nu**, **Senere** og **Spring denne version over**.
- **Download og verificering:** `CoreDeck-Portable.zip` downloades fra GitHub-udgivelsen af `snakex21/CoreDeck`; dens størrelse og SHA-256 kontrolleres mod `SHA256SUMS.txt`. Ved uoverensstemmelse afbrydes opdateringen, og intet ændres.
- **Installation:** kun `CoreDeck.exe` og `data\tools\disk-helper.exe` pakkes ud, til mellemlagringsmappen `.coredeck-update\` ved siden af exe-filen. De tidligere binærfiler sikkerhedskopieres i `.coredeck-update\backup\<version>\`, filerne udskiftes ved at omdøbe de gamle til `.old`, og CoreDeck genstarter.
- **Sundhedstjek:** hvis den nye version ikke melder en korrekt start inden for 60 sekunder (eller afsluttes), gendannes den tidligere version automatisk og startes igen. **Gendan den forrige version** findes også i Indstillinger.
- Fungerer i det portable USB-layout (`<drev>\_Launcher\CoreDeck.exe`), kræver ingen administratorrettigheder til selve opdateringen og ændrer aldrig andet i `data\`.
- Ingen telemetri: der kontaktes kun GitHub API og download af udgivelsen.

## Grænsefladesprog

Grænsefladen findes på 30 sprog: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 og 한국어.

Windows' visningssprog registreres automatisk. Sprogvælgeren sidder i den øverste bjælke og har en **Automatisk**-indstilling. Tekst, der mangler i en oversættelse, vises på engelsk. Udgivelsesnoter og [CHANGELOG](../CHANGELOG.md) er på polsk og engelsk.

## Byg

Krav:

- Windows,
- Go 1.26 eller nyere (`go.mod` fastlåser værktøjskæden `go1.26.11`, som `go` downloader automatisk),
- en C-compiler (MinGW gcc, kræves af WebView),
- valgfrit `go-winres` (ikon og versionsoplysninger i exe-filen),
- Node.js,
- Python.

```bat
build.bat
```

Dette kontrollerer JavaScript-syntaksen, kører Go-testene og bygger en lokal `CoreDeck.exe` og `data`. Release-builds er strippede (`-trimpath -ldflags "-s -w"`), og `build.bat` skriver version, produktnavn og copyright i Windows-ressourcerne. Variant med diagnostisk konsol:

```bat
build.bat debug
```

WebView2-udviklerværktøjerne kan slås til uden et separat build:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` bygger udgivelsesarkivet `dist\CoreDeck-Portable.zip`.

## Test

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Browsertest: kør `python -m http.server 8765` i repositoriets mappe, og åbn `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoger og toast-beskeder, "PASS n/n" i titlen) eller `test/web/launcher-smoke.html` (launcheren med et stubbet API).

CI (GitHub Actions) kører kontrollen af JS-syntaks, Python-testene, `go vet`, `go test -race`, `govulncheck` og buildet.

## Projektstruktur

```text
cmd/launcher/         program og indlejret grænseflade
cmd/disk-helper/      hjælper til opslag af drev
internal/             programlogik
scripts/              build og pakning
test/                 test
```

De tomme standarddata, der er indlejret i exe-filen, ligger i `cmd/launcher/web/data/`. Data, der oprettes under brug af programmet, er lokale og ignoreres af Git.

## Privatliv

Programmet arbejder lokalt. Repositoriet indeholder intet brugerkatalog, ingen historik, ingen private sider, ingen AI-modeller og intet bibliotek af installationsprogrammer. CoreDecks eneste egen netværksadgang er søgning efter og download af opdateringer, som kun kontakter GitHub API og projektets officielle GitHub-udgivelser.

## Sikkerhed

- Den lokale grænsefladeserver lytter kun på `127.0.0.1`, svarer kun på `Host`-headeren `localhost:<port>` / `127.0.0.1:<port>` og sender ingen CORS-headere (alt er same-origin).
- Justeringer kommer fra `data\tweaks.json` (import, redigering) og kører med administratorrettigheder. Før en justering køres første gang, og efter hver ændring af dens kommando, viser konfiguratoren det præcise script og beder om bekræftelse. Godkendelser (SHA-256 af typen og kommandoen) gemmes i `data\state\tweak_approvals.json`; programmet kører kommandoer fra kataloget, ikke hvad som helst siden sender.
- Opdateringer installeres først, når størrelsen og SHA-256 stemmer med `SHA256SUMS.txt` (se [Opdateringer](#opdateringer)).

## WebView2-profil

Siden version 2.3.1 ligger WebView2-profilen (tema, sprog, `localStorage`) i `data\webview\` ved siden af exe-filen. Der bruges en lokal, patchet kopi af `webview_go` (`third_party/webview_go`, beskrevet i `PATCHED.md`). Ved første start kopieres en eksisterende profil fra `%APPDATA%\CoreDeck.exe` (uden cachen); originalen bliver liggende og kan senere slettes manuelt.

## Licens

CoreDeck udgives under MIT-licensen, se [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
