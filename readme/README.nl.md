**Talen:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#bouwen)

Een portable toolkit voor Windows om programma's, stuurprogramma's, installatieprogramma's, systeemaanpassingen, configuratieprofielen en lokale documentatie te ordenen. Versie **2.4.0** · wijzigingen: [CHANGELOG.md](../CHANGELOG.md).

De repository is een schoon startpunt. Hij bevat geen programma's, installatieprogramma's, privélinks of gegevens van de auteur. Bij de eerste start maakt de applicatie lege bestanden aan in de map `data`; inhoud wordt toegevoegd in het scherm **Beheer**.

## Inhoud

- [Functies](#functies)
- [Snel aan de slag](#snel-aan-de-slag)
- [Portable indeling](#portable-indeling)
- [Gegevensmap](#gegevensmap)
- [Updates](#updates)
- [Interfacetaal](#interfacetaal)
- [Bouwen](#bouwen)
- [Tests](#tests)
- [Projectstructuur](#projectstructuur)
- [Privacy](#privacy)
- [Beveiliging](#beveiliging)
- [WebView2-profiel](#webview2-profiel)
- [Licentie](#licentie)

## Functies

- **Catalogus** — je eigen catalogus van programma's en installatieprogramma's, programmavarianten en configuratieprofielen, aparte pagina's voor stuurprogramma's en programma-instellingen, je eigen HTML-pagina's.
- **Starten** — EXE, MSI, BAT, CMD en PowerShell. Portable programma's en tools hebben een knop **▶ Starten** en starten ter plaatse vanaf de USB-stick (zoals een dubbelklik in Verkenner, UAC inbegrepen): het bestand dat in de catalogus is ingesteld, anders de enige passende `.exe` in de map (verwijderprogramma's, installatieprogramma's en hulpprogramma's worden overgeslagen, 64-bit heeft de voorkeur). Bij meerdere kandidaten vraagt CoreDeck welke gestart moet worden; staat er alleen een archief of een installatieprogramma, dan wordt de map geopend.
- **Resultaten van installatieprogramma's** — de afsluitcodes 1602 (geannuleerd in het installatieprogramma) en 1223 (UAC geweigerd) tonen een korte melding "Installatie geannuleerd"; veelvoorkomende MSI-codes (1603, 1618, 1619, 1638, 1639 en meer) krijgen begrijpelijke berichten; 3010/1641 gelden als geslaagd, met een opmerking over opnieuw opstarten.
- **Dialoogvensters in de app** — geen `alert/confirm/prompt`-vensters van de browser; dialoogvensters en meldingen werken met het toetsenbord (Enter, Esc, Tab, herstel van de focus). Een installatieprogramma, script of programma starten vraagt geen bevestiging; bevestiging blijft waar iets wordt verwijderd of het systeem wordt gewijzigd (verwijderen, aanpassingen, veel programma's tegelijk installeren).
- **KnowledgeBase** met linkcontrole, en een optionele eigen KnowledgeBase-site die vanuit `data\KnowledgeBase\` wordt aangeboden.
- **Export en import** van gegevens.
- **Windows-configurator** — aanpassingen en debloat, met voorbeeldweergave van het script en goedkeuring (zie [Beveiliging](#beveiliging)).
- **Auditlogboek** van beheerhandelingen in `data/logs/audit.jsonl`.
- **Updates** vanuit GitHub-releases met een stabiel en een testkanaal, controle van controlesommen en automatisch terugdraaien (zie [Updates](#updates)).
- **Interface** — venstergrootte en -modus worden onthouden, donker en licht thema, 30 interfacetalen.

## Snel aan de slag

Download `CoreDeck-Portable.zip` van **Releases** en pak het uit op een willekeurige plek. Het programma is portable en hoeft niet geïnstalleerd te worden: start `CoreDeck.exe`. Bij de eerste start wordt de map `data\` naast de exe aangemaakt (zie [Gegevensmap](#gegevensmap)).

Handmatig bijwerken in plaats van met de ingebouwde updater: sluit CoreDeck en vervang `CoreDeck.exe` (en `data\tools\disk-helper.exe`). De map `data\` blijft ongewijzigd.

## Portable indeling

De bedoelde indeling op een USB-stick is `<schijf>\_Launcher\CoreDeck.exe`, met de programmabibliotheek in mappen naast `_Launcher`. Paden in de catalogus worden relatief ten opzichte van de exe opgeslagen (`../…`), en de standaardpaden voor import/export, programma's en stuurprogramma's worden afgeleid van de schijf waarvan CoreDeck draait. Het project heeft dus geen vast pad nodig en kan naar een andere schijf of computer worden verplaatst.

Krijgt de USB-stick een andere letter, dan wordt een pad dat met de oude letter is opgeslagen (bijv. `J:\…` nadat de schijf `L:` is geworden) opgezocht op de schijf waarvan CoreDeck draait. Bij het starten worden zulke paden in `programs.json`, `drivers.json` en `settings.json` eenmalig naar relatieve vorm herschreven, waarbij een kopie `*.drive-<datum>.bak` wordt bewaard.

## Gegevensmap

Bij de eerste start wordt het volgende naast de exe aangemaakt:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (optioneel: je eigen KnowledgeBase-site, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-profiel: thema, taal, localStorage)
├── logs/
└── state/
```

**Privégegevens en één exe.** Alle gebruikersgegevens worden gelezen uit `data\` naast `CoreDeck.exe`. De bestanden die in de exe zijn ingebed (`cmd/launcher/web/data/*.json`), zijn alleen lege sjablonen die worden gebruikt als het bestand in `data\` niet bestaat. Daardoor zijn de openbare `CoreDeck.exe` uit Releases en een privé-installatie hetzelfde bestand: de privécatalogus, de KnowledgeBase (`knowledge.json`) en de site `data\KnowledgeBase\` staan alleen in `data\`. Een beschadigd bestand in `data\` wordt als fout gemeld en niet vervangen door het lege sjabloon (de volgende keer opslaan overschrijft de gegevens niet). Tests zorgen ervoor dat er geen privégegevens in de ingebedde sjablonen of in `cmd/launcher/web/` terechtkomen.

## Updates

Instellingen → **Updates**.

- **Kanalen:** Stabiel (standaard) en Test (pre-releases).
- **Controleren:** optionele controle bij het starten (hooguit eenmaal per dag) en een handmatige knop **Controleren op updates**. Als er een nieuwere versie is, toont een venster in de app de wijzigingen, met **Nu bijwerken**, **Later** en **Deze versie overslaan**.
- **Downloaden en controleren:** `CoreDeck-Portable.zip` wordt gedownload van de GitHub-release van `snakex21/CoreDeck`; de grootte en SHA-256 worden vergeleken met `SHA256SUMS.txt`. Bij een afwijking wordt de update afgebroken en wordt er niets gewijzigd.
- **Installatie:** alleen `CoreDeck.exe` en `data\tools\disk-helper.exe` worden uitgepakt, naar de tussenmap `.coredeck-update\` naast de exe. Van de vorige binaire bestanden wordt een back-up gemaakt in `.coredeck-update\backup\<version>\`, de bestanden worden verwisseld door de oude te hernoemen naar `.old`, en CoreDeck start opnieuw.
- **Statuscontrole:** als de nieuwe versie binnen 60 seconden geen correcte start meldt (of afsluit), wordt de vorige versie automatisch hersteld en opnieuw gestart. **Vorige versie herstellen** is ook beschikbaar in Instellingen.
- Werkt in de portable USB-indeling (`<schijf>\_Launcher\CoreDeck.exe`), heeft voor de update zelf geen beheerdersrechten nodig en wijzigt nooit iets anders in `data\`.
- Geen telemetrie: er wordt alleen contact gemaakt met de GitHub API en de download van de release.

## Interfacetaal

De interface is beschikbaar in 30 talen: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 en 한국어.

De weergavetaal van Windows wordt automatisch herkend. De taalkiezer staat in de bovenste balk en heeft een optie **Automatisch**. Tekst die in een vertaling ontbreekt, wordt in het Engels getoond. Releaseopmerkingen en de [CHANGELOG](../CHANGELOG.md) zijn in het Pools en Engels.

## Bouwen

Vereisten:

- Windows,
- Go 1.26 of nieuwer (`go.mod` legt de toolchain `go1.26.11` vast, die `go` automatisch downloadt),
- een C-compiler (MinGW gcc, vereist door WebView),
- optioneel `go-winres` (pictogram en versie-informatie in de exe),
- Node.js,
- Python.

```bat
build.bat
```

Dit controleert de JavaScript-syntaxis, voert de Go-tests uit en maakt een lokale `CoreDeck.exe` en `data`. Release-builds zijn gestript (`-trimpath -ldflags "-s -w"`), en `build.bat` schrijft de versie, productnaam en copyright in de Windows-resources. Variant met een diagnostische console:

```bat
build.bat debug
```

De ontwikkelaarstools van WebView2 kunnen zonder aparte build worden ingeschakeld:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` bouwt het release-archief `dist\CoreDeck-Portable.zip`.

## Tests

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Browsertests: voer `python -m http.server 8765` uit in de map van de repository en open `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoogvensters en meldingen, "PASS n/n" in de titel) of `test/web/launcher-smoke.html` (de launcher met een gestubde API).

CI (GitHub Actions) voert de JS-syntaxiscontrole, de Python-tests, `go vet`, `go test -race`, `govulncheck` en de build uit.

## Projectstructuur

```text
cmd/launcher/         applicatie en ingebedde interface
cmd/disk-helper/      hulpprogramma voor het opzoeken van schijven
internal/             programmalogica
scripts/              bouwen en verpakken
test/                 tests
```

De standaard lege gegevens die in de exe zijn ingebed, staan in `cmd/launcher/web/data/`. Gegevens die tijdens het gebruik van het programma worden aangemaakt, zijn lokaal en worden door Git genegeerd.

## Privacy

Het programma werkt lokaal. De repository bevat geen gebruikerscatalogus, geschiedenis, privépagina's, AI-modellen of bibliotheek met installatieprogramma's. De enige eigen netwerktoegang van CoreDeck is het controleren op en downloaden van updates, waarbij alleen contact wordt gemaakt met de GitHub API en de officiële GitHub-releases van het project.

## Beveiliging

- De lokale interfaceserver luistert alleen op `127.0.0.1`, antwoordt alleen op de `Host`-header `localhost:<port>` / `127.0.0.1:<port>` en verstuurt geen CORS-headers (alles is same-origin).
- Aanpassingen komen uit `data\tweaks.json` (import, bewerken) en worden met beheerdersrechten uitgevoerd. Voordat een aanpassing voor het eerst wordt uitgevoerd, en na elke wijziging van de opdracht, toont de configurator het exacte script en vraagt om bevestiging. Goedkeuringen (SHA-256 van het type en de opdracht) worden opgeslagen in `data\state\tweak_approvals.json`; het programma voert opdrachten uit de catalogus uit, niet wat de pagina toevallig verstuurt.
- Updates worden pas geïnstalleerd als de grootte en SHA-256 overeenkomen met `SHA256SUMS.txt` (zie [Updates](#updates)).

## WebView2-profiel

Sinds versie 2.3.1 staat het WebView2-profiel (thema, taal, `localStorage`) in `data\webview\` naast de exe. Er wordt een lokale, gepatchte kopie van `webview_go` gebruikt (`third_party/webview_go`, beschreven in `PATCHED.md`). Bij de eerste start wordt een bestaand profiel uit `%APPDATA%\CoreDeck.exe` gekopieerd (zonder de cache); het origineel blijft staan en kan later handmatig worden verwijderd.

## Licentie

CoreDeck wordt uitgebracht onder de MIT-licentie, zie [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
