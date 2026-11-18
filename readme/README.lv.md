**Valodas:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#būvēšana)

Portatīvs Windows rīku komplekts lietotņu, draiveru, instalētāju, sistēmas pielāgojumu, iestatīšanas profilu un lokālās dokumentācijas kārtošanai. Versija **2.4.0** · izmaiņas: [CHANGELOG.md](../CHANGELOG.md).

Repozitorijs ir tīrs sākumpunkts. Tajā nav programmu, instalētāju, privātu saišu vai autora datu. Pirmajā palaišanā lietotne izveido tukšus failus mapē `data`; saturu pievieno ekrānā **Pārvaldība**.

## Saturs

- [Iespējas](#iespējas)
- [Ātrā sākšana](#ātrā-sākšana)
- [Portatīvais izkārtojums](#portatīvais-izkārtojums)
- [Datu mape](#datu-mape)
- [Atjauninājumi](#atjauninājumi)
- [Saskarnes valoda](#saskarnes-valoda)
- [Būvēšana](#būvēšana)
- [Testi](#testi)
- [Projekta struktūra](#projekta-struktūra)
- [Privātums](#privātums)
- [Drošība](#drošība)
- [WebView2 profils](#webview2-profils)
- [Licence](#licence)

## Iespējas

- **Katalogs** — savs programmu un instalētāju katalogs, programmu varianti un iestatīšanas profili, atsevišķas lapas draiveriem un programmu iestatījumiem, savas HTML lapas.
- **Palaišana** — EXE, MSI, BAT, CMD un PowerShell. Portatīvajām programmām un rīkiem ir poga **▶ Palaist**, un tās startē tieši no zibatmiņas (kā dubultklikšķis failu pārlūkā, ieskaitot UAC): katalogā norādītais fails, citādi vienīgais atbilstošais `.exe` mapē (atinstalētāji, instalētāji un palīgrīki tiek izlaisti, priekšroka 64 bitu versijai). Ja ir vairāki kandidāti, CoreDeck jautā, kuru palaist; ja mapē ir tikai arhīvs vai instalētājs, tiek atvērta mape.
- **Instalētāju rezultāti** — izejas kodi 1602 (atcelts instalētājā) un 1223 (UAC noraidīts) parāda īsu ziņojumu „Instalēšana atcelta”; biežākie MSI kodi (1603, 1618, 1619, 1638, 1639 u. c.) saņem saprotamus ziņojumus; 3010/1641 ir veiksme ar piezīmi par restartēšanu.
- **Lietotnes iekšējie dialogi** — nav pārlūka `alert/confirm/prompt` logu; dialogi un paziņojumi darbojas ar tastatūru (Enter, Esc, Tab, fokusa atjaunošana). Instalētāja, skripta vai programmas palaišanai apstiprinājums nav vajadzīgs; apstiprinājums paliek tur, kur kaut kas tiek dzēsts vai mainīta sistēma (dzēšana, pielāgojumi, daudzu programmu instalēšana vienlaikus).
- **Zināšanu bāze** (KnowledgeBase) ar saišu pārbaudi un neobligāta sava KnowledgeBase vietne, kas tiek pasniegta no `data\KnowledgeBase\`.
- Datu **eksports un imports**.
- **Windows konfigurators** — pielāgojumi un debloat ar skripta priekšskatījumu un apstiprināšanu (sk. [Drošība](#drošība)).
- Administratīvo darbību **audita žurnāls** failā `data/logs/audit.jsonl`.
- **Atjauninājumi** no GitHub laidieniem ar stabilo un testa kanālu, kontrolsummu pārbaudi un automātisku atgriešanu (sk. [Atjauninājumi](#atjauninājumi)).
- **Saskarne** — loga izmērs un režīms tiek atcerēti, tumšā un gaišā tēma, 30 saskarnes valodas.

## Ātrā sākšana

Lejupielādējiet `CoreDeck-Portable.zip` no sadaļas **Releases** un atarhivējiet jebkurā vietā. Programma ir portatīva un nav jāinstalē: palaidiet `CoreDeck.exe`. Pirmajā palaišanā blakus exe tiek izveidota mape `data\` (sk. [Datu mape](#datu-mape)).

Lai atjauninātu manuāli, nevis ar iebūvēto atjauninātāju: aizveriet CoreDeck un aizstājiet `CoreDeck.exe` (un `data\tools\disk-helper.exe`). Mape `data\` paliek nemainīta.

## Portatīvais izkārtojums

Paredzētais izkārtojums zibatmiņā ir `<drive>\_Launcher\CoreDeck.exe`, bet programmu bibliotēka atrodas mapēs blakus `_Launcher`. Kataloga ceļi tiek glabāti relatīvi pret exe (`../…`), un noklusējuma importa/eksporta, programmu un draiveru ceļi tiek atvasināti no diska, no kura darbojas CoreDeck, tāpēc projektam nav vajadzīgs fiksēts ceļš un to var pārvietot uz citu disku vai datoru.

Ja zibatmiņa saņem citu burtu, ar veco burtu saglabātu ceļu (piem., `J:\…` pēc tam, kad disks kļuvis par `L:`) meklē diskā, no kura darbojas CoreDeck. Startējot šādi ceļi failos `programs.json`, `drivers.json` un `settings.json` vienreiz tiek pārrakstīti relatīvā formā, saglabājot kopiju `*.drive-<date>.bak`.

## Datu mape

Pirmajā palaišanā blakus exe tiek izveidots:

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

**Privāti dati un viens exe.** Visi lietotāja dati tiek nolasīti no `data\` blakus `CoreDeck.exe`. Exe iegultie faili (`cmd/launcher/web/data/*.json`) ir tikai tukšas veidnes, ko izmanto, ja fails mapē `data\` neeksistē. Tādējādi publiskais `CoreDeck.exe` no Releases un privāta instalācija ir viens un tas pats fails: privātais katalogs, zināšanu bāze (`knowledge.json`) un vietne `data\KnowledgeBase\` atrodas tikai mapē `data\`. Bojāts fails mapē `data\` tiek parādīts kā kļūda un netiek aizstāts ar tukšo veidni (nākamā saglabāšana datus nepārrakstīs). Testi nodrošina, ka privāti dati nenonāk iegultajās veidnēs vai mapē `cmd/launcher/web/`.

## Atjauninājumi

Iestatījumi → **Atjauninājumi**.

- **Kanāli:** **Stabilais** (noklusējums) un **Testa (pirmsizlaides versijas)**.
- **Pārbaude:** neobligāta pārbaude startējot (ne biežāk kā reizi dienā) un manuāla poga **Pārbaudīt atjauninājumus**. Ja ir jaunāka versija, lietotnes logs parāda tās izmaiņu žurnālu ar pogām **Atjaunināt tagad**, **Vēlāk** un **Izlaist šo versiju**.
- **Lejupielāde un pārbaude:** `CoreDeck-Portable.zip` tiek lejupielādēts no `snakex21/CoreDeck` GitHub laidiena; tā izmērs un SHA-256 tiek salīdzināti ar `SHA256SUMS.txt`. Neatbilstības gadījumā atjaunināšana tiek pārtraukta un nekas netiek mainīts.
- **Instalēšana:** tiek atarhivēti tikai `CoreDeck.exe` un `data\tools\disk-helper.exe` sagatavošanas mapē `.coredeck-update\` blakus exe. Iepriekšējie binārie faili tiek dublēti mapē `.coredeck-update\backup\<version>\`, faili tiek nomainīti, pārdēvējot vecos par `.old`, un CoreDeck tiek restartēts.
- **Veselības pārbaude:** ja jaunā versija 60 sekunžu laikā neziņo par pareizu startu (vai aizveras), iepriekšējā versija tiek automātiski atjaunota un atkal palaista. Iestatījumos pieejama arī poga **Atjaunot iepriekšējo versiju**.
- Darbojas portatīvajā USB izkārtojumā (`<drive>\_Launcher\CoreDeck.exe`), pašai atjaunināšanai nav vajadzīgas administratora tiesības, un nekas cits mapē `data\` netiek mainīts.
- Bez telemetrijas: tiek sazināts tikai ar GitHub API un laidiena lejupielādi.

## Saskarnes valoda

Saskarne ir pieejama 30 valodās: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 un 한국어.

Windows attēlošanas valoda tiek noteikta automātiski. Valodas izvēle atrodas augšējā joslā, un tai ir opcija **Automātiski**. Tulkojumā trūkstošais teksts tiek rādīts angliski. Laidienu piezīmes un [CHANGELOG](../CHANGELOG.md) ir poļu un angļu valodā.

## Būvēšana

Prasības:

- Windows,
- Go 1.26 vai jaunāka (`go.mod` piesaista rīkķēdi `go1.26.11`, ko `go` lejupielādē automātiski),
- C kompilators (MinGW gcc, nepieciešams WebView),
- pēc izvēles `go-winres` (ikona un versijas informācija exe failā),
- Node.js,
- Python.

```bat
build.bat
```

Tas pārbauda JavaScript sintaksi, palaiž Go testus un izveido lokālu `CoreDeck.exe` un `data`. Laidiena būvējumi ir bez atkļūdošanas informācijas (`-trimpath -ldflags "-s -w"`), un `build.bat` ieraksta Windows resursos versiju, produkta nosaukumu un autortiesības. Variants ar diagnostikas konsoli:

```bat
build.bat debug
```

WebView2 izstrādātāja rīkus var ieslēgt bez atsevišķa būvējuma:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` izveido laidiena arhīvu `dist\CoreDeck-Portable.zip`.

## Testi

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Pārlūka testi: repozitorija mapē palaidiet `python -m http.server 8765` un atveriet `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialogi un paziņojumi, virsrakstā „PASS n/n”) vai `test/web/launcher-smoke.html` (palaidējs ar aizstātu API).

CI (GitHub Actions) palaiž JS sintakses pārbaudi, Python testus, `go vet`, `go test -race`, `govulncheck` un būvēšanu.

## Projekta struktūra

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Noklusējuma tukšie dati, kas iegulti exe, atrodas `cmd/launcher/web/data/`. Programmas lietošanas laikā izveidotie dati ir lokāli, un Git tos ignorē.

## Privātums

Programma darbojas lokāli. Repozitorijā nav lietotāja kataloga, vēstures, privātu lapu, MI modeļu vai instalētāju bibliotēkas. Vienīgā paša CoreDeck piekļuve tīklam ir atjauninājumu pārbaude un lejupielāde, kas sazinās tikai ar GitHub API un projekta oficiālajiem GitHub laidieniem.

## Drošība

- Lokālais saskarnes serveris klausās tikai uz `127.0.0.1`, atbild tikai uz galveni `Host` ar vērtību `localhost:<port>` / `127.0.0.1:<port>` un nesūta CORS galvenes (viss ir no tās pašas izcelsmes).
- Pielāgojumi nāk no `data\tweaks.json` (imports, rediģēšana) un darbojas ar administratora tiesībām. Pirms pielāgojuma pirmās palaišanas un pēc katras tā komandas izmaiņas konfigurators parāda precīzu skriptu un prasa apstiprinājumu. Apstiprinājumi (tipa un komandas SHA-256) tiek glabāti `data\state\tweak_approvals.json`; programma izpilda komandas no kataloga, nevis to, ko sūta lapa.
- Atjauninājumi tiek instalēti tikai pēc tam, kad izmērs un SHA-256 atbilst `SHA256SUMS.txt` (sk. [Atjauninājumi](#atjauninājumi)).

## WebView2 profils

Kopš versijas 2.3.1 WebView2 profils (tēma, valoda, `localStorage`) atrodas `data\webview\` blakus exe. Tiek izmantota lokāla, labota `webview_go` kopija (`third_party/webview_go`, aprakstīta `PATCHED.md`). Pirmajā palaišanā esošais profils tiek nokopēts no `%APPDATA%\CoreDeck.exe` (bez kešatmiņas); oriģināls paliek vietā, un to vēlāk var izdzēst manuāli.

## Licence

CoreDeck tiek izplatīts saskaņā ar MIT licenci, sk. [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
