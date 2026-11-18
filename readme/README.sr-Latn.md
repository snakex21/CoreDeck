**Jezici:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#izgradnja)

Prenosivi skup alata za Windows za organizovanje aplikacija, drajvera, instalera, sistemskih podešavanja, profila za podešavanje i lokalne dokumentacije. Verzija **2.4.0** · izmene: [CHANGELOG.md](../CHANGELOG.md).

Repozitorijum je čista početna osnova. Ne sadrži programe, instalere, privatne linkove niti podatke autora. Pri prvom pokretanju aplikacija pravi prazne datoteke u fascikli `data`; sadržaj se dodaje na ekranu **Upravljanje**.

## Sadržaj

- [Funkcije](#funkcije)
- [Brzi početak](#brzi-početak)
- [Prenosivi raspored](#prenosivi-raspored)
- [Fascikla sa podacima](#fascikla-sa-podacima)
- [Ažuriranja](#ažuriranja)
- [Jezik interfejsa](#jezik-interfejsa)
- [Izgradnja](#izgradnja)
- [Testovi](#testovi)
- [Struktura projekta](#struktura-projekta)
- [Privatnost](#privatnost)
- [Bezbednost](#bezbednost)
- [WebView2 profil](#webview2-profil)
- [Licenca](#licenca)

## Funkcije

- **Katalog** — sopstveni katalog programa i instalera, varijante programa i profili za podešavanje, zasebne stranice za drajvere i podešavanja programa, sopstvene HTML stranice.
- **Pokretanje** — EXE, MSI, BAT, CMD i PowerShell. Prenosivi programi i alati imaju dugme **▶ Pokreni** i pokreću se direktno sa USB fleš diska (kao dvoklik u Istraživaču datoteka, uključujući UAC): datoteka zadata u katalogu, a inače jedini odgovarajući `.exe` u fascikli (deinstaleri, instaleri i pomoćni programi se preskaču, prednost ima 64-bitna verzija). Kada ima više kandidata, CoreDeck pita koji da pokrene; ako je u fascikli samo arhiva ili instaler, otvara fasciklu.
- **Rezultati instalera** — izlazni kodovi 1602 (otkazano u instaleru) i 1223 (UAC odbijen) prikazuju kratku poruku „Instalacija je otkazana”; česti MSI kodovi (1603, 1618, 1619, 1638, 1639 i drugi) dobijaju razumljive poruke; 3010/1641 su uspeh uz napomenu o ponovnom pokretanju.
- **Dijalozi u aplikaciji** — nema prozora pregledača `alert/confirm/prompt`; dijalozi i obaveštenja rade sa tastaturom (Enter, Esc, Tab, vraćanje fokusa). Pokretanje instalera, skripte ili programa ne traži potvrdu; potvrda ostaje tamo gde se nešto briše ili menja sistem (brisanja, podešavanja sistema, instaliranje više programa odjednom).
- **Baza znanja** (KnowledgeBase) sa proverom linkova i opcioni sopstveni KnowledgeBase sajt koji se služi iz `data\KnowledgeBase\`.
- **Izvoz i uvoz** podataka.
- **Windows konfigurator** — podešavanja sistema i debloat, sa pregledom skripte i odobravanjem (vidi [Bezbednost](#bezbednost)).
- **Dnevnik revizije** administrativnih operacija u `data/logs/audit.jsonl`.
- **Ažuriranja** iz GitHub izdanja sa stabilnim i test kanalom, proverom kontrolnih zbirova i automatskim vraćanjem (vidi [Ažuriranja](#ažuriranja)).
- **Interfejs** — veličina i režim prozora se pamte, tamna i svetla tema, 30 jezika interfejsa.

## Brzi početak

Preuzmite `CoreDeck-Portable.zip` iz odeljka **Releases** i raspakujte ga bilo gde. Program je prenosiv i ne zahteva instalaciju: pokrenite `CoreDeck.exe`. Pri prvom pokretanju pored exe datoteke pravi se fascikla `data\` (vidi [Fascikla sa podacima](#fascikla-sa-podacima)).

Za ručno ažuriranje umesto ugrađenog alata za ažuriranje: zatvorite CoreDeck i zamenite `CoreDeck.exe` (i `data\tools\disk-helper.exe`). Fascikla `data\` ostaje netaknuta.

## Prenosivi raspored

Predviđeni raspored na USB fleš disku je `<drive>\_Launcher\CoreDeck.exe`, a biblioteka programa je u fasciklama pored `_Launcher`. Putanje u katalogu čuvaju se relativno u odnosu na exe (`../…`), a podrazumevane putanje za uvoz/izvoz, programe i drajvere izvode se iz diska sa kog CoreDeck radi, pa projektu nije potrebna fiksna putanja i može se premestiti na drugi disk ili računar.

Ako USB fleš disk dobije drugo slovo, putanja sačuvana sa starim slovom (npr. `J:\…` nakon što je disk postao `L:`) traži se na disku sa kog CoreDeck radi. Pri pokretanju se takve putanje u `programs.json`, `drivers.json` i `settings.json` jednom prepisuju u relativni oblik, uz čuvanje kopije `*.drive-<date>.bak`.

## Fascikla sa podacima

Pri prvom pokretanju pored exe datoteke pravi se sledeće:

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

**Privatni podaci i jedan exe.** Svi korisnički podaci čitaju se iz `data\` pored `CoreDeck.exe`. Datoteke ugrađene u exe (`cmd/launcher/web/data/*.json`) samo su prazni šabloni koji se koriste kada datoteka u `data\` ne postoji. Zbog toga su javni `CoreDeck.exe` iz Releases i privatna instalacija ista datoteka: privatni katalog, baza znanja (`knowledge.json`) i sajt `data\KnowledgeBase\` nalaze se samo u `data\`. Oštećena datoteka u `data\` prijavljuje se kao greška i ne zamenjuje se praznim šablonom (sledeće čuvanje neće prepisati podatke). Testovi obezbeđuju da privatni podaci ne dospeju u ugrađene šablone niti u `cmd/launcher/web/`.

## Ažuriranja

Podešavanja → **Ažuriranja**.

- **Kanali:** **Stabilni** (podrazumevano) i **Test (probne verzije)**.
- **Provera:** opciona provera pri pokretanju (najviše jednom dnevno) i ručno dugme **Proveri ažuriranja**. Kada postoji novija verzija, prozor u aplikaciji prikazuje njen dnevnik izmena sa dugmadima **Ažuriraj sada**, **Kasnije** i **Preskoči ovu verziju**.
- **Preuzimanje i provera:** `CoreDeck-Portable.zip` preuzima se iz GitHub izdanja repozitorijuma `snakex21/CoreDeck`; njegova veličina i SHA-256 proveravaju se prema `SHA256SUMS.txt`. Pri neslaganju ažuriranje se prekida i ništa se ne menja.
- **Instalacija:** raspakuju se samo `CoreDeck.exe` i `data\tools\disk-helper.exe`, u pripremnu fasciklu `.coredeck-update\` pored exe datoteke. Prethodne binarne datoteke čuvaju se kao rezervna kopija u `.coredeck-update\backup\<version>\`, datoteke se zamenjuju preimenovanjem starih u `.old`, a CoreDeck se ponovo pokreće.
- **Provera ispravnosti:** ako nova verzija u roku od 60 sekundi ne prijavi ispravno pokretanje (ili se zatvori), prethodna verzija se automatski vraća i ponovo pokreće. U podešavanjima je dostupno i **Vrati prethodnu verziju**.
- Radi u prenosivom USB rasporedu (`<drive>\_Launcher\CoreDeck.exe`), samo ažuriranje ne zahteva administratorska prava i nikada ne menja ništa drugo u `data\`.
- Bez telemetrije: kontaktiraju se samo GitHub API i preuzimanje izdanja.

## Jezik interfejsa

Interfejs je dostupan na 30 jezika: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 i 한국어.

Jezik prikaza sistema Windows otkriva se automatski. Izbor jezika nalazi se na gornjoj traci i ima opciju **Automatski**. Tekst koji nedostaje u prevodu prikazuje se na engleskom. Napomene uz izdanja i [CHANGELOG](../CHANGELOG.md) su na poljskom i engleskom.

## Izgradnja

Zahtevi:

- Windows,
- Go 1.26 ili noviji (`go.mod` fiksira lanac alata `go1.26.11`, koji `go` preuzima automatski),
- C kompajler (MinGW gcc, potreban za WebView),
- opciono `go-winres` (ikona i informacije o verziji u exe datoteci),
- Node.js,
- Python.

```bat
build.bat
```

Ovo proverava sintaksu JavaScript-a, pokreće Go testove i pravi lokalne `CoreDeck.exe` i `data`. Izdanja se grade bez informacija za otklanjanje grešaka (`-trimpath -ldflags "-s -w"`), a `build.bat` upisuje verziju, naziv proizvoda i autorska prava u Windows resurse. Varijanta sa dijagnostičkom konzolom:

```bat
build.bat debug
```

WebView2 alatke za programere mogu se uključiti bez posebne izgradnje:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` pravi arhivu izdanja `dist\CoreDeck-Portable.zip`.

## Testovi

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testovi u pregledaču: u fascikli repozitorijuma pokrenite `python -m http.server 8765` i otvorite `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dijalozi i obaveštenja, „PASS n/n” u naslovu) ili `test/web/launcher-smoke.html` (pokretač sa zamenskim API-jem).

CI (GitHub Actions) pokreće proveru JS sintakse, Python testove, `go vet`, `go test -race`, `govulncheck` i izgradnju.

## Struktura projekta

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Podrazumevani prazni podaci ugrađeni u exe nalaze se u `cmd/launcher/web/data/`. Podaci nastali tokom korišćenja programa su lokalni i Git ih ignoriše.

## Privatnost

Program radi lokalno. Repozitorijum ne sadrži korisnički katalog, istoriju, privatne stranice, AI modele niti biblioteku instalera. Jedini mrežni pristup samog CoreDeck-a su provera i preuzimanje ažuriranja, koji kontaktiraju samo GitHub API i zvanična GitHub izdanja projekta.

## Bezbednost

- Lokalni server interfejsa sluša samo na `127.0.0.1`, odgovara samo na zaglavlje `Host` sa vrednošću `localhost:<port>` / `127.0.0.1:<port>` i ne šalje CORS zaglavlja (sve je istog porekla).
- Podešavanja sistema dolaze iz `data\tweaks.json` (uvoz, uređivanje) i izvršavaju se sa administratorskim pravima. Pre prvog pokretanja podešavanja i posle svake izmene njegove komande konfigurator prikazuje tačnu skriptu i traži potvrdu. Odobrenja (SHA-256 tipa i komande) čuvaju se u `data\state\tweak_approvals.json`; program izvršava komande iz kataloga, a ne ono što stranica pošalje.
- Ažuriranja se instaliraju tek kada se veličina i SHA-256 poklope sa `SHA256SUMS.txt` (vidi [Ažuriranja](#ažuriranja)).

## WebView2 profil

Od verzije 2.3.1 WebView2 profil (tema, jezik, `localStorage`) nalazi se u `data\webview\` pored exe datoteke. Koristi se lokalna, zakrpljena kopija `webview_go` (`third_party/webview_go`, opisano u `PATCHED.md`). Pri prvom pokretanju postojeći profil kopira se iz `%APPDATA%\CoreDeck.exe` (bez keša); original ostaje na mestu i kasnije se može ručno obrisati.

## Licenca

CoreDeck se objavljuje pod MIT licencom, vidi [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
