**Idiomas:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#compilação)

Um kit de ferramentas portátil para Windows para organizar aplicativos, drivers, instaladores, ajustes do sistema, perfis de configuração e documentação local. Versão **2.4.0** · alterações: [CHANGELOG.md](../CHANGELOG.md).

O repositório é um ponto de partida limpo. Ele não contém programas, instaladores, links privados nem dados do autor. Na primeira execução, o aplicativo cria arquivos vazios na pasta `data`; o conteúdo é adicionado na tela **Gestão**.

## Conteúdo

- [Recursos](#recursos)
- [Início rápido](#início-rápido)
- [Estrutura portátil](#estrutura-portátil)
- [Pasta de dados](#pasta-de-dados)
- [Atualizações](#atualizações)
- [Idioma da interface](#idioma-da-interface)
- [Compilação](#compilação)
- [Testes](#testes)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Privacidade](#privacidade)
- [Segurança](#segurança)
- [Perfil do WebView2](#perfil-do-webview2)
- [Licença](#licença)

## Recursos

- **Catálogo** – seu próprio catálogo de programas e instaladores, variantes de programas e perfis de configuração, páginas separadas para drivers e configurações de programas, suas próprias páginas HTML.
- **Execução** – EXE, MSI, BAT, CMD e PowerShell. Programas e ferramentas portáteis têm um botão **▶ Executar** e são iniciados no próprio local a partir do pendrive (como um clique duplo no Explorador de Arquivos, UAC incluído): o arquivo definido no catálogo ou, se não houver, o único `.exe` correspondente na pasta (desinstaladores, instaladores e utilitários auxiliares são ignorados, e a versão de 64 bits tem preferência). Com vários candidatos, o CoreDeck pergunta qual executar; se houver apenas um arquivo compactado ou um instalador, ele abre a pasta.
- **Resultados dos instaladores** – os códigos de saída 1602 (cancelado no instalador) e 1223 (UAC recusado) exibem um breve "Instalação cancelada"; os códigos MSI comuns (1603, 1618, 1619, 1638, 1639 e outros) recebem mensagens compreensíveis; 3010/1641 indicam sucesso com um aviso de reinicialização.
- **Caixas de diálogo no aplicativo** – nenhuma janela `alert/confirm/prompt` do navegador; caixas de diálogo e notificações funcionam pelo teclado (Enter, Esc, Tab, restauração do foco). Executar um instalador, script ou programa não exige confirmação; a confirmação permanece onde algo é excluído ou o sistema é alterado (exclusões, ajustes, instalação de muitos programas de uma vez).
- **Base de Conhecimento** com verificação de links e um site opcional próprio de Base de Conhecimento servido a partir de `data\KnowledgeBase\`.
- **Exportação e importação** de dados.
- **Configurador do Windows** – tweaks e debloat, com visualização do script e aprovação (veja [Segurança](#segurança)).
- **Log de auditoria** das operações administrativas em `data/logs/audit.jsonl`.
- **Atualizações** a partir das releases do GitHub com canais estável e de teste, verificação de checksums e reversão automática (veja [Atualizações](#atualizações)).
- **Interface** – tamanho e modo da janela são lembrados, tema escuro e claro, 30 idiomas de interface.

## Início rápido

Baixe `CoreDeck-Portable.zip` em **Releases** e descompacte em qualquer lugar. O programa é portátil e não precisa de instalação: execute `CoreDeck.exe`. Na primeira execução, a pasta `data\` é criada ao lado do exe (veja [Pasta de dados](#pasta-de-dados)).

Para atualizar manualmente em vez de usar o atualizador integrado: feche o CoreDeck e substitua `CoreDeck.exe` (e `data\tools\disk-helper.exe`). A pasta `data\` permanece como está.

## Estrutura portátil

A estrutura prevista em um pendrive é `<drive>\_Launcher\CoreDeck.exe`, com a biblioteca de programas em pastas ao lado de `_Launcher`. Os caminhos do catálogo são salvos de forma relativa ao exe (`../…`), e os caminhos padrão de importação/exportação, programas e drivers são derivados da unidade a partir da qual o CoreDeck é executado. Assim, o projeto não precisa de um caminho fixo e pode ser movido para outra unidade ou outro computador.

Se o pendrive receber outra letra, um caminho salvo com a letra antiga (por exemplo, `J:\…` depois que a unidade passou a ser `L:`) é procurado na unidade a partir da qual o CoreDeck é executado. Na inicialização, esses caminhos em `programs.json`, `drivers.json` e `settings.json` são reescritos uma única vez em forma relativa, mantendo uma cópia `*.drive-<date>.bak`.

## Pasta de dados

Na primeira execução, o seguinte é criado ao lado do exe:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (opcional: seu próprio site de Base de Conhecimento, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (perfil do WebView2: tema, idioma, localStorage)
├── logs/
└── state/
```

**Dados privados e um único exe.** Todos os dados do usuário são lidos de `data\` ao lado de `CoreDeck.exe`. Os arquivos incorporados ao exe (`cmd/launcher/web/data/*.json`) são apenas modelos vazios, usados quando o arquivo em `data\` não existe. Com isso, o `CoreDeck.exe` público das Releases e uma instalação privada são o mesmo arquivo: o catálogo privado, a Base de Conhecimento (`knowledge.json`) e o site `data\KnowledgeBase\` ficam apenas em `data\`. Um arquivo corrompido em `data\` é relatado como erro e não é substituído pelo modelo vazio (o próximo salvamento não sobrescreverá os dados). Os testes garantem que nenhum dado privado vá parar nos modelos incorporados ou em `cmd/launcher/web/`.

## Atualizações

Configurações → **Atualizações**.

- **Canais:** Estável (padrão) e Teste (pré-lançamentos).
- **Verificação:** verificação opcional na inicialização (no máximo uma vez por dia) e um botão **Verificar atualizações** para a verificação manual. Quando existe uma versão mais recente, uma janela do aplicativo mostra o changelog dela com **Atualizar agora**, **Mais tarde** e **Pular esta versão**.
- **Download e verificação:** `CoreDeck-Portable.zip` é baixado da release do GitHub de `snakex21/CoreDeck`; o tamanho e o SHA-256 são conferidos com `SHA256SUMS.txt`. Se não coincidirem, a atualização é interrompida e nada é alterado.
- **Instalação:** somente `CoreDeck.exe` e `data\tools\disk-helper.exe` são descompactados, na pasta de preparação `.coredeck-update\` ao lado do exe. Os binários anteriores são salvos como backup em `.coredeck-update\backup\<version>\`, os arquivos são trocados renomeando os antigos para `.old`, e o CoreDeck reinicia.
- **Verificação de inicialização:** se a nova versão não informar uma inicialização correta em 60 segundos (ou se encerrar), a versão anterior é restaurada automaticamente e iniciada de novo. **Restaurar a versão anterior** também está disponível nas Configurações.
- Funciona na estrutura portátil em USB (`<drive>\_Launcher\CoreDeck.exe`), não precisa de direitos de administrador para a atualização em si e nunca altera nada mais em `data\`.
- Sem telemetria: apenas a API do GitHub e o download da release são contatados.

## Idioma da interface

A interface está disponível em 30 idiomas: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 e 한국어.

O idioma de exibição do Windows é detectado automaticamente. O seletor de idioma fica na barra superior e tem a opção **Automático**. Textos que faltam em uma tradução são exibidos em inglês. As notas de versão e o [CHANGELOG](../CHANGELOG.md) estão em polonês e inglês.

## Compilação

Requisitos:

- Windows,
- Go 1.26 ou mais recente (`go.mod` fixa a toolchain `go1.26.11`, baixada automaticamente pelo `go`),
- um compilador C (MinGW gcc, exigido pelo WebView),
- opcionalmente `go-winres` (ícone e informações de versão no exe),
- Node.js,
- Python.

```bat
build.bat
```

Isso verifica a sintaxe JavaScript, executa os testes Go e gera localmente `CoreDeck.exe` e `data`. As builds de release são geradas sem símbolos (`-trimpath -ldflags "-s -w"`), e `build.bat` grava a versão, o nome do produto e o copyright nos recursos do Windows. Variante com console de diagnóstico:

```bat
build.bat debug
```

As ferramentas de desenvolvedor do WebView2 podem ser ativadas sem uma build separada:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` cria o arquivo de release `dist\CoreDeck-Portable.zip`.

## Testes

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Testes no navegador: execute `python -m http.server 8765` na pasta do repositório e abra `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (caixas de diálogo e notificações, "PASS n/n" no título) ou `test/web/launcher-smoke.html` (o launcher com uma API simulada).

A CI (GitHub Actions) executa a verificação de sintaxe JS, os testes Python, `go vet`, `go test -race`, `govulncheck` e a build.

## Estrutura do projeto

```text
cmd/launcher/         aplicativo e interface incorporada
cmd/disk-helper/      utilitário para localizar a unidade
internal/             lógica do programa
scripts/              build e empacotamento
test/                 testes
```

Os dados vazios padrão incorporados ao exe ficam em `cmd/launcher/web/data/`. Os dados criados durante o uso do programa são locais e ignorados pelo Git.

## Privacidade

O programa funciona localmente. O repositório não contém catálogo de usuário, histórico, páginas privadas, modelos de IA nem biblioteca de instaladores. O único acesso à rede do próprio CoreDeck é a verificação e o download de atualizações, que contatam apenas a API do GitHub e as releases oficiais do projeto no GitHub.

## Segurança

- O servidor local da interface escuta apenas em `127.0.0.1`, responde apenas ao cabeçalho `Host` `localhost:<port>` / `127.0.0.1:<port>` e não envia cabeçalhos CORS (tudo é same-origin).
- Os tweaks vêm de `data\tweaks.json` (importação, edição) e são executados com direitos de administrador. Antes de um tweak ser executado pela primeira vez, e após cada alteração do seu comando, o configurador mostra o script exato e pede confirmação. As aprovações (SHA-256 do tipo e do comando) são armazenadas em `data\state\tweak_approvals.json`; o programa executa os comandos do catálogo, e não o que a página envia.
- As atualizações só são instaladas depois que o tamanho e o SHA-256 coincidem com `SHA256SUMS.txt` (veja [Atualizações](#atualizações)).

## Perfil do WebView2

Desde a versão 2.3.1, o perfil do WebView2 (tema, idioma, `localStorage`) fica em `data\webview\` ao lado do exe. É usada uma cópia local e modificada de `webview_go` (`third_party/webview_go`, descrita em `PATCHED.md`). Na primeira execução, um perfil existente em `%APPDATA%\CoreDeck.exe` é copiado (sem o cache); o original permanece no lugar e pode ser excluído manualmente depois.

## Licença

O CoreDeck é distribuído sob a licença MIT, veja [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
