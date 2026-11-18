**Kalbos:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#kompiliavimas)

Nešiojamas Windows įrankių rinkinys programoms, tvarkyklėms, diegimo programoms, sistemos patobulinimams, sąrankos profiliams ir vietinei dokumentacijai tvarkyti. Versija **2.4.0** · pakeitimai: [CHANGELOG.md](../CHANGELOG.md).

Saugykla yra švarus atspirties taškas. Joje nėra programų, diegimo programų, privačių nuorodų ar autoriaus duomenų. Pirmą kartą paleidus programa sukuria tuščius failus aplanke `data`; turinys pridedamas ekrane **Tvarkymas**.

## Turinys

- [Funkcijos](#funkcijos)
- [Greita pradžia](#greita-pradžia)
- [Nešiojamoji struktūra](#nešiojamoji-struktūra)
- [Duomenų aplankas](#duomenų-aplankas)
- [Naujinimai](#naujinimai)
- [Sąsajos kalba](#sąsajos-kalba)
- [Kompiliavimas](#kompiliavimas)
- [Testai](#testai)
- [Projekto struktūra](#projekto-struktūra)
- [Privatumas](#privatumas)
- [Sauga](#sauga)
- [WebView2 profilis](#webview2-profilis)
- [Licencija](#licencija)

## Funkcijos

- **Katalogas** — nuosavas programų ir diegimo programų katalogas, programų variantai ir sąrankos profiliai, atskiri tvarkyklių ir programų nustatymų puslapiai, nuosavi HTML puslapiai.
- **Paleidimas** — EXE, MSI, BAT, CMD ir PowerShell. Nešiojamos programos ir įrankiai turi mygtuką **▶ Paleisti** ir paleidžiami tiesiai iš USB atmintinės (kaip dukart spustelėjus failų naršyklėje, įskaitant UAC): kataloge nurodytas failas, o jei jo nėra – vienintelis tinkamas `.exe` aplanke (pašalinimo, diegimo ir pagalbinės programos praleidžiamos, pirmenybė teikiama 64 bitų). Jei tinka keli failai, CoreDeck klausia, kurį paleisti; jei aplanke yra tik archyvas ar diegimo programa, atidaromas aplankas.
- **Diegimo programų rezultatai** — išėjimo kodai 1602 (atšaukta diegimo programoje) ir 1223 (UAC atmesta) rodo trumpą pranešimą „Diegimas atšauktas“; dažni MSI kodai (1603, 1618, 1619, 1638, 1639 ir kt.) gauna suprantamus pranešimus; 3010/1641 laikomi sėkme su pastaba apie paleidimą iš naujo.
- **Programos vidiniai dialogai** — jokių naršyklės `alert/confirm/prompt` langų; dialogai ir pranešimai valdomi klaviatūra (Enter, Esc, Tab, fokuso atkūrimas). Diegimo programos, scenarijaus ar programos paleidimui patvirtinimo nereikia; patvirtinimas lieka ten, kur kas nors trinama ar keičiama sistema (trynimas, patobulinimai, daugelio programų diegimas vienu metu).
- **Žinių bazė** (KnowledgeBase) su nuorodų tikrinimu ir pasirinktinė nuosava KnowledgeBase svetainė, teikiama iš `data\KnowledgeBase\`.
- Duomenų **eksportas ir importas**.
- **Windows konfigūratorius** — patobulinimai ir debloat su scenarijaus peržiūra ir patvirtinimu (žr. [Sauga](#sauga)).
- Administravimo operacijų **audito žurnalas** faile `data/logs/audit.jsonl`.
- **Naujinimai** iš GitHub leidimų su stabiliu ir bandomuoju kanalu, kontrolinių sumų tikrinimu ir automatiniu grąžinimu (žr. [Naujinimai](#naujinimai)).
- **Sąsaja** — lango dydis ir režimas įsimenami, tamsi ir šviesi tema, 30 sąsajos kalbų.

## Greita pradžia

Atsisiųskite `CoreDeck-Portable.zip` iš skilties **Releases** ir išskleiskite bet kur. Programa nešiojama ir nereikalauja diegimo: paleiskite `CoreDeck.exe`. Pirmą kartą paleidus šalia exe sukuriamas aplankas `data\` (žr. [Duomenų aplankas](#duomenų-aplankas)).

Norėdami atnaujinti rankiniu būdu, o ne integruota naujinimo priemone, uždarykite CoreDeck ir pakeiskite `CoreDeck.exe` (ir `data\tools\disk-helper.exe`). Aplankas `data\` paliekamas nepakeistas.

## Nešiojamoji struktūra

Numatyta struktūra USB atmintinėje yra `<drive>\_Launcher\CoreDeck.exe`, o programų biblioteka – aplankuose šalia `_Launcher`. Katalogo keliai saugomi santykinai exe atžvilgiu (`../…`), o numatytieji importo/eksporto, programų ir tvarkyklių keliai nustatomi pagal diską, iš kurio veikia CoreDeck, todėl projektui nereikia fiksuoto kelio ir jį galima perkelti į kitą diską ar kompiuterį.

Jei USB atmintinė gauna kitą raidę, keliai, išsaugoti su senąja raide (pvz., `J:\…`, kai diskas tapo `L:`), ieškomi diske, iš kurio veikia CoreDeck. Paleidžiant tokie keliai failuose `programs.json`, `drivers.json` ir `settings.json` vieną kartą perrašomi į santykinę formą, išsaugant kopiją `*.drive-<date>.bak`.

## Duomenų aplankas

Pirmą kartą paleidus šalia exe sukuriama:

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

**Privatūs duomenys ir vienas exe.** Visi naudotojo duomenys skaitomi iš `data\` šalia `CoreDeck.exe`. Į exe įterpti failai (`cmd/launcher/web/data/*.json`) yra tik tušti šablonai, naudojami, kai `data\` failo nėra. Todėl viešas `CoreDeck.exe` iš Releases ir privatus diegimas yra tas pats failas: privatus katalogas, žinių bazė (`knowledge.json`) ir svetainė `data\KnowledgeBase\` yra tik `data\`. Sugadintas failas `data\` pranešamas kaip klaida ir nepakeičiamas tuščiu šablonu (kitas išsaugojimas duomenų neperrašys). Testai užtikrina, kad privatūs duomenys nepatektų į įterptus šablonus ar į `cmd/launcher/web/`.

## Naujinimai

Nustatymai → **Naujinimai**.

- **Kanalai:** **Stabilus** (numatytasis) ir **Bandomasis (išankstinės versijos)**.
- **Tikrinimas:** pasirinktinis tikrinimas paleidžiant (ne dažniau kaip kartą per dieną) ir rankinis mygtukas **Tikrinti naujinimus**. Kai yra naujesnė versija, programos langas parodo jos pakeitimų žurnalą su mygtukais **Atnaujinti dabar**, **Vėliau** ir **Praleisti šią versiją**.
- **Atsisiuntimas ir tikrinimas:** `CoreDeck-Portable.zip` atsisiunčiamas iš `snakex21/CoreDeck` GitHub leidimo; jo dydis ir SHA-256 tikrinami pagal `SHA256SUMS.txt`. Neatitikus naujinimas nutraukiamas ir niekas nekeičiama.
- **Diegimas:** išskleidžiami tik `CoreDeck.exe` ir `data\tools\disk-helper.exe` į parengiamąjį aplanką `.coredeck-update\` šalia exe. Ankstesni dvejetainiai failai išsaugomi `.coredeck-update\backup\<version>\`, failai sukeičiami senuosius pervadinant į `.old`, ir CoreDeck paleidžiama iš naujo.
- **Būklės patikra:** jei nauja versija per 60 sekundžių nepraneša apie sėkmingą paleidimą (arba užsidaro), ankstesnė versija automatiškai atkuriama ir vėl paleidžiama. Nustatymuose taip pat yra **Atkurti ankstesnę versiją**.
- Veikia nešiojamoje USB struktūroje (`<drive>\_Launcher\CoreDeck.exe`), pačiam naujinimui nereikia administratoriaus teisių ir niekada nekeičiama nieko kita `data\`.
- Jokios telemetrijos: kreipiamasi tik į GitHub API ir leidimo atsisiuntimą.

## Sąsajos kalba

Sąsaja pasiekiama 30 kalbų: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 ir 한국어.

Windows rodymo kalba aptinkama automatiškai. Kalbos pasirinkimas yra viršutinėje juostoje ir turi parinktį **Automatiškai**. Vertime trūkstamas tekstas rodomas angliškai. Leidimo pastabos ir [CHANGELOG](../CHANGELOG.md) yra lenkų ir anglų kalbomis.

## Kompiliavimas

Reikalavimai:

- Windows,
- Go 1.26 arba naujesnė (`go.mod` fiksuoja įrankių grandinę `go1.26.11`, kurią `go` atsisiunčia automatiškai),
- C kompiliatorius (MinGW gcc, reikalingas WebView),
- pasirinktinai `go-winres` (piktograma ir versijos informacija exe faile),
- Node.js,
- Python.

```bat
build.bat
```

Tai patikrina JavaScript sintaksę, paleidžia Go testus ir sukuria vietinius `CoreDeck.exe` ir `data`. Leidimo versijos be derinimo informacijos (`-trimpath -ldflags "-s -w"`), o `build.bat` įrašo versiją, produkto pavadinimą ir autorių teises į Windows išteklius. Variantas su diagnostine konsole:

```bat
build.bat debug
```

WebView2 kūrėjo įrankius galima įjungti be atskiro kompiliavimo:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` sukuria leidimo archyvą `dist\CoreDeck-Portable.zip`.

## Testai

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Naršyklės testai: saugyklos aplanke paleiskite `python -m http.server 8765` ir atidarykite `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialogai ir pranešimai, pavadinime „PASS n/n“) arba `test/web/launcher-smoke.html` (paleidiklis su imituota API).

CI (GitHub Actions) paleidžia JS sintaksės patikrą, Python testus, `go vet`, `go test -race`, `govulncheck` ir kompiliavimą.

## Projekto struktūra

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Numatytieji tušti duomenys, įterpti į exe, yra `cmd/launcher/web/data/`. Naudojant programą sukurti duomenys yra vietiniai ir Git juos ignoruoja.

## Privatumas

Programa veikia vietoje. Saugykloje nėra naudotojo katalogo, istorijos, privačių puslapių, DI modelių ar diegimo programų bibliotekos. Vienintelė paties CoreDeck tinklo prieiga yra naujinimų tikrinimas ir atsisiuntimas, kurie kreipiasi tik į GitHub API ir oficialius projekto GitHub leidimus.

## Sauga

- Vietinis sąsajos serveris klausosi tik `127.0.0.1`, atsako tik į `Host` antraštę `localhost:<port>` / `127.0.0.1:<port>` ir nesiunčia CORS antraščių (viskas yra tos pačios kilmės).
- Patobulinimai imami iš `data\tweaks.json` (importas, redagavimas) ir vykdomi su administratoriaus teisėmis. Prieš pirmą patobulinimo vykdymą ir po kiekvieno jo komandos pakeitimo konfigūratorius parodo tikslų scenarijų ir prašo patvirtinimo. Patvirtinimai (tipo ir komandos SHA-256) saugomi `data\state\tweak_approvals.json`; programa vykdo komandas iš katalogo, o ne tai, ką siunčia puslapis.
- Naujinimai įdiegiami tik tada, kai dydis ir SHA-256 atitinka `SHA256SUMS.txt` (žr. [Naujinimai](#naujinimai)).

## WebView2 profilis

Nuo 2.3.1 versijos WebView2 profilis (tema, kalba, `localStorage`) yra `data\webview\` šalia exe. Naudojama vietinė pataisyta `webview_go` kopija (`third_party/webview_go`, aprašyta `PATCHED.md`). Pirmą kartą paleidus esamas profilis nukopijuojamas iš `%APPDATA%\CoreDeck.exe` (be talpyklos); originalas lieka vietoje ir vėliau gali būti ištrintas rankiniu būdu.

## Licencija

CoreDeck platinama pagal MIT licenciją, žr. [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
