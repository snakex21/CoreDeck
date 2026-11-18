**Jazyky:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#sestavení)

Přenosná sada nástrojů pro Windows k uspořádání programů, ovladačů, instalátorů, systémových úprav (tweaků), instalačních profilů a místní dokumentace. Verze **2.4.0** · změny: [CHANGELOG.md](../CHANGELOG.md).

Repozitář je čistý výchozí bod. Neobsahuje žádné programy, instalátory, soukromé odkazy ani data autora. Při prvním spuštění aplikace vytvoří prázdné soubory ve složce `data`; obsah se přidává na obrazovce **Správa**.

## Obsah

- [Funkce](#funkce)
- [Rychlý start](#rychlý-start)
- [Přenosné uspořádání](#přenosné-uspořádání)
- [Složka s daty](#složka-s-daty)
- [Aktualizace](#aktualizace)
- [Jazyk rozhraní](#jazyk-rozhraní)
- [Sestavení](#sestavení)
- [Testy](#testy)
- [Struktura projektu](#struktura-projektu)
- [Soukromí](#soukromí)
- [Zabezpečení](#zabezpečení)
- [Profil WebView2](#profil-webview2)
- [Licence](#licence)

## Funkce

- **Katalog** — vlastní katalog programů a instalátorů, varianty programů a instalační profily, samostatné stránky pro ovladače a nastavení programů, vlastní HTML stránky.
- **Spouštění** — EXE, MSI, BAT, CMD a PowerShell. Přenosné programy a nástroje mají tlačítko **▶ Spustit** a spouštějí se přímo z flash disku (jako dvojklik v Průzkumníku, včetně UAC): soubor nastavený v katalogu, jinak jediný odpovídající `.exe` ve složce (odinstalátory, instalátory a pomocné programy se přeskakují, přednost mají 64bitové). Při více kandidátech se CoreDeck zeptá, který spustit; pokud je k dispozici jen archiv nebo instalátor, otevře složku.
- **Výsledky instalátorů** — návratové kódy 1602 (zrušeno v instalátoru) a 1223 (odmítnuto v UAC) zobrazí krátké „Instalace byla zrušena“; běžné kódy MSI (1603, 1618, 1619, 1638, 1639 a další) mají srozumitelné zprávy; 3010/1641 znamenají úspěch s poznámkou o restartu.
- **Dialogy v aplikaci** — žádná okna prohlížeče `alert/confirm/prompt`; dialogy a oznámení fungují z klávesnice (Enter, Esc, Tab, obnovení fokusu). Spuštění instalátoru, skriptu nebo programu nevyžaduje potvrzení; potvrzení zůstává tam, kde se něco maže nebo mění systém (mazání, tweaky, instalace mnoha programů najednou).
- **KnowledgeBase** s kontrolou odkazů a volitelný vlastní web KnowledgeBase obsluhovaný z `data\KnowledgeBase\`.
- **Export a import** dat.
- **Konfigurátor Windows** — tweaky a debloat s náhledem a schválením skriptu (viz [Zabezpečení](#zabezpečení)).
- **Auditní protokol** administrativních operací v `data/logs/audit.jsonl`.
- **Aktualizace** z vydání na GitHubu se stabilním a testovacím kanálem, ověřením kontrolních součtů a automatickým návratem (viz [Aktualizace](#aktualizace)).
- **Rozhraní** — pamatuje si velikost a režim okna, tmavý a světlý motiv, 30 jazyků rozhraní.

## Rychlý start

Stáhněte `CoreDeck-Portable.zip` z **Releases** a rozbalte ho kamkoli. Program je přenosný a nevyžaduje instalaci: spusťte `CoreDeck.exe`. Při prvním spuštění se vedle exe vytvoří složka `data\` (viz [Složka s daty](#složka-s-daty)).

Chcete-li aktualizovat ručně místo vestavěného aktualizátoru: zavřete CoreDeck a nahraďte `CoreDeck.exe` (a `data\tools\disk-helper.exe`). Složka `data\` zůstane beze změny.

## Přenosné uspořádání

Zamýšlené uspořádání na flash disku je `<jednotka>\_Launcher\CoreDeck.exe`, s knihovnou programů ve složkách vedle `_Launcher`. Cesty v katalogu se ukládají relativně k exe (`../…`) a výchozí cesty pro import/export, programy a ovladače se odvozují od jednotky, ze které CoreDeck běží, takže projekt nepotřebuje pevnou cestu a lze ho přesunout na jinou jednotku nebo jiný počítač.

Pokud flash disk dostane jiné písmeno, cesta uložená se starým písmenem (např. `J:\…` poté, co se jednotka stala `L:`) se vyhledá na jednotce, ze které CoreDeck běží. Při spuštění se takové cesty v `programs.json`, `drivers.json` a `settings.json` jednou přepíšou do relativního tvaru a zachová se kopie `*.drive-<datum>.bak`.

## Složka s daty

Při prvním spuštění se vedle exe vytvoří:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (volitelné: vlastní web KnowledgeBase, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (profil WebView2: motiv, jazyk, localStorage)
├── logs/
└── state/
```

**Soukromá data a jeden exe.** Všechna uživatelská data se čtou z `data\` vedle `CoreDeck.exe`. Soubory vložené do exe (`cmd/launcher/web/data/*.json`) jsou jen prázdné šablony, které se použijí, když soubor v `data\` neexistuje. Díky tomu je veřejný `CoreDeck.exe` z Releases a soukromá instalace tentýž soubor: soukromý katalog, KnowledgeBase (`knowledge.json`) a web `data\KnowledgeBase\` jsou pouze v `data\`. Poškozený soubor v `data\` se ohlásí jako chyba a nenahradí se prázdnou šablonou (další uložení data nepřepíše). Testy hlídají, aby se žádná soukromá data nedostala do vložených šablon ani do `cmd/launcher/web/`.

## Aktualizace

Nastavení → **Aktualizace**.

- **Kanály:** **Stabilní** (výchozí) a **Testovací (předběžné verze)**.
- **Kontrola:** volitelná kontrola při spuštění (nejvýše jednou denně) a ruční tlačítko **Zkontrolovat aktualizace**. Pokud existuje novější verze, okno v aplikaci zobrazí její seznam změn s tlačítky **Aktualizovat nyní**, **Později** a **Přeskočit tuto verzi**.
- **Stažení a ověření:** `CoreDeck-Portable.zip` se stáhne z vydání `snakex21/CoreDeck` na GitHubu; jeho velikost a SHA-256 se porovnají s `SHA256SUMS.txt`. Při neshodě se aktualizace přeruší a nic se nezmění.
- **Instalace:** rozbalí se pouze `CoreDeck.exe` a `data\tools\disk-helper.exe`, a to do přípravné složky `.coredeck-update\` vedle exe. Předchozí binární soubory se zálohují do `.coredeck-update\backup\<verze>\`, soubory se vymění přejmenováním starých na `.old` a CoreDeck se restartuje.
- **Kontrola funkčnosti:** pokud nová verze do 60 sekund neohlásí správné spuštění (nebo se ukončí), předchozí verze se automaticky obnoví a znovu spustí. V Nastavení je k dispozici také **Obnovit předchozí verzi**.
- Funguje v přenosném uspořádání na USB (`<jednotka>\_Launcher\CoreDeck.exe`), samotná aktualizace nevyžaduje práva správce a nikdy nemění nic jiného v `data\`.
- Bez telemetrie: program kontaktuje pouze GitHub API a stažení vydání.

## Jazyk rozhraní

Rozhraní je k dispozici ve 30 jazycích: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 a 한국어.

Jazyk zobrazení Windows se rozpozná automaticky. Výběr jazyka je v horní liště a nabízí možnost **Automaticky**. Text, který v překladu chybí, se zobrazí anglicky. Poznámky k vydání a [CHANGELOG](../CHANGELOG.md) jsou v polštině a angličtině.

## Sestavení

Požadavky:

- Windows,
- Go 1.26 nebo novější (`go.mod` připíná toolchain `go1.26.11`, který `go` stáhne automaticky),
- kompilátor C (MinGW gcc, vyžadovaný WebView),
- volitelně `go-winres` (ikona a informace o verzi v exe),
- Node.js,
- Python.

```bat
build.bat
```

Skript zkontroluje syntaxi JavaScriptu, spustí testy Go a vytvoří místní `CoreDeck.exe` a `data`. Sestavení pro vydání jsou očištěná (`-trimpath -ldflags "-s -w"`) a `build.bat` zapisuje verzi, název produktu a copyright do prostředků Windows. Varianta s diagnostickou konzolí:

```bat
build.bat debug
```

Vývojářské nástroje WebView2 lze zapnout bez samostatného sestavení:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` vytvoří archiv vydání `dist\CoreDeck-Portable.zip`.

## Testy

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testy v prohlížeči: spusťte `python -m http.server 8765` ve složce repozitáře a otevřete `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialogy a oznámení, „PASS n/n“ v titulku) nebo `test/web/launcher-smoke.html` (launcher s náhradním API).

CI (GitHub Actions) spouští kontrolu syntaxe JS, testy Pythonu, `go vet`, `go test -race`, `govulncheck` a sestavení.

## Struktura projektu

```text
cmd/launcher/         aplikace a vložené rozhraní
cmd/disk-helper/      pomocník pro vyhledání jednotky
internal/             logika programu
scripts/              sestavení a balení
test/                 testy
```

Výchozí prázdná data vložená do exe jsou v `cmd/launcher/web/data/`. Data vytvořená při používání programu jsou místní a Git je ignoruje.

## Soukromí

Program pracuje lokálně. Repozitář neobsahuje žádný uživatelský katalog, historii, soukromé stránky, modely AI ani knihovnu instalátorů. Jediný síťový přístup samotného CoreDeck je kontrola a stažení aktualizací, při kterých se kontaktuje pouze GitHub API a oficiální vydání projektu na GitHubu.

## Zabezpečení

- Místní server rozhraní naslouchá pouze na `127.0.0.1`, odpovídá pouze na hlavičku `Host` `localhost:<port>` / `127.0.0.1:<port>` a neposílá žádné hlavičky CORS (vše je same-origin).
- Tweaky pocházejí z `data\tweaks.json` (import, úpravy) a běží s právy správce. Před prvním spuštěním tweaku a po každé změně jeho příkazu konfigurátor zobrazí přesný skript a požádá o potvrzení. Schválení (SHA-256 typu a příkazu) se ukládají do `data\state\tweak_approvals.json`; program spouští příkazy z katalogu, ne cokoli, co pošle stránka.
- Aktualizace se instalují až poté, co velikost a SHA-256 odpovídají `SHA256SUMS.txt` (viz [Aktualizace](#aktualizace)).

## Profil WebView2

Od verze 2.3.1 je profil WebView2 (motiv, jazyk, `localStorage`) uložen v `data\webview\` vedle exe. Používá se místní upravená kopie `webview_go` (`third_party/webview_go`, popsaná v `PATCHED.md`). Při prvním spuštění se existující profil z `%APPDATA%\CoreDeck.exe` zkopíruje (bez mezipaměti); originál zůstane na místě a lze ho později ručně smazat.

## Licence

CoreDeck je vydán pod licencí MIT, viz [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
