**Kielet:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#kääntäminen)

Siirrettävä Windows-työkalupakki ohjelmien, ajurien, asennusohjelmien, järjestelmän säätöjen, asetusprofiilien ja paikallisen dokumentaation järjestämiseen. Versio **2.4.0** · muutokset: [CHANGELOG.md](../CHANGELOG.md).

Repositorio on puhdas lähtökohta. Se ei sisällä ohjelmia, asennusohjelmia, yksityisiä linkkejä eikä tekijän tietoja. Ensimmäisellä käynnistyskerralla sovellus luo tyhjät tiedostot `data`-kansioon; sisältöä lisätään **Hallinta**-näkymässä.

## Sisältö

- [Ominaisuudet](#ominaisuudet)
- [Pika-aloitus](#pika-aloitus)
- [Siirrettävä rakenne](#siirrettävä-rakenne)
- [Datakansio](#datakansio)
- [Päivitykset](#päivitykset)
- [Käyttöliittymän kieli](#käyttöliittymän-kieli)
- [Kääntäminen](#kääntäminen)
- [Testit](#testit)
- [Projektin rakenne](#projektin-rakenne)
- [Yksityisyys](#yksityisyys)
- [Tietoturva](#tietoturva)
- [WebView2-profiili](#webview2-profiili)
- [Lisenssi](#lisenssi)

## Ominaisuudet

- **Luettelo** — oma luettelo ohjelmista ja asennusohjelmista, ohjelmavariantit ja asetusprofiilit, erilliset sivut ajureille ja ohjelmien asetuksille, omat HTML-sivut.
- **Käynnistäminen** — EXE, MSI, BAT, CMD ja PowerShell. Siirrettävissä ohjelmissa ja työkaluissa on **▶ Käynnistä**-painike, ja ne käynnistyvät suoraan muistitikulta (kuin kaksoisnapsautus Resurssienhallinnassa, UAC mukaan lukien): luetteloon asetettu tiedosto, muuten kansion ainoa sopiva `.exe` (asennuksen poisto-ohjelmat, asennusohjelmat ja apuohjelmat ohitetaan, 64-bittistä suositaan). Jos ehdokkaita on useita, CoreDeck kysyy, mikä käynnistetään; jos kansiossa on vain arkisto tai asennusohjelma, kansio avataan.
- **Asennusohjelmien tulokset** — paluukoodit 1602 (peruttu asennusohjelmassa) ja 1223 (UAC hylätty) näyttävät lyhyen ilmoituksen "Asennus peruttu"; yleiset MSI-koodit (1603, 1618, 1619, 1638, 1639 ja muita) saavat selkeät viestit; 3010/1641 ovat onnistumisia, joihin liittyy huomautus uudelleenkäynnistyksestä.
- **Sovelluksen omat valintaikkunat** — ei selaimen `alert/confirm/prompt`-ikkunoita; valintaikkunat ja ilmoitukset toimivat näppäimistöllä (Enter, Esc, Tab, kohdistuksen palautus). Asennusohjelman, skriptin tai ohjelman käynnistäminen ei vaadi vahvistusta; vahvistus säilyy siellä, missä jotain poistetaan tai järjestelmää muutetaan (poistot, säädöt, monen ohjelman asentaminen kerralla).
- **KnowledgeBase** linkkien tarkistuksella sekä valinnainen oma KnowledgeBase-sivusto, joka tarjoillaan kansiosta `data\KnowledgeBase\`.
- **Tietojen vienti ja tuonti.**
- **Windows-määrittäjä** — säädöt ja debloat, skriptin esikatselu ja hyväksyntä (ks. [Tietoturva](#tietoturva)).
- **Valvontaloki** hallinnollisista toiminnoista tiedostossa `data/logs/audit.jsonl`.
- **Päivitykset** GitHub-julkaisuista vakaalla ja testikanavalla, tarkistussummien varmennuksella ja automaattisella palautuksella (ks. [Päivitykset](#päivitykset)).
- **Käyttöliittymä** — ikkunan koko ja tila muistetaan, tumma ja vaalea teema, 30 käyttöliittymän kieltä.

## Pika-aloitus

Lataa `CoreDeck-Portable.zip` kohdasta **Releases** ja pura se mihin tahansa. Ohjelma on siirrettävä eikä vaadi asennusta: käynnistä `CoreDeck.exe`. Ensimmäisellä käynnistyskerralla exe-tiedoston viereen luodaan `data\`-kansio (ks. [Datakansio](#datakansio)).

Manuaalinen päivitys sisäänrakennetun päivittäjän sijaan: sulje CoreDeck ja korvaa `CoreDeck.exe` (sekä `data\tools\disk-helper.exe`). `data\`-kansio jätetään ennalleen.

## Siirrettävä rakenne

Muistitikun suunniteltu rakenne on `<asema>\_Launcher\CoreDeck.exe`, ja ohjelmakirjasto on kansioissa `_Launcher`-kansion vieressä. Luettelon polut tallennetaan suhteessa exe-tiedostoon (`../…`), ja tuonnin/viennin, ohjelmien ja ajurien oletuspolut johdetaan asemasta, jolta CoreDeck käynnistetään, joten projekti ei tarvitse kiinteää polkua ja sen voi siirtää toiselle asemalle tai toiseen tietokoneeseen.

Jos muistitikku saa eri kirjaimen, vanhalla kirjaimella tallennettu polku (esim. `J:\…`, kun asema on nyt `L:`) haetaan asemalta, jolta CoreDeck käynnistetään. Käynnistyksessä tällaiset polut tiedostoissa `programs.json`, `drivers.json` ja `settings.json` kirjoitetaan kerran uudelleen suhteelliseen muotoon, ja niistä säilytetään kopio `*.drive-<päivämäärä>.bak`.

## Datakansio

Ensimmäisellä käynnistyskerralla exe-tiedoston viereen luodaan seuraavat:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (valinnainen: oma KnowledgeBase-sivusto, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (WebView2-profiili: teema, kieli, localStorage)
├── logs/
└── state/
```

**Yksityiset tiedot ja yksi exe-tiedosto.** Kaikki käyttäjän tiedot luetaan `CoreDeck.exe`-tiedoston vieressä olevasta `data\`-kansiosta. Exe-tiedostoon upotetut tiedostot (`cmd/launcher/web/data/*.json`) ovat vain tyhjiä malleja, joita käytetään, kun tiedostoa ei ole `data\`-kansiossa. Siksi julkinen `CoreDeck.exe` Releases-sivulta ja yksityinen asennus ovat sama tiedosto: yksityinen luettelo, KnowledgeBase (`knowledge.json`) ja `data\KnowledgeBase\`-sivusto ovat vain `data\`-kansiossa. Vioittunut tiedosto `data\`-kansiossa ilmoitetaan virheenä, eikä sitä korvata tyhjällä mallilla (seuraava tallennus ei korvaa tietoja). Testit varmistavat, ettei yksityisiä tietoja päädy upotettuihin malleihin eikä hakemistoon `cmd/launcher/web/`.

## Päivitykset

Asetukset → **Päivitykset**.

- **Kanavat:** Vakaa (oletus) ja Testi (esiversiot).
- **Tarkistus:** valinnainen tarkistus käynnistyksessä (enintään kerran päivässä) ja manuaalinen **Tarkista päivitykset** -painike. Kun uudempi versio on saatavilla, sovelluksen ikkuna näyttää sen muutoslokin ja painikkeet **Päivitä nyt**, **Myöhemmin** ja **Ohita tämä versio**.
- **Lataus ja varmennus:** `CoreDeck-Portable.zip` ladataan projektin `snakex21/CoreDeck` GitHub-julkaisusta; sen koko ja SHA-256 tarkistetaan tiedostoa `SHA256SUMS.txt` vasten. Jos ne eivät täsmää, päivitys keskeytetään eikä mitään muuteta.
- **Asennus:** vain `CoreDeck.exe` ja `data\tools\disk-helper.exe` puretaan väliaikaiskansioon `.coredeck-update\` exe-tiedoston viereen. Edelliset binäärit varmuuskopioidaan kansioon `.coredeck-update\backup\<version>\`, tiedostot vaihdetaan nimeämällä vanhat uudelleen `.old`-päätteisiksi, ja CoreDeck käynnistyy uudelleen.
- **Kuntotarkistus:** jos uusi versio ei ilmoita onnistuneesta käynnistyksestä 60 sekunnin kuluessa (tai se sulkeutuu), edellinen versio palautetaan automaattisesti ja käynnistetään uudelleen. **Palauta edellinen versio** on saatavilla myös asetuksissa.
- Toimii siirrettävässä USB-rakenteessa (`<asema>\_Launcher\CoreDeck.exe`), ei vaadi järjestelmänvalvojan oikeuksia itse päivitykseen eikä koskaan muuta mitään muuta `data\`-kansiossa.
- Ei telemetriaa: yhteys otetaan vain GitHub APIin ja julkaisun lataukseen.

## Käyttöliittymän kieli

Käyttöliittymä on saatavilla 30 kielellä: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 ja 한국어.

Windowsin näyttökieli tunnistetaan automaattisesti. Kielivalitsin on yläpalkissa, ja siinä on **Automaattinen**-vaihtoehto. Käännöksestä puuttuvat tekstit näytetään englanniksi. Julkaisutiedot ja [CHANGELOG](../CHANGELOG.md) ovat puolaksi ja englanniksi.

## Kääntäminen

Vaatimukset:

- Windows,
- Go 1.26 tai uudempi (`go.mod` kiinnittää työkaluketjun `go1.26.11`, jonka `go` lataa automaattisesti),
- C-kääntäjä (MinGW gcc, WebView vaatii sen),
- valinnaisesti `go-winres` (kuvake ja versiotiedot exe-tiedostoon),
- Node.js,
- Python.

```bat
build.bat
```

Tämä tarkistaa JavaScript-syntaksin, ajaa Go-testit ja tuottaa paikallisen `CoreDeck.exe`-tiedoston ja `data`-kansion. Julkaisukäännöksistä on poistettu symbolit (`-trimpath -ldflags "-s -w"`), ja `build.bat` kirjoittaa version, tuotenimen ja tekijänoikeustiedot Windows-resursseihin. Versio diagnostiikkakonsolilla:

```bat
build.bat debug
```

WebView2:n kehittäjätyökalut voi ottaa käyttöön ilman erillistä käännöstä:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` rakentaa julkaisuarkiston `dist\CoreDeck-Portable.zip`.

## Testit

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Selaintestit: suorita `python -m http.server 8765` repositorion kansiossa ja avaa `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (valintaikkunat ja ilmoitukset, "PASS n/n" otsikossa) tai `test/web/launcher-smoke.html` (käynnistin tynkä-API:lla).

CI (GitHub Actions) ajaa JS-syntaksin tarkistuksen, Python-testit, `go vet`, `go test -race`, `govulncheck` ja käännöksen.

## Projektin rakenne

```text
cmd/launcher/         sovellus ja upotettu käyttöliittymä
cmd/disk-helper/      apuohjelma asemien hakuun
internal/             ohjelmalogiikka
scripts/              kääntäminen ja paketointi
test/                 testit
```

Exe-tiedostoon upotetut tyhjät oletustiedot ovat kansiossa `cmd/launcher/web/data/`. Ohjelman käytön aikana luodut tiedot ovat paikallisia, ja Git ohittaa ne.

## Yksityisyys

Ohjelma toimii paikallisesti. Repositorio ei sisällä käyttäjän luetteloa, historiaa, yksityisiä sivuja, tekoälymalleja eikä asennusohjelmakirjastoa. CoreDeckin ainoa oma verkkoyhteys on päivitysten tarkistus ja lataus, jotka ottavat yhteyttä vain GitHub APIin ja projektin virallisiin GitHub-julkaisuihin.

## Tietoturva

- Paikallinen käyttöliittymäpalvelin kuuntelee vain osoitetta `127.0.0.1`, vastaa vain `Host`-otsakkeeseen `localhost:<port>` / `127.0.0.1:<port>` eikä lähetä CORS-otsakkeita (kaikki on same-origin).
- Säädöt tulevat tiedostosta `data\tweaks.json` (tuonti, muokkaus) ja suoritetaan järjestelmänvalvojan oikeuksin. Ennen kuin säätö suoritetaan ensimmäistä kertaa, ja jokaisen sen komennon muutoksen jälkeen, määrittäjä näyttää tarkan skriptin ja pyytää vahvistusta. Hyväksynnät (tyypin ja komennon SHA-256) tallennetaan tiedostoon `data\state\tweak_approvals.json`; ohjelma suorittaa komentoja luettelosta, ei mitä tahansa sivu lähettää.
- Päivitykset asennetaan vasta, kun koko ja SHA-256 vastaavat tiedostoa `SHA256SUMS.txt` (ks. [Päivitykset](#päivitykset)).

## WebView2-profiili

Versiosta 2.3.1 alkaen WebView2-profiili (teema, kieli, `localStorage`) on kansiossa `data\webview\` exe-tiedoston vieressä. Käytössä on paikallinen, paikattu kopio `webview_go`-kirjastosta (`third_party/webview_go`, kuvattu tiedostossa `PATCHED.md`). Ensimmäisellä käynnistyskerralla olemassa oleva profiili kopioidaan sijainnista `%APPDATA%\CoreDeck.exe` (ilman välimuistia); alkuperäinen jää paikalleen, ja sen voi myöhemmin poistaa käsin.

## Lisenssi

CoreDeck julkaistaan MIT-lisenssillä, ks. [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
