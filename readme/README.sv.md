**Språk:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#bygga)

En portabel verktygslåda för Windows för att organisera program, drivrutiner, installationsprogram, systemjusteringar, installationsprofiler och lokal dokumentation. Version **2.4.0** · ändringar: [CHANGELOG.md](../CHANGELOG.md).

Repositoriet är en ren utgångspunkt. Det innehåller inga program, installationsprogram, privata länkar eller författardata. Vid första start skapar programmet tomma filer i mappen `data`; innehåll läggs till på skärmen **Hantering**.

## Innehåll

- [Funktioner](#funktioner)
- [Snabbstart](#snabbstart)
- [Portabel struktur](#portabel-struktur)
- [Datamapp](#datamapp)
- [Uppdateringar](#uppdateringar)
- [Gränssnittsspråk](#gränssnittsspråk)
- [Bygga](#bygga)
- [Tester](#tester)
- [Projektstruktur](#projektstruktur)
- [Integritet](#integritet)
- [Säkerhet](#säkerhet)
- [WebView2-profil](#webview2-profil)
- [Licens](#licens)

## Funktioner

- **Katalog** — din egen katalog över program och installationsprogram, programvarianter och installationsprofiler, separata sidor för drivrutiner och programinställningar, dina egna HTML-sidor.
- **Körning** — EXE, MSI, BAT, CMD och PowerShell. Portabla program och verktyg har en **▶ Kör**-knapp och startas direkt från USB-minnet (som ett dubbelklick i Utforskaren, UAC inräknat): filen som angetts i katalogen, annars den enda matchande `.exe` i mappen (avinstallerare, installationsprogram och hjälpprogram hoppas över, 64-bitars föredras). Finns flera kandidater frågar CoreDeck vilken som ska köras; finns bara ett arkiv eller ett installationsprogram öppnas mappen.
- **Resultat från installationsprogram** — slutkoderna 1602 (avbrutet i installationsprogrammet) och 1223 (UAC nekades) visar ett kort "Installationen avbröts"; vanliga MSI-koder (1603, 1618, 1619, 1638, 1639 med flera) får begripliga meddelanden; 3010/1641 räknas som lyckade med en notering om omstart.
- **Dialoger i appen** — inga `alert/confirm/prompt`-rutor från webbläsaren; dialoger och aviseringar fungerar med tangentbordet (Enter, Esc, Tab, återställt fokus). Att köra ett installationsprogram, ett skript eller ett program kräver ingen bekräftelse; bekräftelse finns kvar där något tas bort eller systemet ändras (borttagningar, justeringar, installation av många program samtidigt).
- **KnowledgeBase** med länkkontroll och en valfri egen KnowledgeBase-webbplats som serveras från `data\KnowledgeBase\`.
- **Export och import** av data.
- **Windows-konfigurator** — justeringar och debloat, med förhandsgranskning av skript och godkännande (se [Säkerhet](#säkerhet)).
- **Granskningslogg** över administrativa åtgärder i `data/logs/audit.jsonl`.
- **Uppdateringar** från GitHub-utgåvor med stabil kanal och testkanal, verifiering av kontrollsummor och automatisk återställning (se [Uppdateringar](#uppdateringar)).
- **Gränssnitt** — fönstrets storlek och läge sparas, mörkt och ljust tema, 30 gränssnittsspråk.

## Snabbstart

Ladda ned `CoreDeck-Portable.zip` från **Releases** och packa upp den var som helst. Programmet är portabelt och kräver ingen installation: kör `CoreDeck.exe`. Vid första start skapas mappen `data\` bredvid exe-filen (se [Datamapp](#datamapp)).

Så uppdaterar du manuellt i stället för med den inbyggda uppdateraren: stäng CoreDeck och ersätt `CoreDeck.exe` (och `data\tools\disk-helper.exe`). Mappen `data\` lämnas orörd.

## Portabel struktur

Den avsedda strukturen på ett USB-minne är `<enhet>\_Launcher\CoreDeck.exe`, med programbiblioteket i mappar bredvid `_Launcher`. Sökvägar i katalogen sparas relativt till exe-filen (`../…`), och standardsökvägarna för import/export, program och drivrutiner härleds från den enhet CoreDeck körs från, så projektet kräver ingen fast sökväg och kan flyttas till en annan enhet eller dator.

Om USB-minnet får en annan bokstav slås en sökväg som sparats med den gamla bokstaven (t.ex. `J:\…` efter att enheten blivit `L:`) upp på den enhet CoreDeck körs från. Vid start skrivs sådana sökvägar i `programs.json`, `drivers.json` och `settings.json` om en gång till relativ form, och en kopia `*.drive-<datum>.bak` sparas.

## Datamapp

Vid första start skapas följande bredvid exe-filen:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (valfritt: din egen KnowledgeBase-webbplats, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-profil: tema, språk, localStorage)
├── logs/
└── state/
```

**Privata data och en enda exe-fil.** Alla användardata läses från `data\` bredvid `CoreDeck.exe`. Filerna som är inbäddade i exe-filen (`cmd/launcher/web/data/*.json`) är bara tomma mallar som används när filen i `data\` inte finns. Därför är den offentliga `CoreDeck.exe` från Releases och en privat installation samma fil: den privata katalogen, KnowledgeBase (`knowledge.json`) och webbplatsen `data\KnowledgeBase\` finns bara i `data\`. En skadad fil i `data\` rapporteras som ett fel och ersätts inte av den tomma mallen (nästa sparning skriver inte över data). Tester säkerställer att inga privata data hamnar i de inbäddade mallarna eller i `cmd/launcher/web/`.

## Uppdateringar

Inställningar → **Uppdateringar**.

- **Kanaler:** Stabil (standard) och Test (förhandsversioner).
- **Sökning:** valfri sökning vid start (högst en gång per dag) och en manuell knapp **Sök efter uppdateringar**. När en nyare version finns visar ett fönster i appen dess ändringslogg med **Uppdatera nu**, **Senare** och **Hoppa över den här versionen**.
- **Hämtning och verifiering:** `CoreDeck-Portable.zip` hämtas från GitHub-utgåvan av `snakex21/CoreDeck`; dess storlek och SHA-256 kontrolleras mot `SHA256SUMS.txt`. Vid avvikelse avbryts uppdateringen och inget ändras.
- **Installation:** bara `CoreDeck.exe` och `data\tools\disk-helper.exe` packas upp, till mellanlagringsmappen `.coredeck-update\` bredvid exe-filen. De tidigare binärfilerna säkerhetskopieras i `.coredeck-update\backup\<version>\`, filerna byts genom att de gamla döps om till `.old`, och CoreDeck startas om.
- **Hälsokontroll:** om den nya versionen inte rapporterar en korrekt start inom 60 sekunder (eller avslutas) återställs den tidigare versionen automatiskt och startas igen. **Återställ den tidigare versionen** finns även i Inställningar.
- Fungerar i den portabla USB-strukturen (`<enhet>\_Launcher\CoreDeck.exe`), kräver inga administratörsbehörigheter för själva uppdateringen och ändrar aldrig något annat i `data\`.
- Ingen telemetri: endast GitHub API och hämtningen av utgåvan kontaktas.

## Gränssnittsspråk

Gränssnittet finns på 30 språk: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 och 한국어.

Windows visningsspråk identifieras automatiskt. Språkväljaren finns i det övre fältet och har alternativet **Automatiskt**. Text som saknas i en översättning visas på engelska. Versionsinformation och [CHANGELOG](../CHANGELOG.md) finns på polska och engelska.

## Bygga

Krav:

- Windows,
- Go 1.26 eller senare (`go.mod` låser verktygskedjan `go1.26.11`, som `go` hämtar automatiskt),
- en C-kompilator (MinGW gcc, krävs av WebView),
- valfritt `go-winres` (ikon och versionsinformation i exe-filen),
- Node.js,
- Python.

```bat
build.bat
```

Detta kontrollerar JavaScript-syntaxen, kör Go-testerna och skapar en lokal `CoreDeck.exe` och `data`. Release-byggen är strippade (`-trimpath -ldflags "-s -w"`), och `build.bat` skriver version, produktnamn och copyright i Windows-resurserna. Variant med diagnostikkonsol:

```bat
build.bat debug
```

Utvecklarverktygen i WebView2 kan aktiveras utan ett separat bygge:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` bygger utgåvearkivet `dist\CoreDeck-Portable.zip`.

## Tester

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Webbläsartester: kör `python -m http.server 8765` i repositoriets mapp och öppna `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoger och aviseringar, "PASS n/n" i titeln) eller `test/web/launcher-smoke.html` (startprogrammet med ett stubbat API).

CI (GitHub Actions) kör kontrollen av JS-syntax, Python-testerna, `go vet`, `go test -race`, `govulncheck` och bygget.

## Projektstruktur

```text
cmd/launcher/         program och inbäddat gränssnitt
cmd/disk-helper/      hjälpprogram för att hitta enheter
internal/             programlogik
scripts/              bygge och paketering
test/                 tester
```

De tomma standarddata som är inbäddade i exe-filen finns i `cmd/launcher/web/data/`. Data som skapas när programmet används är lokala och ignoreras av Git.

## Integritet

Programmet arbetar lokalt. Repositoriet innehåller ingen användarkatalog, ingen historik, inga privata sidor, inga AI-modeller och inget bibliotek med installationsprogram. CoreDecks enda egna nätverksåtkomst är sökning efter och hämtning av uppdateringar, som bara kontaktar GitHub API och projektets officiella GitHub-utgåvor.

## Säkerhet

- Den lokala gränssnittsservern lyssnar bara på `127.0.0.1`, svarar bara på `Host`-huvudet `localhost:<port>` / `127.0.0.1:<port>` och skickar inga CORS-huvuden (allt är same-origin).
- Justeringar kommer från `data\tweaks.json` (import, redigering) och körs med administratörsbehörighet. Innan en justering körs första gången, och efter varje ändring av dess kommando, visar konfiguratorn exakt skript och ber om bekräftelse. Godkännanden (SHA-256 av typen och kommandot) sparas i `data\state\tweak_approvals.json`; programmet kör kommandon från katalogen, inte vad sidan än skickar.
- Uppdateringar installeras först när storleken och SHA-256 stämmer med `SHA256SUMS.txt` (se [Uppdateringar](#uppdateringar)).

## WebView2-profil

Sedan version 2.3.1 ligger WebView2-profilen (tema, språk, `localStorage`) i `data\webview\` bredvid exe-filen. En lokal, patchad kopia av `webview_go` används (`third_party/webview_go`, beskriven i `PATCHED.md`). Vid första start kopieras en befintlig profil från `%APPDATA%\CoreDeck.exe` (utan cache); originalet ligger kvar och kan tas bort manuellt senare.

## Licens

CoreDeck släpps under MIT-licensen, se [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
