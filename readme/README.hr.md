**Jezici:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#izgradnja)

Prijenosni skup alata za Windows za organiziranje aplikacija, upravljačkih programa, instalacijskih programa, prilagodbi sustava, profila postavljanja i lokalne dokumentacije. Verzija **2.4.0** · promjene: [CHANGELOG.md](../CHANGELOG.md).

Repozitorij je čista polazna točka. Ne sadrži programe, instalacijske programe, privatne poveznice ni podatke autora. Pri prvom pokretanju aplikacija stvara prazne datoteke u mapi `data`; sadržaj se dodaje na zaslonu **Upravljanje**.

## Sadržaj

- [Značajke](#značajke)
- [Brzi početak](#brzi-početak)
- [Prijenosni raspored](#prijenosni-raspored)
- [Mapa s podacima](#mapa-s-podacima)
- [Ažuriranja](#ažuriranja)
- [Jezik sučelja](#jezik-sučelja)
- [Izgradnja](#izgradnja)
- [Testovi](#testovi)
- [Struktura projekta](#struktura-projekta)
- [Privatnost](#privatnost)
- [Sigurnost](#sigurnost)
- [Profil WebView2](#profil-webview2)
- [Licenca](#licenca)

## Značajke

- **Katalog** — vlastiti katalog programa i instalacijskih programa, varijante programa i profili postavljanja, zasebne stranice za upravljačke programe i postavke programa, vlastite HTML stranice.
- **Pokretanje** — EXE, MSI, BAT, CMD i PowerShell. Prijenosni programi i alati imaju gumb **▶ Pokreni** i pokreću se na mjestu, s USB sticka (kao dvostruki klik u Exploreru, uključujući UAC): datoteka zadana u katalogu, a inače jedini odgovarajući `.exe` u mapi (programi za deinstalaciju, instalacijski programi i pomoćni alati preskaču se, prednost imaju 64-bitni). Ako postoji više kandidata, CoreDeck pita koji pokrenuti; ako postoji samo arhiva ili instalacijski program, otvara mapu.
- **Rezultati instalacije** — izlazni kodovi 1602 (otkazano u instalacijskom programu) i 1223 (UAC odbijen) prikazuju kratku poruku „Instalacija je otkazana”; uobičajeni MSI kodovi (1603, 1618, 1619, 1638, 1639 i drugi) dobivaju razumljive poruke; 3010/1641 smatraju se uspjehom uz napomenu o ponovnom pokretanju.
- **Dijalozi unutar aplikacije** — bez preglednikovih okvira `alert/confirm/prompt`; dijalozi i obavijesti rade s tipkovnicom (Enter, Esc, Tab, vraćanje fokusa). Pokretanje instalacijskog programa, skripte ili programa ne zahtijeva potvrdu; potvrda ostaje tamo gdje se nešto briše ili mijenja sustav (brisanja, prilagodbe, instalacija više programa odjednom).
- **KnowledgeBase** s provjerom poveznica i, po želji, vlastito KnowledgeBase web-mjesto posluženo iz `data\KnowledgeBase\`.
- **Izvoz i uvoz** podataka.
- **Konfigurator sustava Windows** — prilagodbe i debloat, s pregledom i odobravanjem skripte (vidi [Sigurnost](#sigurnost)).
- **Revizijski zapisnik** administrativnih radnji u `data/logs/audit.jsonl`.
- **Ažuriranja** iz izdanja na GitHubu sa stabilnim i testnim kanalom, provjerom kontrolnog zbroja i automatskim vraćanjem (vidi [Ažuriranja](#ažuriranja)).
- **Sučelje** — veličina i način prozora se pamte, tamna i svijetla tema, 30 jezika sučelja.

## Brzi početak

Preuzmite `CoreDeck-Portable.zip` s **Releases** i raspakirajte ga bilo gdje. Program je prijenosni i ne treba instalaciju: pokrenite `CoreDeck.exe`. Pri prvom pokretanju pokraj exe datoteke stvara se mapa `data\` (vidi [Mapa s podacima](#mapa-s-podacima)).

Za ručno ažuriranje umjesto ugrađenog alata za ažuriranje: zatvorite CoreDeck i zamijenite `CoreDeck.exe` (i `data\tools\disk-helper.exe`). Mapa `data\` ostaje nepromijenjena.

## Prijenosni raspored

Predviđeni raspored na USB sticku je `<drive>\_Launcher\CoreDeck.exe`, s bibliotekom programa u mapama pokraj `_Launcher`. Putanje u katalogu spremaju se relativno u odnosu na exe (`../…`), a zadane putanje za uvoz/izvoz, programe i upravljačke programe izvode se iz pogona s kojeg se CoreDeck pokreće, pa projekt ne treba fiksnu putanju i može se premjestiti na drugi pogon ili računalo.

Ako USB stick dobije drugo slovo, putanja spremljena sa starim slovom (npr. `J:\…` nakon što je pogon postao `L:`) traži se na pogonu s kojeg se CoreDeck pokreće. Pri pokretanju se takve putanje u `programs.json`, `drivers.json` i `settings.json` jednom prepisuju u relativni oblik, uz čuvanje kopije `*.drive-<date>.bak`.

## Mapa s podacima

Pri prvom pokretanju pokraj exe datoteke stvara se sljedeće:

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

**Privatni podaci i jedna exe datoteka.** Svi korisnički podaci čitaju se iz `data\` pokraj `CoreDeck.exe`. Datoteke ugrađene u exe (`cmd/launcher/web/data/*.json`) samo su prazni predlošci koji se koriste kada datoteka u `data\` ne postoji. Zato su javni `CoreDeck.exe` iz Releases i privatna instalacija ista datoteka: privatni katalog, KnowledgeBase (`knowledge.json`) i web-mjesto `data\KnowledgeBase\` nalaze se samo u `data\`. Oštećena datoteka u `data\` prijavljuje se kao pogreška i ne zamjenjuje se praznim predloškom (sljedeće spremanje neće prebrisati podatke). Testovi osiguravaju da nikakvi privatni podaci ne završe u ugrađenim predlošcima ni u `cmd/launcher/web/`.

## Ažuriranja

Postavke → **Ažuriranja**.

- **Kanali:** Stabilni (zadano) i Testni (predizdanja).
- **Provjera:** neobavezna provjera pri pokretanju (najviše jednom dnevno) i ručni gumb **Provjeri ažuriranja**. Kada postoji novija verzija, prozor unutar aplikacije prikazuje popis promjena s gumbima **Ažuriraj sada**, **Kasnije** i **Preskoči ovu verziju**.
- **Preuzimanje i provjera:** `CoreDeck-Portable.zip` preuzima se iz GitHub izdanja `snakex21/CoreDeck`; njegova veličina i SHA-256 provjeravaju se prema `SHA256SUMS.txt`. Ako se ne podudaraju, ažuriranje se prekida i ništa se ne mijenja.
- **Instalacija:** raspakiravaju se samo `CoreDeck.exe` i `data\tools\disk-helper.exe`, u međumapu `.coredeck-update\` pokraj exe datoteke. Prethodne binarne datoteke spremaju se kao sigurnosna kopija u `.coredeck-update\backup\<version>\`, datoteke se zamjenjuju preimenovanjem starih u `.old`, a CoreDeck se ponovno pokreće.
- **Provjera ispravnosti:** ako nova verzija ne prijavi ispravno pokretanje unutar 60 sekundi (ili se zatvori), prethodna verzija automatski se vraća i ponovno pokreće. **Vrati prethodnu verziju** dostupno je i u Postavkama.
- Radi u prijenosnom USB rasporedu (`<drive>\_Launcher\CoreDeck.exe`), za samo ažuriranje ne treba administratorska prava i nikada ne mijenja ništa drugo u `data\`.
- Bez telemetrije: kontaktiraju se samo GitHub API i preuzimanje izdanja.

## Jezik sučelja

Sučelje je dostupno na 30 jezika: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 i 한국어.

Jezik prikaza sustava Windows prepoznaje se automatski. Birač jezika nalazi se na gornjoj traci i ima mogućnost **Automatski**. Tekst koji nedostaje u prijevodu prikazuje se na engleskom. Bilješke o izdanju i [CHANGELOG](../CHANGELOG.md) su na poljskom i engleskom.

## Izgradnja

Zahtjevi:

- Windows,
- Go 1.26 ili noviji (`go.mod` fiksira alatni lanac `go1.26.11`, koji `go` preuzima automatski),
- C prevoditelj (MinGW gcc, potreban za WebView),
- po želji `go-winres` (ikona i podaci o verziji u exe datoteci),
- Node.js,
- Python.

```bat
build.bat
```

Ovo provjerava sintaksu JavaScripta, pokreće Go testove i lokalno stvara `CoreDeck.exe` i `data`. Izdanja se grade bez simbola (`-trimpath -ldflags "-s -w"`), a `build.bat` upisuje verziju, naziv proizvoda i autorska prava u resurse sustava Windows. Varijanta s dijagnostičkom konzolom:

```bat
build.bat debug
```

Alati za razvojne programere WebView2 mogu se uključiti bez zasebne izgradnje:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` stvara arhivu izdanja `dist\CoreDeck-Portable.zip`.

## Testovi

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testovi u pregledniku: pokrenite `python -m http.server 8765` u mapi repozitorija i otvorite `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dijalozi i obavijesti, „PASS n/n” u naslovu) ili `test/web/launcher-smoke.html` (pokretač sa simuliranim API-jem).

CI (GitHub Actions) pokreće provjeru sintakse JS-a, Python testove, `go vet`, `go test -race`, `govulncheck` i izgradnju.

## Struktura projekta

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Zadani prazni podaci ugrađeni u exe nalaze se u `cmd/launcher/web/data/`. Podaci stvoreni tijekom korištenja programa lokalni su i Git ih zanemaruje.

## Privatnost

Program radi lokalno. Repozitorij ne sadrži korisnički katalog, povijest, privatne stranice, AI modele ni biblioteku instalacijskih programa. Jedini mrežni pristup samog CoreDecka jest provjera i preuzimanje ažuriranja, koji se povezuju samo s GitHub API-jem i službenim GitHub izdanjima projekta.

## Sigurnost

- Lokalni poslužitelj sučelja sluša samo na `127.0.0.1`, odgovara samo na zaglavlje `Host` `localhost:<port>` / `127.0.0.1:<port>` i ne šalje CORS zaglavlja (sve je same-origin).
- Prilagodbe dolaze iz `data\tweaks.json` (uvoz, uređivanje) i izvršavaju se s administratorskim pravima. Prije prvog pokretanja prilagodbe i nakon svake promjene njezine naredbe konfigurator prikazuje točnu skriptu i traži potvrdu. Odobrenja (SHA-256 vrste i naredbe) spremaju se u `data\state\tweak_approvals.json`; program pokreće naredbe iz kataloga, a ne ono što stranica pošalje.
- Ažuriranja se instaliraju tek kada se veličina i SHA-256 podudaraju sa `SHA256SUMS.txt` (vidi [Ažuriranja](#ažuriranja)).

## Profil WebView2

Od verzije 2.3.1 profil WebView2 (tema, jezik, `localStorage`) nalazi se u `data\webview\` pokraj exe datoteke. Koristi se lokalna, zakrpana kopija `webview_go` (`third_party/webview_go`, opisano u `PATCHED.md`). Pri prvom pokretanju postojeći profil iz `%APPDATA%\CoreDeck.exe` kopira se (bez predmemorije); izvornik ostaje na mjestu i kasnije se može ručno izbrisati.

## Licenca

CoreDeck se objavljuje pod licencom MIT, vidi [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
