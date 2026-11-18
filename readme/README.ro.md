**Limbi:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#compilare)

Un set de instrumente portabil pentru Windows, pentru organizarea aplicațiilor, driverelor, programelor de instalare, ajustărilor de sistem, profilurilor de configurare și documentației locale. Versiunea **2.4.0** · modificări: [CHANGELOG.md](../CHANGELOG.md).

Depozitul este un punct de plecare curat. Nu conține programe, programe de instalare, linkuri private sau date ale autorului. La prima pornire, aplicația creează fișiere goale în folderul `data`; conținutul se adaugă în ecranul **Gestionare**.

## Cuprins

- [Funcționalități](#funcționalități)
- [Pornire rapidă](#pornire-rapidă)
- [Structura portabilă](#structura-portabilă)
- [Folderul de date](#folderul-de-date)
- [Actualizări](#actualizări)
- [Limba interfeței](#limba-interfeței)
- [Compilare](#compilare)
- [Teste](#teste)
- [Structura proiectului](#structura-proiectului)
- [Confidențialitate](#confidențialitate)
- [Securitate](#securitate)
- [Profilul WebView2](#profilul-webview2)
- [Licență](#licență)

## Funcționalități

- **Catalog** — propriul catalog de programe și programe de instalare, variante de programe și profiluri de configurare, pagini separate pentru drivere și setările programelor, propriile pagini HTML.
- **Rulare** — EXE, MSI, BAT, CMD și PowerShell. Programele și instrumentele portabile au un buton **▶ Rulează** și pornesc pe loc de pe stick-ul USB (ca un dublu clic în Explorer, inclusiv UAC): fișierul setat în catalog, altfel singurul `.exe` potrivit din folder (programele de dezinstalare, de instalare și utilitarele auxiliare sunt omise, se preferă cele pe 64 de biți). Când există mai mulți candidați, CoreDeck întreabă pe care să-l ruleze; când există doar o arhivă sau un program de instalare, deschide folderul.
- **Rezultatele instalării** — codurile de ieșire 1602 (anulat în programul de instalare) și 1223 (UAC refuzat) afișează un scurt „Instalare anulată”; codurile MSI obișnuite (1603, 1618, 1619, 1638, 1639 și altele) primesc mesaje ușor de înțeles; 3010/1641 înseamnă succes, cu o notă despre repornire.
- **Dialoguri în aplicație** — fără casete `alert/confirm/prompt` ale browserului; dialogurile și notificările funcționează cu tastatura (Enter, Esc, Tab, restaurarea focusului). Rularea unui program de instalare, a unui script sau a unui program nu necesită confirmare; confirmarea rămâne acolo unde ceva este șters sau sistemul este modificat (ștergeri, ajustări, instalarea mai multor programe deodată).
- **KnowledgeBase** cu verificarea linkurilor și, opțional, un site KnowledgeBase propriu, servit din `data\KnowledgeBase\`.
- **Export și import** de date.
- **Configurator Windows** — ajustări și debloat, cu previzualizarea și aprobarea scriptului (vezi [Securitate](#securitate)).
- **Jurnal de audit** al operațiunilor administrative în `data/logs/audit.jsonl`.
- **Actualizări** din versiunile publicate pe GitHub, cu canal stabil și de test, verificarea sumei de control și restaurare automată (vezi [Actualizări](#actualizări)).
- **Interfață** — dimensiunea și modul ferestrei sunt memorate, temă întunecată și luminoasă, 30 de limbi ale interfeței.

## Pornire rapidă

Descărcați `CoreDeck-Portable.zip` din **Releases** și dezarhivați-l oriunde. Programul este portabil și nu necesită instalare: rulați `CoreDeck.exe`. La prima pornire, folderul `data\` este creat lângă exe (vezi [Folderul de date](#folderul-de-date)).

Pentru actualizare manuală în locul celei integrate: închideți CoreDeck și înlocuiți `CoreDeck.exe` (și `data\tools\disk-helper.exe`). Folderul `data\` rămâne neschimbat.

## Structura portabilă

Structura recomandată pe un stick USB este `<drive>\_Launcher\CoreDeck.exe`, cu biblioteca de programe în foldere aflate lângă `_Launcher`. Căile din catalog sunt stocate relativ la exe (`../…`), iar căile implicite pentru import/export, programe și drivere sunt derivate din unitatea de pe care rulează CoreDeck, astfel încât proiectul nu are nevoie de o cale fixă și poate fi mutat pe altă unitate sau alt computer.

Dacă stick-ul USB primește altă literă, o cale salvată cu litera veche (de ex. `J:\…` după ce unitatea a devenit `L:`) este căutată pe unitatea de pe care rulează CoreDeck. La pornire, astfel de căi din `programs.json`, `drivers.json` și `settings.json` sunt rescrise o singură dată în formă relativă, păstrându-se o copie `*.drive-<date>.bak`.

## Folderul de date

La prima pornire, lângă exe se creează următoarele:

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

**Date private și un singur exe.** Toate datele utilizatorului sunt citite din `data\` de lângă `CoreDeck.exe`. Fișierele încorporate în exe (`cmd/launcher/web/data/*.json`) sunt doar șabloane goale, folosite când fișierul din `data\` nu există. Astfel, `CoreDeck.exe` public din Releases și o instalare privată sunt același fișier: catalogul privat, KnowledgeBase (`knowledge.json`) și site-ul `data\KnowledgeBase\` se află doar în `data\`. Un fișier corupt din `data\` este raportat ca eroare și nu este înlocuit cu șablonul gol (următoarea salvare nu va suprascrie datele). Testele se asigură că nicio dată privată nu ajunge în șabloanele încorporate sau în `cmd/launcher/web/`.

## Actualizări

Setări → **Actualizări**.

- **Canale:** Stabil (implicit) și Test (versiuni preliminare).
- **Verificare:** verificare opțională la pornire (cel mult o dată pe zi) și un buton manual **Caută actualizări**. Când există o versiune mai nouă, o fereastră din aplicație afișează lista de modificări, cu **Actualizează acum**, **Mai târziu** și **Omite această versiune**.
- **Descărcare și verificare:** `CoreDeck-Portable.zip` este descărcat din versiunea GitHub a `snakex21/CoreDeck`; dimensiunea și SHA-256 sunt verificate pe baza `SHA256SUMS.txt`. O nepotrivire oprește actualizarea și nu se modifică nimic.
- **Instalare:** sunt dezarhivate doar `CoreDeck.exe` și `data\tools\disk-helper.exe`, în folderul intermediar `.coredeck-update\` de lângă exe. Binarele anterioare sunt salvate în `.coredeck-update\backup\<version>\`, fișierele sunt înlocuite prin redenumirea celor vechi în `.old`, iar CoreDeck repornește.
- **Verificarea funcționării:** dacă noua versiune nu raportează o pornire corectă în 60 de secunde (sau se închide), versiunea anterioară este restaurată automat și pornită din nou. **Restaurează versiunea anterioară** este disponibil și în Setări.
- Funcționează în structura USB portabilă (`<drive>\_Launcher\CoreDeck.exe`), nu necesită drepturi de administrator pentru actualizarea propriu-zisă și nu modifică niciodată altceva în `data\`.
- Fără telemetrie: sunt contactate doar API-ul GitHub și descărcarea versiunii.

## Limba interfeței

Interfața este disponibilă în 30 de limbi: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 și 한국어.

Limba de afișare Windows este detectată automat. Selectorul de limbă se află în bara de sus și are opțiunea **Automat**. Textul care lipsește dintr-o traducere este afișat în engleză. Notele de versiune și [CHANGELOG](../CHANGELOG.md) sunt în poloneză și engleză.

## Compilare

Cerințe:

- Windows,
- Go 1.26 sau mai nou (`go.mod` fixează toolchain-ul `go1.26.11`, descărcat automat de `go`),
- un compilator C (MinGW gcc, necesar pentru WebView),
- opțional `go-winres` (pictograma și informațiile de versiune din exe),
- Node.js,
- Python.

```bat
build.bat
```

Aceasta verifică sintaxa JavaScript, rulează testele Go și produce local `CoreDeck.exe` și `data`. Versiunile de lansare sunt compilate fără simboluri (`-trimpath -ldflags "-s -w"`), iar `build.bat` scrie versiunea, numele produsului și drepturile de autor în resursele Windows. Variantă cu consolă de diagnosticare:

```bat
build.bat debug
```

Instrumentele pentru dezvoltatori WebView2 pot fi activate fără o compilare separată:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` creează arhiva de lansare `dist\CoreDeck-Portable.zip`.

## Teste

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Teste în browser: rulați `python -m http.server 8765` în folderul depozitului și deschideți `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (dialoguri și notificări, „PASS n/n” în titlu) sau `test/web/launcher-smoke.html` (lansatorul cu un API simulat).

CI (GitHub Actions) rulează verificarea sintaxei JS, testele Python, `go vet`, `go test -race`, `govulncheck` și compilarea.

## Structura proiectului

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Datele goale implicite încorporate în exe se află în `cmd/launcher/web/data/`. Datele create în timpul utilizării programului sunt locale și ignorate de Git.

## Confidențialitate

Programul funcționează local. Depozitul nu conține catalogul utilizatorului, istoric, pagini private, modele AI sau o bibliotecă de programe de instalare. Singurul acces la rețea al CoreDeck însuși este verificarea și descărcarea actualizărilor, care contactează doar API-ul GitHub și versiunile oficiale ale proiectului de pe GitHub.

## Securitate

- Serverul local al interfeței ascultă doar pe `127.0.0.1`, răspunde doar la antetul `Host` `localhost:<port>` / `127.0.0.1:<port>` și nu trimite antete CORS (totul este same-origin).
- Ajustările provin din `data\tweaks.json` (import, editare) și rulează cu drepturi de administrator. Înainte ca o ajustare să ruleze prima dată și după fiecare modificare a comenzii sale, configuratorul afișează scriptul exact și cere confirmare. Aprobările (SHA-256 al tipului și al comenzii) sunt stocate în `data\state\tweak_approvals.json`; programul rulează comenzi din catalog, nu ceea ce trimite pagina.
- Actualizările sunt instalate doar după ce dimensiunea și SHA-256 corespund cu `SHA256SUMS.txt` (vezi [Actualizări](#actualizări)).

## Profilul WebView2

Începând cu versiunea 2.3.1, profilul WebView2 (temă, limbă, `localStorage`) se află în `data\webview\` lângă exe. Se folosește o copie locală, modificată, a `webview_go` (`third_party/webview_go`, descrisă în `PATCHED.md`). La prima pornire, un profil existent din `%APPDATA%\CoreDeck.exe` este copiat (fără cache); originalul rămâne pe loc și poate fi șters manual ulterior.

## Licență

CoreDeck este publicat sub licența MIT, vezi [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
