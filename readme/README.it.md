**Lingue:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#compilazione)

Un toolkit portatile per Windows per organizzare app, driver, programmi di installazione, tweak di sistema, profili di configurazione e documentazione locale. Versione **2.4.0** · modifiche: [CHANGELOG.md](../CHANGELOG.md).

Il repository è un punto di partenza pulito. Non contiene programmi, programmi di installazione, link privati né dati dell'autore. Al primo avvio l'applicazione crea file vuoti nella cartella `data`; i contenuti si aggiungono nella schermata **Gestione**.

## Indice

- [Funzionalità](#funzionalità)
- [Avvio rapido](#avvio-rapido)
- [Struttura portatile](#struttura-portatile)
- [Cartella dei dati](#cartella-dei-dati)
- [Aggiornamenti](#aggiornamenti)
- [Lingua dell'interfaccia](#lingua-dellinterfaccia)
- [Compilazione](#compilazione)
- [Test](#test)
- [Struttura del progetto](#struttura-del-progetto)
- [Privacy](#privacy)
- [Sicurezza](#sicurezza)
- [Profilo WebView2](#profilo-webview2)
- [Licenza](#licenza)

## Funzionalità

- **Catalogo** – un catalogo personale di programmi e programmi di installazione, varianti di programmi e profili di configurazione, pagine separate per i driver e per le impostazioni dei programmi, pagine HTML personalizzate.
- **Avvio** – EXE, MSI, BAT, CMD e PowerShell. I programmi e gli strumenti portatili hanno un pulsante **▶ Avvia** e partono direttamente dalla chiavetta USB (come un doppio clic in Esplora file, UAC compreso): il file impostato nel catalogo, altrimenti l'unico `.exe` corrispondente nella cartella (programmi di disinstallazione, di installazione e di supporto vengono ignorati, la versione a 64 bit è preferita). Se ci sono più candidati, CoreDeck chiede quale avviare; se c'è solo un archivio o un programma di installazione, apre la cartella.
- **Esiti dei programmi di installazione** – i codici di uscita 1602 (annullato nel programma di installazione) e 1223 (UAC rifiutato) mostrano un breve "Installazione annullata"; i codici MSI comuni (1603, 1618, 1619, 1638, 1639 e altri) ricevono messaggi comprensibili; 3010/1641 indicano un successo con una nota sul riavvio.
- **Finestre di dialogo integrate** – nessuna finestra `alert/confirm/prompt` del browser; finestre di dialogo e notifiche funzionano da tastiera (Invio, Esc, Tab, ripristino del focus). Avviare un programma di installazione, uno script o un programma non richiede conferma; la conferma resta dove qualcosa viene eliminato o il sistema viene modificato (eliminazioni, tweak, installazione di molti programmi insieme).
- **Knowledge Base** con verifica dei link e un sito Knowledge Base personale facoltativo, servito da `data\KnowledgeBase\`.
- **Esportazione e importazione** dei dati.
- **Configuratore di Windows** – tweak e debloat, con anteprima dello script e approvazione (vedi [Sicurezza](#sicurezza)).
- **Registro di audit** delle operazioni amministrative in `data/logs/audit.jsonl`.
- **Aggiornamenti** dalle release GitHub con canale stabile e di test, verifica dei checksum e rollback automatico (vedi [Aggiornamenti](#aggiornamenti)).
- **Interfaccia** – dimensioni e modalità della finestra vengono ricordate, tema scuro e chiaro, 30 lingue dell'interfaccia.

## Avvio rapido

Scarica `CoreDeck-Portable.zip` da **Releases** ed estrailo dove vuoi. Il programma è portatile e non richiede installazione: avvia `CoreDeck.exe`. Al primo avvio viene creata la cartella `data\` accanto all'exe (vedi [Cartella dei dati](#cartella-dei-dati)).

Per aggiornare manualmente invece di usare l'aggiornamento integrato: chiudi CoreDeck e sostituisci `CoreDeck.exe` (e `data\tools\disk-helper.exe`). La cartella `data\` resta invariata.

## Struttura portatile

La struttura prevista su una chiavetta USB è `<drive>\_Launcher\CoreDeck.exe`, con la libreria dei programmi in cartelle accanto a `_Launcher`. I percorsi del catalogo vengono salvati come relativi all'exe (`../…`) e i percorsi predefiniti di importazione/esportazione, programmi e driver sono ricavati dall'unità da cui viene eseguito CoreDeck: il progetto non richiede quindi un percorso fisso e può essere spostato su un'altra unità o un altro computer.

Se la chiavetta USB riceve una lettera diversa, un percorso salvato con la vecchia lettera (ad es. `J:\…` dopo che l'unità è diventata `L:`) viene cercato sull'unità da cui viene eseguito CoreDeck. All'avvio questi percorsi in `programs.json`, `drivers.json` e `settings.json` vengono riscritti una sola volta in forma relativa, conservando una copia `*.drive-<date>.bak`.

## Cartella dei dati

Al primo avvio accanto all'exe viene creato quanto segue:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (facoltativo: sito Knowledge Base personale, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (profilo WebView2: tema, lingua, localStorage)
├── logs/
└── state/
```

**Dati privati e un solo exe.** Tutti i dati dell'utente vengono letti da `data\` accanto a `CoreDeck.exe`. I file incorporati nell'exe (`cmd/launcher/web/data/*.json`) sono solo modelli vuoti, usati quando il file in `data\` non esiste. In questo modo il `CoreDeck.exe` pubblico delle Releases e un'installazione privata sono lo stesso file: il catalogo privato, la Knowledge Base (`knowledge.json`) e il sito `data\KnowledgeBase\` si trovano solo in `data\`. Un file danneggiato in `data\` viene segnalato come errore e non viene sostituito dal modello vuoto (il salvataggio successivo non sovrascriverà i dati). I test verificano che nessun dato privato finisca nei modelli incorporati o in `cmd/launcher/web/`.

## Aggiornamenti

Impostazioni → **Aggiornamenti**.

- **Canali:** Stabile (predefinito) e Test (versioni preliminari).
- **Ricerca:** ricerca facoltativa all'avvio (al massimo una volta al giorno) e un pulsante **Cerca aggiornamenti** per la ricerca manuale. Quando esiste una versione più recente, una finestra dell'app ne mostra il changelog con **Aggiorna ora**, **Più tardi** e **Salta questa versione**.
- **Download e verifica:** `CoreDeck-Portable.zip` viene scaricato dalla release GitHub di `snakex21/CoreDeck`; dimensione e SHA-256 vengono confrontati con `SHA256SUMS.txt`. In caso di mancata corrispondenza l'aggiornamento viene interrotto e non viene modificato nulla.
- **Installazione:** vengono estratti solo `CoreDeck.exe` e `data\tools\disk-helper.exe`, nella cartella di staging `.coredeck-update\` accanto all'exe. I binari precedenti vengono salvati in `.coredeck-update\backup\<version>\`, i file vengono sostituiti rinominando quelli vecchi in `.old` e CoreDeck si riavvia.
- **Controllo di avvio:** se la nuova versione non segnala un avvio corretto entro 60 secondi (o si chiude), la versione precedente viene ripristinata automaticamente e riavviata. **Ripristina la versione precedente** è disponibile anche nelle Impostazioni.
- Funziona nella struttura portatile su USB (`<drive>\_Launcher\CoreDeck.exe`), non richiede diritti di amministratore per l'aggiornamento stesso e non modifica mai nient'altro in `data\`.
- Nessuna telemetria: vengono contattati solo l'API di GitHub e il download della release.

## Lingua dell'interfaccia

L'interfaccia è disponibile in 30 lingue: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 e 한국어.

La lingua di visualizzazione di Windows viene rilevata automaticamente. Il selettore della lingua si trova nella barra superiore e ha l'opzione **Automatica**. I testi mancanti in una traduzione vengono mostrati in inglese. Le note di rilascio e il [CHANGELOG](../CHANGELOG.md) sono in polacco e in inglese.

## Compilazione

Requisiti:

- Windows,
- Go 1.26 o successivo (`go.mod` fissa la toolchain `go1.26.11`, scaricata automaticamente da `go`),
- un compilatore C (MinGW gcc, richiesto da WebView),
- facoltativamente `go-winres` (icona e informazioni sulla versione nell'exe),
- Node.js,
- Python.

```bat
build.bat
```

Questo comando verifica la sintassi JavaScript, esegue i test Go e produce localmente `CoreDeck.exe` e `data`. Le build di release sono prive di simboli (`-trimpath -ldflags "-s -w"`) e `build.bat` scrive versione, nome del prodotto e copyright nelle risorse di Windows. Variante con console di diagnostica:

```bat
build.bat debug
```

Gli strumenti di sviluppo di WebView2 possono essere attivati senza una build separata:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` crea l'archivio di release `dist\CoreDeck-Portable.zip`.

## Test

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Test nel browser: esegui `python -m http.server 8765` nella cartella del repository e apri `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (finestre di dialogo e notifiche, "PASS n/n" nel titolo) oppure `test/web/launcher-smoke.html` (il launcher con un'API simulata).

La CI (GitHub Actions) esegue il controllo della sintassi JS, i test Python, `go vet`, `go test -race`, `govulncheck` e la build.

## Struttura del progetto

```text
cmd/launcher/         applicazione e interfaccia incorporata
cmd/disk-helper/      programma di supporto per la ricerca dell'unità
internal/             logica del programma
scripts/              build e creazione dei pacchetti
test/                 test
```

I dati vuoti predefiniti incorporati nell'exe si trovano in `cmd/launcher/web/data/`. I dati creati durante l'uso del programma sono locali e ignorati da Git.

## Privacy

Il programma funziona in locale. Il repository non contiene cataloghi utente, cronologia, pagine private, modelli di IA né librerie di programmi di installazione. L'unico accesso alla rete di CoreDeck stesso è la ricerca e il download degli aggiornamenti, che contattano solo l'API di GitHub e le release GitHub ufficiali del progetto.

## Sicurezza

- Il server locale dell'interfaccia è in ascolto solo su `127.0.0.1`, risponde solo all'header `Host` `localhost:<port>` / `127.0.0.1:<port>` e non invia header CORS (tutto è same-origin).
- I tweak provengono da `data\tweaks.json` (importazione, modifica) e vengono eseguiti con diritti di amministratore. Prima che un tweak venga eseguito per la prima volta, e dopo ogni modifica del suo comando, il configuratore mostra lo script esatto e chiede conferma. Le approvazioni (SHA-256 del tipo e del comando) vengono salvate in `data\state\tweak_approvals.json`; il programma esegue i comandi del catalogo, non ciò che invia la pagina.
- Gli aggiornamenti vengono installati solo dopo che dimensione e SHA-256 corrispondono a `SHA256SUMS.txt` (vedi [Aggiornamenti](#aggiornamenti)).

## Profilo WebView2

Dalla versione 2.3.1 il profilo WebView2 (tema, lingua, `localStorage`) si trova in `data\webview\` accanto all'exe. Viene usata una copia locale con patch di `webview_go` (`third_party/webview_go`, descritta in `PATCHED.md`). Al primo avvio un profilo esistente in `%APPDATA%\CoreDeck.exe` viene copiato (senza la cache); l'originale resta al suo posto e può essere eliminato manualmente in seguito.

## Licenza

CoreDeck è distribuito con licenza MIT, vedi [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
