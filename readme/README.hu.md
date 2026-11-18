**Nyelvek:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#fordítás)

Hordozható Windows-eszközkészlet alkalmazások, illesztőprogramok, telepítők, rendszerbeállítások, telepítési profilok és helyi dokumentáció rendezéséhez. Verzió: **2.4.0** · változások: [CHANGELOG.md](../CHANGELOG.md).

A tároló tiszta kiindulópont. Nem tartalmaz programokat, telepítőket, privát hivatkozásokat vagy a szerző adatait. Első indításkor az alkalmazás üres fájlokat hoz létre a `data` mappában; a tartalmat a **Kezelés** képernyőn lehet hozzáadni.

## Tartalom

- [Funkciók](#funkciók)
- [Gyors kezdés](#gyors-kezdés)
- [Hordozható elrendezés](#hordozható-elrendezés)
- [Adatmappa](#adatmappa)
- [Frissítések](#frissítések)
- [Felület nyelve](#felület-nyelve)
- [Fordítás](#fordítás)
- [Tesztek](#tesztek)
- [Projektszerkezet](#projektszerkezet)
- [Adatvédelem](#adatvédelem)
- [Biztonság](#biztonság)
- [WebView2-profil](#webview2-profil)
- [Licenc](#licenc)

## Funkciók

- **Katalógus** — saját katalógus programokból és telepítőkből, programváltozatok és telepítési profilok, külön oldalak az illesztőprogramokhoz és a programbeállításokhoz, saját HTML-oldalak.
- **Futtatás** — EXE, MSI, BAT, CMD és PowerShell. A hordozható programoknak és eszközöknek **▶ Futtatás** gombjuk van, és helyben, a pendrive-ról indulnak (mint egy dupla kattintás az Intézőben, UAC-vel együtt): a katalógusban megadott fájl, ennek hiányában a mappában található egyetlen megfelelő `.exe` (az eltávolítók, telepítők és segédprogramok kimaradnak, a 64 bites változat előnyt élvez). Több jelölt esetén a CoreDeck megkérdezi, melyiket futtassa; ha csak egy archívum vagy telepítő van, megnyitja a mappát.
- **Telepítési eredmények** — az 1602 (megszakítva a telepítőben) és az 1223 (UAC elutasítva) kilépési kód rövid „Telepítés megszakítva” üzenetet jelenít meg; a gyakori MSI-kódok (1603, 1618, 1619, 1638, 1639 és mások) érthető üzenetet kapnak; a 3010/1641 sikernek számít, újraindítási megjegyzéssel.
- **Alkalmazáson belüli párbeszédablakok** — nincsenek böngészős `alert/confirm/prompt` ablakok; a párbeszédablakok és értesítések billentyűzettel is kezelhetők (Enter, Esc, Tab, fókusz visszaállítása). Telepítő, szkript vagy program futtatásához nem kell megerősítés; megerősítés ott marad, ahol valami törlődik vagy a rendszer megváltozik (törlések, finomhangolások, sok program egyszerre történő telepítése).
- **KnowledgeBase** hivatkozás-ellenőrzéssel, valamint opcionálisan saját KnowledgeBase-webhely, amelyet a `data\KnowledgeBase\` szolgál ki.
- Adatok **exportálása és importálása**.
- **Windows-konfigurátor** — finomhangolások és debloat, a szkript előnézetével és jóváhagyásával (lásd [Biztonság](#biztonság)).
- Az adminisztratív műveletek **naplója** a `data/logs/audit.jsonl` fájlban.
- **Frissítések** a GitHub-kiadásokból stabil és tesztcsatornával, ellenőrzőösszeg-vizsgálattal és automatikus visszaállítással (lásd [Frissítések](#frissítések)).
- **Felület** — az ablak mérete és módja megmarad, sötét és világos téma, 30 felületi nyelv.

## Gyors kezdés

Töltse le a `CoreDeck-Portable.zip` fájlt a **Releases** oldalról, és csomagolja ki bárhová. A program hordozható, nem kell telepíteni: futtassa a `CoreDeck.exe` fájlt. Első indításkor az exe mellett létrejön a `data\` mappa (lásd [Adatmappa](#adatmappa)).

Ha a beépített frissítő helyett kézzel szeretne frissíteni: zárja be a CoreDecket, és cserélje le a `CoreDeck.exe` (és a `data\tools\disk-helper.exe`) fájlt. A `data\` mappa változatlan marad.

## Hordozható elrendezés

A pendrive-on javasolt elrendezés: `<drive>\_Launcher\CoreDeck.exe`, a programkönyvtár pedig a `_Launcher` melletti mappákban. A katalógus útvonalai az exe-hez képest relatívan tárolódnak (`../…`), az alapértelmezett import/export-, program- és illesztőprogram-útvonalak pedig abból a meghajtóból származnak, amelyről a CoreDeck fut, így a projektnek nincs szüksége rögzített útvonalra, és áthelyezhető másik meghajtóra vagy számítógépre.

Ha a pendrive más betűjelet kap, a régi betűjellel mentett útvonalat (pl. `J:\…`, miután a meghajtó `L:` lett) a program azon a meghajtón keresi, amelyről a CoreDeck fut. Indításkor az ilyen útvonalak a `programs.json`, `drivers.json` és `settings.json` fájlokban egyszer relatív formára íródnak át, és egy `*.drive-<date>.bak` másolat megmarad.

## Adatmappa

Első indításkor az exe mellett a következők jönnek létre:

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

**Privát adatok és egyetlen exe.** Minden felhasználói adat a `CoreDeck.exe` melletti `data\` mappából töltődik be. Az exe-be ágyazott fájlok (`cmd/launcher/web/data/*.json`) csak üres sablonok, amelyeket akkor használ a program, ha a fájl nem létezik a `data\` mappában. Így a Releases oldalon közzétett nyilvános `CoreDeck.exe` és egy privát telepítés ugyanaz a fájl: a privát katalógus, a KnowledgeBase (`knowledge.json`) és a `data\KnowledgeBase\` webhely csak a `data\` mappában található. A `data\` mappában lévő sérült fájlt a program hibaként jelzi, és nem cseréli le az üres sablonra (a következő mentés nem írja felül az adatokat). Tesztek biztosítják, hogy privát adat ne kerüljön a beágyazott sablonokba vagy a `cmd/launcher/web/` mappába.

## Frissítések

Beállítások → **Frissítések**.

- **Csatornák:** Stabil (alapértelmezett) és Teszt (előzetes kiadások).
- **Ellenőrzés:** opcionális ellenőrzés indításkor (legfeljebb naponta egyszer) és kézi **Frissítések keresése** gomb. Ha újabb verzió érhető el, egy alkalmazáson belüli ablak mutatja a változásait a **Frissítés most**, **Később** és **Verzió kihagyása** gombokkal.
- **Letöltés és ellenőrzés:** a `CoreDeck-Portable.zip` a `snakex21/CoreDeck` GitHub-kiadásából töltődik le; a méretét és SHA-256 értékét a program a `SHA256SUMS.txt` alapján ellenőrzi. Eltérés esetén a frissítés megszakad, és semmi sem változik.
- **Telepítés:** csak a `CoreDeck.exe` és a `data\tools\disk-helper.exe` csomagolódik ki, az exe melletti `.coredeck-update\` átmeneti mappába. Az előző binárisokról biztonsági másolat készül a `.coredeck-update\backup\<version>\` mappába, a fájlcsere a régiek `.old` kiterjesztésre való átnevezésével történik, majd a CoreDeck újraindul.
- **Állapotellenőrzés:** ha az új verzió 60 másodpercen belül nem jelez sikeres indulást (vagy kilép), az előző verzió automatikusan visszaáll és újraindul. Az **Előző verzió visszaállítása** a Beállításokban is elérhető.
- Működik a hordozható USB-elrendezésben (`<drive>\_Launcher\CoreDeck.exe`), magához a frissítéshez nem kell rendszergazdai jog, és soha nem módosít semmi mást a `data\` mappában.
- Nincs telemetria: a program csak a GitHub API-hoz és a kiadás letöltéséhez kapcsolódik.

## Felület nyelve

A felület 30 nyelven érhető el: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 és 한국어.

A Windows megjelenítési nyelvét a program automatikusan felismeri. A nyelvválasztó a felső sávban található, és van **Automatikus** lehetősége. A fordításból hiányzó szöveg angolul jelenik meg. A kiadási megjegyzések és a [CHANGELOG](../CHANGELOG.md) lengyel és angol nyelvűek.

## Fordítás

Követelmények:

- Windows,
- Go 1.26 vagy újabb (a `go.mod` a `go1.26.11` eszközláncot rögzíti, amelyet a `go` automatikusan letölt),
- C-fordító (MinGW gcc, a WebView igényli),
- opcionálisan `go-winres` (ikon és verzióinformáció az exe-ben),
- Node.js,
- Python.

```bat
build.bat
```

Ez ellenőrzi a JavaScript szintaxisát, lefuttatja a Go-teszteket, és helyben elkészíti a `CoreDeck.exe` fájlt és a `data` mappát. A kiadási buildek szimbólumok nélkül készülnek (`-trimpath -ldflags "-s -w"`), a `build.bat` pedig beírja a verziót, a terméknevet és a szerzői jogot a Windows-erőforrásokba. Diagnosztikai konzollal rendelkező változat:

```bat
build.bat debug
```

A WebView2 fejlesztői eszközei külön build nélkül is bekapcsolhatók:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

A `scripts\package.bat` elkészíti a `dist\CoreDeck-Portable.zip` kiadási archívumot.

## Tesztek

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Böngészős tesztek: futtassa a `python -m http.server 8765` parancsot a tároló mappájában, és nyissa meg a `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (párbeszédablakok és értesítések, „PASS n/n” a címben) vagy a `test/web/launcher-smoke.html` (az indító helyettesített API-val) oldalt.

A CI (GitHub Actions) lefuttatja a JS-szintaxis ellenőrzését, a Python-teszteket, a `go vet`, `go test -race`, `govulncheck` parancsokat és a buildet.

## Projektszerkezet

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Az exe-be ágyazott alapértelmezett üres adatok a `cmd/launcher/web/data/` mappában vannak. A program használata közben létrejövő adatok helyiek, és a Git figyelmen kívül hagyja őket.

## Adatvédelem

A program helyben működik. A tároló nem tartalmaz felhasználói katalógust, előzményeket, privát oldalakat, MI-modelleket vagy telepítőkönyvtárat. Maga a CoreDeck csak a frissítések ellenőrzéséhez és letöltéséhez használja a hálózatot, és ekkor is csak a GitHub API-hoz és a projekt hivatalos GitHub-kiadásaihoz kapcsolódik.

## Biztonság

- A helyi felületi kiszolgáló csak a `127.0.0.1` címen figyel, csak a `localhost:<port>` / `127.0.0.1:<port>` `Host` fejlécre válaszol, és nem küld CORS-fejléceket (minden same-origin).
- A finomhangolások a `data\tweaks.json` fájlból származnak (importálás, szerkesztés), és rendszergazdai jogokkal futnak. Mielőtt egy finomhangolás először lefut, illetve a parancsa minden módosítása után, a konfigurátor megmutatja a pontos szkriptet, és megerősítést kér. A jóváhagyások (a típus és a parancs SHA-256 értéke) a `data\state\tweak_approvals.json` fájlban tárolódnak; a program a katalógusból futtat parancsokat, nem azt, amit az oldal küld.
- A frissítések csak akkor települnek, ha a méret és az SHA-256 egyezik a `SHA256SUMS.txt` tartalmával (lásd [Frissítések](#frissítések)).

## WebView2-profil

A 2.3.1-es verzió óta a WebView2-profil (téma, nyelv, `localStorage`) az exe melletti `data\webview\` mappában található. A program a `webview_go` helyi, javított másolatát használja (`third_party/webview_go`, leírása a `PATCHED.md` fájlban). Első indításkor a `%APPDATA%\CoreDeck.exe` mappában lévő meglévő profil átmásolódik (gyorsítótár nélkül); az eredeti a helyén marad, és később kézzel törölhető.

## Licenc

A CoreDeck MIT-licenc alatt jelenik meg, lásd [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
