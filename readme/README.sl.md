**Jeziki:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#gradnja)

Prenosni nabor orodij za Windows za urejanje aplikacij, gonilnikov, namestitvenih programov, sistemskih prilagoditev, namestitvenih profilov in lokalne dokumentacije. Različica **2.4.0** · spremembe: [CHANGELOG.md](../CHANGELOG.md).

Repozitorij je čisto izhodišče. Ne vsebuje programov, namestitvenih programov, zasebnih povezav ali podatkov avtorja. Ob prvem zagonu aplikacija ustvari prazne datoteke v mapi `data`; vsebino dodate na zaslonu **Upravljanje**.

## Vsebina

- [Funkcije](#funkcije)
- [Hitri začetek](#hitri-začetek)
- [Prenosna postavitev](#prenosna-postavitev)
- [Mapa s podatki](#mapa-s-podatki)
- [Posodobitve](#posodobitve)
- [Jezik vmesnika](#jezik-vmesnika)
- [Gradnja](#gradnja)
- [Testi](#testi)
- [Struktura projekta](#struktura-projekta)
- [Zasebnost](#zasebnost)
- [Varnost](#varnost)
- [Profil WebView2](#profil-webview2)
- [Licenca](#licenca)

## Funkcije

- **Katalog** — lasten katalog programov in namestitvenih programov, različice programov in namestitveni profili, ločene strani za gonilnike in nastavitve programov, lastne strani HTML.
- **Zaganjanje** — EXE, MSI, BAT, CMD in PowerShell. Prenosni programi in orodja imajo gumb **▶ Zaženi** in se zaženejo kar s ključka USB (kot dvoklik v Raziskovalcu, vključno z UAC): datoteka, nastavljena v katalogu, sicer edina ustrezna `.exe` v mapi (programi za odstranitev, namestitveni programi in pomožni programi so preskočeni, prednost ima 64-bitna različica). Pri več kandidatih CoreDeck vpraša, katerega naj zažene; če je v mapi le arhiv ali namestitveni program, odpre mapo.
- **Rezultati namestitvenih programov** — izhodni kodi 1602 (preklicano v namestitvenem programu) in 1223 (UAC zavrnjen) prikažeta kratko sporočilo »Namestitev preklicana«; pogoste kode MSI (1603, 1618, 1619, 1638, 1639 in druge) dobijo razumljiva sporočila; 3010/1641 pomenita uspeh z opombo o ponovnem zagonu.
- **Pogovorna okna v aplikaciji** — brez brskalniških oken `alert/confirm/prompt`; pogovorna okna in obvestila delujejo s tipkovnico (Enter, Esc, Tab, obnovitev fokusa). Zagon namestitvenega programa, skripta ali programa ne zahteva potrditve; potrditev ostaja tam, kjer se kaj briše ali spreminja sistem (brisanje, prilagoditve, namestitev več programov hkrati).
- **Baza znanja** (KnowledgeBase) s preverjanjem povezav in neobvezno lastno spletno mesto KnowledgeBase, ki se streže iz `data\KnowledgeBase\`.
- **Izvoz in uvoz** podatkov.
- **Konfigurator Windows** — prilagoditve in debloat s predogledom skripta in odobritvijo (glejte [Varnost](#varnost)).
- **Revizijski dnevnik** skrbniških operacij v `data/logs/audit.jsonl`.
- **Posodobitve** iz izdaj na GitHubu s stabilnim in testnim kanalom, preverjanjem kontrolnih vsot in samodejno povrnitvijo (glejte [Posodobitve](#posodobitve)).
- **Vmesnik** — velikost in način okna se zapomnita, temna in svetla tema, 30 jezikov vmesnika.

## Hitri začetek

Prenesite `CoreDeck-Portable.zip` iz razdelka **Releases** in ga razširite kamor koli. Program je prenosen in ne potrebuje namestitve: zaženite `CoreDeck.exe`. Ob prvem zagonu se poleg exe ustvari mapa `data\` (glejte [Mapa s podatki](#mapa-s-podatki)).

Za ročno posodobitev namesto vgrajenega posodabljalnika zaprite CoreDeck in zamenjajte `CoreDeck.exe` (in `data\tools\disk-helper.exe`). Mapa `data\` ostane nespremenjena.

## Prenosna postavitev

Predvidena postavitev na ključku USB je `<drive>\_Launcher\CoreDeck.exe`, knjižnica programov pa je v mapah poleg `_Launcher`. Poti v katalogu so shranjene relativno glede na exe (`../…`), privzete poti za uvoz/izvoz, programe in gonilnike pa se izpeljejo iz pogona, s katerega teče CoreDeck, zato projekt ne potrebuje fiksne poti in ga lahko premaknete na drug pogon ali računalnik.

Če ključek USB dobi drugo črko, se pot, shranjena s staro črko (npr. `J:\…`, ko je pogon postal `L:`), poišče na pogonu, s katerega teče CoreDeck. Ob zagonu se take poti v `programs.json`, `drivers.json` in `settings.json` enkrat prepišejo v relativno obliko, pri čemer se ohrani kopija `*.drive-<date>.bak`.

## Mapa s podatki

Ob prvem zagonu se poleg exe ustvari naslednje:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (optional: your own KnowledgeBase site, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2 profile: theme, language, localStorage)
├── logs/
└── state/
```

**Zasebni podatki in en exe.** Vsi uporabniški podatki se berejo iz `data\` poleg `CoreDeck.exe`. Datoteke, vgrajene v exe (`cmd/launcher/web/data/*.json`), so le prazne predloge, ki se uporabijo, ko datoteka v `data\` ne obstaja. Zato sta javni `CoreDeck.exe` iz Releases in zasebna namestitev ista datoteka: zasebni katalog, baza znanja (`knowledge.json`) in spletno mesto `data\KnowledgeBase\` obstajajo le v `data\`. Poškodovana datoteka v `data\` se prikaže kot napaka in se ne nadomesti s prazno predlogo (naslednje shranjevanje ne bo prepisalo podatkov). Testi zagotavljajo, da zasebni podatki ne končajo v vgrajenih predlogah ali v `cmd/launcher/web/`.

## Posodobitve

Nastavitve → **Posodobitve**.

- **Kanali:** **Stabilni** (privzeto) in **Testni (predizdaje)**.
- **Preverjanje:** neobvezno preverjanje ob zagonu (največ enkrat na dan) in ročni gumb **Preveri posodobitve**. Ko obstaja novejša različica, okno v aplikaciji prikaže njen dnevnik sprememb z gumbi **Posodobi zdaj**, **Pozneje** in **Preskoči to različico**.
- **Prenos in preverjanje:** `CoreDeck-Portable.zip` se prenese iz izdaje na GitHubu repozitorija `snakex21/CoreDeck`; njegova velikost in SHA-256 se preverita glede na `SHA256SUMS.txt`. Ob neujemanju se posodobitev prekine in nič se ne spremeni.
- **Namestitev:** razširita se le `CoreDeck.exe` in `data\tools\disk-helper.exe`, in sicer v pripravljalno mapo `.coredeck-update\` poleg exe. Prejšnje binarne datoteke se varnostno kopirajo v `.coredeck-update\backup\<version>\`, datoteke se zamenjajo s preimenovanjem starih v `.old`, nato se CoreDeck znova zažene.
- **Preverjanje delovanja:** če nova različica v 60 sekundah ne sporoči pravilnega zagona (ali se zapre), se prejšnja različica samodejno obnovi in znova zažene. V nastavitvah je na voljo tudi **Obnovi prejšnjo različico**.
- Deluje v prenosni postavitvi USB (`<drive>\_Launcher\CoreDeck.exe`), sama posodobitev ne potrebuje skrbniških pravic in nikoli ne spremeni ničesar drugega v `data\`.
- Brez telemetrije: vzpostavi se stik le z API-jem GitHub in prenosom izdaje.

## Jezik vmesnika

Vmesnik je na voljo v 30 jezikih: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 in 한국어.

Jezik prikaza sistema Windows se zazna samodejno. Izbirnik jezika je v zgornji vrstici in ima možnost **Samodejno**. Besedilo, ki v prevodu manjka, se prikaže v angleščini. Opombe ob izdaji in [CHANGELOG](../CHANGELOG.md) so v poljščini in angleščini.

## Gradnja

Zahteve:

- Windows,
- Go 1.26 ali novejši (`go.mod` določa verigo orodij `go1.26.11`, ki jo `go` prenese samodejno),
- prevajalnik C (MinGW gcc, potreben za WebView),
- po želji `go-winres` (ikona in podatki o različici v exe),
- Node.js,
- Python.

```bat
build.bat
```

S tem se preveri sintaksa JavaScripta, zaženejo testi Go ter ustvarita lokalna `CoreDeck.exe` in `data`. Izdajne gradnje so brez razhroščevalnih podatkov (`-trimpath -ldflags "-s -w"`), `build.bat` pa v vire Windows zapiše različico, ime izdelka in avtorske pravice. Različica z diagnostično konzolo:

```bat
build.bat debug
```

Razvijalska orodja WebView2 lahko vklopite brez ločene gradnje:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` zgradi arhiv izdaje `dist\CoreDeck-Portable.zip`.

## Testi

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testi v brskalniku: v mapi repozitorija zaženite `python -m http.server 8765` in odprite `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (pogovorna okna in obvestila, »PASS n/n« v naslovu) ali `test/web/launcher-smoke.html` (zaganjalnik z nadomestnim API-jem).

CI (GitHub Actions) izvede preverjanje sintakse JS, teste Python, `go vet`, `go test -race`, `govulncheck` in gradnjo.

## Struktura projekta

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Privzeti prazni podatki, vgrajeni v exe, so v `cmd/launcher/web/data/`. Podatki, ustvarjeni med uporabo programa, so lokalni in jih Git prezre.

## Zasebnost

Program deluje lokalno. Repozitorij ne vsebuje uporabniškega kataloga, zgodovine, zasebnih strani, modelov UI ali knjižnice namestitvenih programov. Edini omrežni dostop samega CoreDecka sta preverjanje in prenos posodobitev, ki se povezujeta le z API-jem GitHub in uradnimi izdajami projekta na GitHubu.

## Varnost

- Lokalni strežnik vmesnika posluša le na `127.0.0.1`, odgovarja le na glavo `Host` z vrednostjo `localhost:<port>` / `127.0.0.1:<port>` in ne pošilja glav CORS (vse je istega izvora).
- Prilagoditve izvirajo iz `data\tweaks.json` (uvoz, urejanje) in se izvajajo s skrbniškimi pravicami. Pred prvim zagonom prilagoditve in po vsaki spremembi njenega ukaza konfigurator prikaže točen skript in zahteva potrditev. Odobritve (SHA-256 vrste in ukaza) se shranjujejo v `data\state\tweak_approvals.json`; program izvaja ukaze iz kataloga, ne tistega, kar pošlje stran.
- Posodobitve se namestijo šele, ko se velikost in SHA-256 ujemata s `SHA256SUMS.txt` (glejte [Posodobitve](#posodobitve)).

## Profil WebView2

Od različice 2.3.1 je profil WebView2 (tema, jezik, `localStorage`) v `data\webview\` poleg exe. Uporablja se lokalna, popravljena kopija `webview_go` (`third_party/webview_go`, opisano v `PATCHED.md`). Ob prvem zagonu se obstoječi profil kopira iz `%APPDATA%\CoreDeck.exe` (brez predpomnilnika); izvirnik ostane na mestu in ga lahko pozneje ročno izbrišete.

## Licenca

CoreDeck je izdan pod licenco MIT, glejte [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
