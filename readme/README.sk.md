**Jazyky:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#zostavenie)

Prenosná sada nástrojov pre Windows na usporiadanie programov, ovládačov, inštalátorov, systémových úprav (tweakov), inštalačných profilov a lokálnej dokumentácie. Verzia **2.4.0** · zmeny: [CHANGELOG.md](../CHANGELOG.md).

Repozitár je čistý východiskový bod. Neobsahuje žiadne programy, inštalátory, súkromné odkazy ani údaje autora. Pri prvom spustení aplikácia vytvorí prázdne súbory v priečinku `data`; obsah sa pridáva na obrazovke **Správa**.

## Obsah

- [Funkcie](#funkcie)
- [Rýchly štart](#rýchly-štart)
- [Prenosné usporiadanie](#prenosné-usporiadanie)
- [Priečinok s údajmi](#priečinok-s-údajmi)
- [Aktualizácie](#aktualizácie)
- [Jazyk rozhrania](#jazyk-rozhrania)
- [Zostavenie](#zostavenie)
- [Testy](#testy)
- [Štruktúra projektu](#štruktúra-projektu)
- [Súkromie](#súkromie)
- [Zabezpečenie](#zabezpečenie)
- [Profil WebView2](#profil-webview2)
- [Licencia](#licencia)

## Funkcie

- **Katalóg** — vlastný katalóg programov a inštalátorov, varianty programov a inštalačné profily, samostatné stránky pre ovládače a nastavenia programov, vlastné HTML stránky.
- **Spúšťanie** — EXE, MSI, BAT, CMD a PowerShell. Prenosné programy a nástroje majú tlačidlo **▶ Spustiť** a spúšťajú sa priamo z USB kľúča (ako dvojklik v Prieskumníkovi, vrátane UAC): súbor nastavený v katalógu, inak jediný zodpovedajúci `.exe` v priečinku (odinštalátory, inštalátory a pomocné programy sa preskakujú, prednosť majú 64-bitové). Pri viacerých kandidátoch sa CoreDeck spýta, ktorý spustiť; ak je k dispozícii len archív alebo inštalátor, otvorí priečinok.
- **Výsledky inštalátorov** — návratové kódy 1602 (zrušené v inštalátore) a 1223 (odmietnuté v UAC) zobrazia krátke „Inštalácia bola zrušená“; bežné kódy MSI (1603, 1618, 1619, 1638, 1639 a ďalšie) majú zrozumiteľné správy; 3010/1641 znamenajú úspech s poznámkou o reštarte.
- **Dialógy v aplikácii** — žiadne okná prehliadača `alert/confirm/prompt`; dialógy a upozornenia fungujú z klávesnice (Enter, Esc, Tab, obnovenie fokusu). Spustenie inštalátora, skriptu alebo programu nevyžaduje potvrdenie; potvrdenie zostáva tam, kde sa niečo odstraňuje alebo mení systém (odstraňovanie, tweaky, inštalácia mnohých programov naraz).
- **KnowledgeBase** s kontrolou odkazov a voliteľný vlastný web KnowledgeBase obsluhovaný z `data\KnowledgeBase\`.
- **Export a import** údajov.
- **Konfigurátor Windows** — tweaky a debloat s náhľadom a schválením skriptu (pozri [Zabezpečenie](#zabezpečenie)).
- **Auditný denník** administratívnych operácií v `data/logs/audit.jsonl`.
- **Aktualizácie** z vydaní na GitHube so stabilným a testovacím kanálom, overením kontrolných súčtov a automatickým návratom (pozri [Aktualizácie](#aktualizácie)).
- **Rozhranie** — pamätá si veľkosť a režim okna, tmavý a svetlý motív, 30 jazykov rozhrania.

## Rýchly štart

Stiahnite `CoreDeck-Portable.zip` z **Releases** a rozbaľte ho kamkoľvek. Program je prenosný a nevyžaduje inštaláciu: spustite `CoreDeck.exe`. Pri prvom spustení sa vedľa exe vytvorí priečinok `data\` (pozri [Priečinok s údajmi](#priečinok-s-údajmi)).

Ak chcete aktualizovať ručne namiesto vstavaného aktualizátora: zatvorte CoreDeck a nahraďte `CoreDeck.exe` (a `data\tools\disk-helper.exe`). Priečinok `data\` zostane bez zmeny.

## Prenosné usporiadanie

Zamýšľané usporiadanie na USB kľúči je `<jednotka>\_Launcher\CoreDeck.exe`, s knižnicou programov v priečinkoch vedľa `_Launcher`. Cesty v katalógu sa ukladajú relatívne k exe (`../…`) a predvolené cesty pre import/export, programy a ovládače sa odvodzujú od jednotky, z ktorej CoreDeck beží, takže projekt nepotrebuje pevnú cestu a dá sa presunúť na inú jednotku alebo iný počítač.

Ak USB kľúč dostane iné písmeno, cesta uložená so starým písmenom (napr. `J:\…` po tom, čo sa jednotka stala `L:`) sa vyhľadá na jednotke, z ktorej CoreDeck beží. Pri spustení sa takéto cesty v `programs.json`, `drivers.json` a `settings.json` raz prepíšu do relatívneho tvaru a zachová sa kópia `*.drive-<dátum>.bak`.

## Priečinok s údajmi

Pri prvom spustení sa vedľa exe vytvorí:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (voliteľné: vlastný web KnowledgeBase, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (profil WebView2: motív, jazyk, localStorage)
├── logs/
└── state/
```

**Súkromné údaje a jeden exe.** Všetky používateľské údaje sa čítajú z `data\` vedľa `CoreDeck.exe`. Súbory vložené do exe (`cmd/launcher/web/data/*.json`) sú len prázdne šablóny, ktoré sa použijú, keď súbor v `data\` neexistuje. Vďaka tomu je verejný `CoreDeck.exe` z Releases a súkromná inštalácia ten istý súbor: súkromný katalóg, KnowledgeBase (`knowledge.json`) a web `data\KnowledgeBase\` sú len v `data\`. Poškodený súbor v `data\` sa ohlási ako chyba a nenahradí sa prázdnou šablónou (ďalšie uloženie údaje neprepíše). Testy strážia, aby sa žiadne súkromné údaje nedostali do vložených šablón ani do `cmd/launcher/web/`.

## Aktualizácie

Nastavenia → **Aktualizácie**.

- **Kanály:** **Stabilný** (predvolený) a **Testovací (predbežné verzie)**.
- **Kontrola:** voliteľná kontrola pri štarte (najviac raz denne) a ručné tlačidlo **Skontrolovať aktualizácie**. Ak existuje novšia verzia, okno v aplikácii zobrazí jej zoznam zmien s tlačidlami **Aktualizovať teraz**, **Neskôr** a **Preskočiť túto verziu**.
- **Stiahnutie a overenie:** `CoreDeck-Portable.zip` sa stiahne z vydania `snakex21/CoreDeck` na GitHube; jeho veľkosť a SHA-256 sa porovnajú s `SHA256SUMS.txt`. Pri nezhode sa aktualizácia preruší a nič sa nezmení.
- **Inštalácia:** rozbalia sa len `CoreDeck.exe` a `data\tools\disk-helper.exe`, a to do prípravného priečinka `.coredeck-update\` vedľa exe. Predchádzajúce binárne súbory sa zálohujú do `.coredeck-update\backup\<verzia>\`, súbory sa vymenia premenovaním starých na `.old` a CoreDeck sa reštartuje.
- **Kontrola funkčnosti:** ak nová verzia do 60 sekúnd neohlási správne spustenie (alebo sa ukončí), predchádzajúca verzia sa automaticky obnoví a znova spustí. V Nastaveniach je k dispozícii aj **Obnoviť predchádzajúcu verziu**.
- Funguje v prenosnom usporiadaní na USB (`<jednotka>\_Launcher\CoreDeck.exe`), samotná aktualizácia nevyžaduje práva správcu a nikdy nemení nič iné v `data\`.
- Bez telemetrie: program kontaktuje len GitHub API a stiahnutie vydania.

## Jazyk rozhrania

Rozhranie je k dispozícii v 30 jazykoch: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 a 한국어.

Jazyk zobrazenia Windows sa rozpozná automaticky. Výber jazyka je v hornej lište a ponúka možnosť **Automaticky**. Text, ktorý v preklade chýba, sa zobrazí v angličtine. Poznámky k vydaniam a [CHANGELOG](../CHANGELOG.md) sú v poľštine a angličtine.

## Zostavenie

Požiadavky:

- Windows,
- Go 1.26 alebo novší (`go.mod` pripína toolchain `go1.26.11`, ktorý `go` stiahne automaticky),
- kompilátor C (MinGW gcc, vyžadovaný WebView),
- voliteľne `go-winres` (ikona a informácie o verzii v exe),
- Node.js,
- Python.

```bat
build.bat
```

Skript skontroluje syntax JavaScriptu, spustí testy Go a vytvorí lokálne `CoreDeck.exe` a `data`. Zostavenia na vydanie sú očistené (`-trimpath -ldflags "-s -w"`) a `build.bat` zapisuje verziu, názov produktu a copyright do prostriedkov Windows. Variant s diagnostickou konzolou:

```bat
build.bat debug
```

Vývojárske nástroje WebView2 sa dajú zapnúť bez samostatného zostavenia:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` vytvorí archív vydania `dist\CoreDeck-Portable.zip`.

## Testy

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testy v prehliadači: spustite `python -m http.server 8765` v priečinku repozitára a otvorte `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialógy a upozornenia, „PASS n/n“ v titulku) alebo `test/web/launcher-smoke.html` (launcher s náhradným API).

CI (GitHub Actions) spúšťa kontrolu syntaxe JS, testy Pythonu, `go vet`, `go test -race`, `govulncheck` a zostavenie.

## Štruktúra projektu

```text
cmd/launcher/         aplikácia a vložené rozhranie
cmd/disk-helper/      pomocník na vyhľadanie jednotky
internal/             logika programu
scripts/              zostavenie a balenie
test/                 testy
```

Predvolené prázdne údaje vložené do exe sú v `cmd/launcher/web/data/`. Údaje vytvorené pri používaní programu sú lokálne a Git ich ignoruje.

## Súkromie

Program pracuje lokálne. Repozitár neobsahuje žiadny používateľský katalóg, históriu, súkromné stránky, modely AI ani knižnicu inštalátorov. Jediný sieťový prístup samotného CoreDeck je kontrola a stiahnutie aktualizácií, pri ktorých sa kontaktuje len GitHub API a oficiálne vydania projektu na GitHube.

## Zabezpečenie

- Lokálny server rozhrania počúva len na `127.0.0.1`, odpovedá len na hlavičku `Host` `localhost:<port>` / `127.0.0.1:<port>` a neposiela žiadne hlavičky CORS (všetko je same-origin).
- Tweaky pochádzajú z `data\tweaks.json` (import, úpravy) a bežia s právami správcu. Pred prvým spustením tweaku a po každej zmene jeho príkazu konfigurátor zobrazí presný skript a požiada o potvrdenie. Schválenia (SHA-256 typu a príkazu) sa ukladajú do `data\state\tweak_approvals.json`; program spúšťa príkazy z katalógu, nie čokoľvek, čo pošle stránka.
- Aktualizácie sa inštalujú až potom, čo veľkosť a SHA-256 zodpovedajú `SHA256SUMS.txt` (pozri [Aktualizácie](#aktualizácie)).

## Profil WebView2

Od verzie 2.3.1 je profil WebView2 (motív, jazyk, `localStorage`) uložený v `data\webview\` vedľa exe. Používa sa lokálna upravená kópia `webview_go` (`third_party/webview_go`, opísaná v `PATCHED.md`). Pri prvom spustení sa existujúci profil z `%APPDATA%\CoreDeck.exe` skopíruje (bez vyrovnávacej pamäte); originál zostane na mieste a neskôr ho možno ručne odstrániť.

## Licencia

CoreDeck je vydaný pod licenciou MIT, pozri [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
