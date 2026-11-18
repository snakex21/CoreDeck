**Keeled:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#ehitamine)

Kaasaskantav Windowsi tööriistakomplekt rakenduste, draiverite, installerite, süsteemi kohanduste, seadistusprofiilide ja kohaliku dokumentatsiooni korrastamiseks. Versioon **2.4.0** · muudatused: [CHANGELOG.md](../CHANGELOG.md).

Hoidla on puhas lähtepunkt. See ei sisalda programme, installereid, privaatseid linke ega autori andmeid. Esimesel käivitamisel loob rakendus kausta `data` tühjad failid; sisu lisatakse kuval **Haldus**.

## Sisukord

- [Funktsioonid](#funktsioonid)
- [Kiirstart](#kiirstart)
- [Kaasaskantav paigutus](#kaasaskantav-paigutus)
- [Andmekaust](#andmekaust)
- [Värskendused](#värskendused)
- [Liidese keel](#liidese-keel)
- [Ehitamine](#ehitamine)
- [Testid](#testid)
- [Projekti struktuur](#projekti-struktuur)
- [Privaatsus](#privaatsus)
- [Turvalisus](#turvalisus)
- [WebView2 profiil](#webview2-profiil)
- [Litsents](#litsents)

## Funktsioonid

- **Kataloog** — oma programmide ja installerite kataloog, programmide variandid ja seadistusprofiilid, eraldi lehed draiverite ja programmide seadete jaoks, oma HTML-lehed.
- **Käivitamine** — EXE, MSI, BAT, CMD ja PowerShell. Kaasaskantavatel programmidel ja tööriistadel on nupp **▶ Käivita** ning need käivituvad otse mälupulgalt (nagu topeltklõps Exploreris, koos UAC-ga): kataloogis määratud fail, muidu kausta ainus sobiv `.exe` (desinstallerid, installerid ja abiprogrammid jäetakse vahele, eelistatakse 64-bitist). Mitme kandidaadi korral küsib CoreDeck, milline käivitada; kui kaustas on ainult arhiiv või installer, avab see kausta.
- **Installerite tulemused** — väljumiskoodid 1602 (installeris tühistatud) ja 1223 (UAC keelduti) näitavad lühikest teadet „Installimine tühistati”; levinud MSI koodid (1603, 1618, 1619, 1638, 1639 jt) saavad loetavad teated; 3010/1641 on edukad koos taaskäivitamise märkusega.
- **Rakendusesisesed dialoogid** — brauseri `alert/confirm/prompt` aknaid pole; dialooge ja teateid saab kasutada klaviatuuriga (Enter, Esc, Tab, fookuse taastamine). Installeri, skripti või programmi käivitamine kinnitust ei nõua; kinnitus jääb sinna, kus midagi kustutatakse või süsteemi muudetakse (kustutamised, kohandused, paljude programmide korraga installimine).
- **Teadmusbaas** (KnowledgeBase) linkide kontrolliga ning valikuline oma KnowledgeBase'i sait, mida serveeritakse kaustast `data\KnowledgeBase\`.
- Andmete **eksport ja import**.
- **Windowsi konfigureerija** — kohandused ja debloat koos skripti eelvaate ja kinnitamisega (vt [Turvalisus](#turvalisus)).
- Halduslike toimingute **auditilogi** failis `data/logs/audit.jsonl`.
- **Värskendused** GitHubi väljalasetest stabiilse ja testkanaliga, kontrollsummade kontrolli ja automaatse tagasipööramisega (vt [Värskendused](#värskendused)).
- **Liides** — akna suurus ja režiim jäetakse meelde, tume ja hele teema, 30 liidese keelt.

## Kiirstart

Laadi jaotisest **Releases** alla `CoreDeck-Portable.zip` ja paki see lahti ükskõik kuhu. Programm on kaasaskantav ega vaja installimist: käivita `CoreDeck.exe`. Esimesel käivitamisel luuakse exe kõrvale kaust `data\` (vt [Andmekaust](#andmekaust)).

Kui soovid sisseehitatud värskendaja asemel käsitsi värskendada, sulge CoreDeck ja asenda `CoreDeck.exe` (ja `data\tools\disk-helper.exe`). Kaust `data\` jääb puutumata.

## Kaasaskantav paigutus

Mälupulgal on ette nähtud paigutus `<drive>\_Launcher\CoreDeck.exe`, programmikogu asub kaustades `_Launcher` kõrval. Kataloogi teed salvestatakse exe suhtes (`../…`) ning impordi/ekspordi, programmide ja draiverite vaiketeed tuletatakse kettast, millelt CoreDeck töötab, nii et projekt ei vaja fikseeritud teed ja selle saab teisele kettale või arvutisse viia.

Kui mälupulk saab teise tähe, otsitakse vana tähega salvestatud teed (nt `J:\…` pärast seda, kui kettast sai `L:`) kettalt, millelt CoreDeck töötab. Käivitamisel kirjutatakse sellised teed failides `programs.json`, `drivers.json` ja `settings.json` üks kord suhtelisele kujule ümber ning luuakse koopia `*.drive-<date>.bak`.

## Andmekaust

Esimesel käivitamisel luuakse exe kõrvale järgmine:

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

**Privaatsed andmed ja üks exe.** Kõik kasutajaandmed loetakse kaustast `data\`, mis asub `CoreDeck.exe` kõrval. Exe-sse manustatud failid (`cmd/launcher/web/data/*.json`) on vaid tühjad mallid, mida kasutatakse siis, kui faili kaustas `data\` pole. Seetõttu on avalik `CoreDeck.exe` jaotisest Releases ja privaatne paigaldus sama fail: privaatne kataloog, teadmusbaas (`knowledge.json`) ja sait `data\KnowledgeBase\` asuvad ainult kaustas `data\`. Rikutud fail kaustas `data\` kuvatakse veana ja seda ei asendata tühja malliga (järgmine salvestamine andmeid üle ei kirjuta). Testid tagavad, et privaatsed andmed ei satu manustatud mallidesse ega kausta `cmd/launcher/web/`.

## Värskendused

Seaded → **Värskendused**.

- **Kanalid:** **Stabiilne** (vaikimisi) ja **Test (eelversioonid)**.
- **Kontrollimine:** valikuline kontroll käivitamisel (kõige rohkem kord päevas) ja käsitsi nupp **Otsi värskendusi**. Kui leidub uuem versioon, näitab rakendusesisene aken selle muudatuste logi koos nuppudega **Värskenda kohe**, **Hiljem** ja **Jäta see versioon vahele**.
- **Allalaadimine ja kontroll:** `CoreDeck-Portable.zip` laaditakse alla hoidla `snakex21/CoreDeck` GitHubi väljalaskest; selle suurust ja SHA-256 võrreldakse failiga `SHA256SUMS.txt`. Mittevastavuse korral värskendamine katkestatakse ja midagi ei muudeta.
- **Paigaldamine:** lahti pakitakse ainult `CoreDeck.exe` ja `data\tools\disk-helper.exe` ettevalmistuskausta `.coredeck-update\` exe kõrval. Eelmised binaarfailid varundatakse kausta `.coredeck-update\backup\<version>\`, failid vahetatakse vanade ümbernimetamisega `.old`-iks ja CoreDeck taaskäivitub.
- **Tervisekontroll:** kui uus versioon ei teata 60 sekundi jooksul õigest käivitumisest (või sulgub), taastatakse eelmine versioon automaatselt ja käivitatakse uuesti. Seadetes on saadaval ka **Taasta eelmine versioon**.
- Töötab kaasaskantavas USB-paigutuses (`<drive>\_Launcher\CoreDeck.exe`), värskendamine ise ei vaja administraatoriõigusi ega muuda kaustas `data\` midagi muud.
- Telemeetriat pole: ühendust võetakse ainult GitHubi API ja väljalaske allalaadimisega.

## Liidese keel

Liides on saadaval 30 keeles: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 ja 한국어.

Windowsi kuvakeel tuvastatakse automaatselt. Keelevalija asub ülemisel ribal ja sellel on valik **Automaatne**. Tõlkest puuduv tekst kuvatakse inglise keeles. Väljalaskemärkmed ja [CHANGELOG](../CHANGELOG.md) on poola ja inglise keeles.

## Ehitamine

Nõuded:

- Windows,
- Go 1.26 või uuem (`go.mod` kinnitab tööriistaketi `go1.26.11`, mille `go` laadib automaatselt alla),
- C-kompilaator (MinGW gcc, vajalik WebView jaoks),
- valikuliselt `go-winres` (ikoon ja versiooniteave exe-s),
- Node.js,
- Python.

```bat
build.bat
```

See kontrollib JavaScripti süntaksit, käivitab Go testid ning loob kohaliku `CoreDeck.exe` ja `data`. Väljalaskeversioonid on ilma silumisteabeta (`-trimpath -ldflags "-s -w"`) ning `build.bat` kirjutab Windowsi ressurssidesse versiooni, tootenime ja autoriõiguse. Diagnostikakonsooliga variant:

```bat
build.bat debug
```

WebView2 arendajatööriistad saab lubada ilma eraldi ehitamiseta:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` ehitab väljalaskearhiivi `dist\CoreDeck-Portable.zip`.

## Testid

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Brauseritestid: käivita hoidla kaustas `python -m http.server 8765` ja ava `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoogid ja teated, pealkirjas „PASS n/n”) või `test/web/launcher-smoke.html` (käivitaja asendatud API-ga).

CI (GitHub Actions) käivitab JS-i süntaksikontrolli, Pythoni testid, `go vet`, `go test -race`, `govulncheck` ja ehitamise.

## Projekti struktuur

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Exe-sse manustatud vaikimisi tühjad andmed asuvad kaustas `cmd/launcher/web/data/`. Programmi kasutamise ajal loodud andmed on kohalikud ja Git neid ei jälgi.

## Privaatsus

Programm töötab kohapeal. Hoidla ei sisalda kasutaja kataloogi, ajalugu, privaatseid lehti, tehisintellekti mudeleid ega installerite kogu. CoreDecki ainus võrgukasutus on värskenduste kontroll ja allalaadimine, mis pöörduvad ainult GitHubi API ja projekti ametlike GitHubi väljalasete poole.

## Turvalisus

- Kohalik liideseserver kuulab ainult aadressil `127.0.0.1`, vastab ainult päisele `Host` väärtusega `localhost:<port>` / `127.0.0.1:<port>` ega saada CORS-päiseid (kõik on sama päritoluga).
- Kohandused pärinevad failist `data\tweaks.json` (import, muutmine) ja töötavad administraatoriõigustega. Enne kohanduse esimest käivitamist ja pärast iga selle käsu muutmist näitab konfigureerija täpset skripti ja küsib kinnitust. Kinnitused (tüübi ja käsu SHA-256) salvestatakse faili `data\state\tweak_approvals.json`; programm käivitab käske kataloogist, mitte seda, mida leht saadab.
- Värskendused paigaldatakse alles siis, kui suurus ja SHA-256 vastavad failile `SHA256SUMS.txt` (vt [Värskendused](#värskendused)).

## WebView2 profiil

Alates versioonist 2.3.1 asub WebView2 profiil (teema, keel, `localStorage`) exe kõrval kaustas `data\webview\`. Kasutatakse kohalikku, paigatud `webview_go` koopiat (`third_party/webview_go`, kirjeldatud failis `PATCHED.md`). Esimesel käivitamisel kopeeritakse olemasolev profiil asukohast `%APPDATA%\CoreDeck.exe` (ilma vahemäluta); originaal jääb alles ja selle saab hiljem käsitsi kustutada.

## Litsents

CoreDeck on avaldatud MIT-litsentsi all, vt [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
