**Idiomas:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#compilación)

Un kit de herramientas portable para Windows para organizar aplicaciones, controladores, instaladores, ajustes del sistema, perfiles de configuración y documentación local. Versión **2.4.0** · cambios: [CHANGELOG.md](../CHANGELOG.md).

El repositorio es un punto de partida limpio. No contiene programas, instaladores, enlaces privados ni datos del autor. En el primer inicio, la aplicación crea archivos vacíos en la carpeta `data`; el contenido se añade en la pantalla **Gestión**.

## Contenido

- [Características](#características)
- [Inicio rápido](#inicio-rápido)
- [Estructura portable](#estructura-portable)
- [Carpeta de datos](#carpeta-de-datos)
- [Actualizaciones](#actualizaciones)
- [Idioma de la interfaz](#idioma-de-la-interfaz)
- [Compilación](#compilación)
- [Pruebas](#pruebas)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Privacidad](#privacidad)
- [Seguridad](#seguridad)
- [Perfil de WebView2](#perfil-de-webview2)
- [Licencia](#licencia)

## Características

- **Catálogo**: tu propio catálogo de programas e instaladores, variantes de programas y perfiles de configuración, páginas separadas para controladores y configuración de programas, y tus propias páginas HTML.
- **Ejecución**: EXE, MSI, BAT, CMD y PowerShell. Los programas y herramientas portables tienen un botón **▶ Ejecutar** y se inician en su sitio desde el pendrive (como un doble clic en el Explorador, UAC incluido): el archivo definido en el catálogo o, si no lo hay, el único `.exe` coincidente de la carpeta (se omiten desinstaladores, instaladores y utilidades auxiliares, y se prefiere la versión de 64 bits). Si hay varios candidatos, CoreDeck pregunta cuál ejecutar; si solo hay un archivo comprimido o un instalador, abre la carpeta.
- **Resultados de los instaladores**: los códigos de salida 1602 (cancelado en el instalador) y 1223 (UAC rechazado) muestran un breve «Instalación cancelada»; los códigos MSI habituales (1603, 1618, 1619, 1638, 1639 y otros) reciben mensajes comprensibles; 3010/1641 se consideran éxito con un aviso de reinicio.
- **Diálogos integrados**: sin ventanas `alert/confirm/prompt` del navegador; los diálogos y las notificaciones funcionan con el teclado (Intro, Esc, Tab, restauración del foco). Ejecutar un instalador, un script o un programa no requiere confirmación; la confirmación se mantiene cuando se elimina algo o se modifica el sistema (eliminaciones, ajustes, instalación de muchos programas a la vez).
- **Base de Conocimiento** con comprobación de enlaces y un sitio opcional propio de Base de Conocimiento servido desde `data\KnowledgeBase\`.
- **Exportación e importación** de datos.
- **Configurador de Windows**: tweaks y debloat, con vista previa del script y aprobación (consulta [Seguridad](#seguridad)).
- **Registro de auditoría** de las operaciones administrativas en `data/logs/audit.jsonl`.
- **Actualizaciones** desde las versiones publicadas en GitHub con canales estable y de prueba, verificación de sumas de comprobación y reversión automática (consulta [Actualizaciones](#actualizaciones)).
- **Interfaz**: se recuerdan el tamaño y el modo de la ventana, tema oscuro y claro, 30 idiomas de interfaz.

## Inicio rápido

Descarga `CoreDeck-Portable.zip` desde **Releases** y descomprímelo en cualquier ubicación. El programa es portable y no necesita instalación: ejecuta `CoreDeck.exe`. En el primer inicio se crea la carpeta `data\` junto al exe (consulta [Carpeta de datos](#carpeta-de-datos)).

Para actualizar manualmente en lugar de usar el actualizador integrado: cierra CoreDeck y sustituye `CoreDeck.exe` (y `data\tools\disk-helper.exe`). La carpeta `data\` se deja tal cual.

## Estructura portable

La estructura prevista en un pendrive es `<drive>\_Launcher\CoreDeck.exe`, con la biblioteca de programas en carpetas junto a `_Launcher`. Las rutas del catálogo se guardan de forma relativa al exe (`../…`), y las rutas predeterminadas de importación/exportación, programas y controladores se derivan de la unidad desde la que se ejecuta CoreDeck, por lo que el proyecto no necesita una ruta fija y puede trasladarse a otra unidad u otro equipo.

Si el pendrive recibe otra letra, una ruta guardada con la letra antigua (p. ej., `J:\…` después de que la unidad pase a ser `L:`) se busca en la unidad desde la que se ejecuta CoreDeck. Al iniciar, esas rutas en `programs.json`, `drivers.json` y `settings.json` se reescriben una sola vez en forma relativa, conservando una copia `*.drive-<date>.bak`.

## Carpeta de datos

En el primer inicio se crea lo siguiente junto al exe:

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (opcional: tu propio sitio de Base de Conocimiento, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (perfil de WebView2: tema, idioma, localStorage)
├── logs/
└── state/
```

**Datos privados y un único exe.** Todos los datos del usuario se leen de `data\` junto a `CoreDeck.exe`. Los archivos integrados en el exe (`cmd/launcher/web/data/*.json`) son solo plantillas vacías, que se usan cuando el archivo en `data\` no existe. Gracias a esto, el `CoreDeck.exe` público de Releases y una instalación privada son el mismo archivo: el catálogo privado, la Base de Conocimiento (`knowledge.json`) y el sitio `data\KnowledgeBase\` residen solo en `data\`. Un archivo dañado en `data\` se notifica como error y no se sustituye por la plantilla vacía (el siguiente guardado no sobrescribirá los datos). Las pruebas garantizan que ningún dato privado acabe en las plantillas integradas ni en `cmd/launcher/web/`.

## Actualizaciones

Ajustes → **Actualizaciones**.

- **Canales:** Estable (predeterminado) y Prueba (versiones preliminares).
- **Búsqueda:** búsqueda opcional al iniciar (como máximo una vez al día) y un botón **Buscar actualizaciones** para buscarlas manualmente. Cuando existe una versión más reciente, una ventana de la aplicación muestra su registro de cambios con **Actualizar ahora**, **Más tarde** y **Omitir esta versión**.
- **Descarga y verificación:** `CoreDeck-Portable.zip` se descarga de la versión publicada en GitHub de `snakex21/CoreDeck`; su tamaño y su SHA-256 se comprueban con `SHA256SUMS.txt`. Si no coinciden, la actualización se cancela y no se cambia nada.
- **Instalación:** solo se descomprimen `CoreDeck.exe` y `data\tools\disk-helper.exe`, en la carpeta de preparación `.coredeck-update\` junto al exe. Los binarios anteriores se guardan como copia de seguridad en `.coredeck-update\backup\<version>\`, los archivos se intercambian renombrando los antiguos a `.old` y CoreDeck se reinicia.
- **Comprobación de arranque:** si la nueva versión no informa de un inicio correcto en 60 segundos (o se cierra), la versión anterior se restaura automáticamente y se vuelve a iniciar. **Restaurar la versión anterior** también está disponible en Ajustes.
- Funciona en la estructura portable en USB (`<drive>\_Launcher\CoreDeck.exe`), no necesita derechos de administrador para la propia actualización y nunca cambia nada más en `data\`.
- Sin telemetría: solo se contactan la API de GitHub y la descarga de la versión publicada.

## Idioma de la interfaz

La interfaz está disponible en 30 idiomas: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 y 한국어.

El idioma de visualización de Windows se detecta automáticamente. El selector de idioma está en la barra superior y tiene la opción **Automático**. Los textos que faltan en una traducción se muestran en inglés. Las notas de la versión y el [CHANGELOG](../CHANGELOG.md) están en polaco e inglés.

## Compilación

Requisitos:

- Windows,
- Go 1.26 o posterior (`go.mod` fija la cadena de herramientas `go1.26.11`, que `go` descarga automáticamente),
- un compilador de C (MinGW gcc, necesario para WebView),
- opcionalmente `go-winres` (icono e información de versión en el exe),
- Node.js,
- Python.

```bat
build.bat
```

Esto comprueba la sintaxis de JavaScript, ejecuta las pruebas de Go y genera localmente `CoreDeck.exe` y `data`. Las compilaciones de versión se generan sin símbolos (`-trimpath -ldflags "-s -w"`), y `build.bat` escribe la versión, el nombre del producto y el copyright en los recursos de Windows. Variante con consola de diagnóstico:

```bat
build.bat debug
```

Las herramientas de desarrollo de WebView2 pueden activarse sin una compilación aparte:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` crea el archivo de la versión `dist\CoreDeck-Portable.zip`.

## Pruebas

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Pruebas en el navegador: ejecuta `python -m http.server 8765` en la carpeta del repositorio y abre `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (diálogos y notificaciones, «PASS n/n» en el título) o `test/web/launcher-smoke.html` (el lanzador con una API simulada).

La CI (GitHub Actions) ejecuta la comprobación de sintaxis de JS, las pruebas de Python, `go vet`, `go test -race`, `govulncheck` y la compilación.

## Estructura del proyecto

```text
cmd/launcher/         aplicación e interfaz integrada
cmd/disk-helper/      utilidad para localizar la unidad
internal/             lógica del programa
scripts/              compilación y empaquetado
test/                 pruebas
```

Los datos vacíos predeterminados integrados en el exe están en `cmd/launcher/web/data/`. Los datos creados al usar el programa son locales y Git los ignora.

## Privacidad

El programa funciona localmente. El repositorio no contiene catálogo de usuario, historial, páginas privadas, modelos de IA ni biblioteca de instaladores. El único acceso a la red del propio CoreDeck es la búsqueda y descarga de actualizaciones, que solo contactan con la API de GitHub y las versiones oficiales del proyecto publicadas en GitHub.

## Seguridad

- El servidor local de la interfaz escucha solo en `127.0.0.1`, responde solo al encabezado `Host` `localhost:<port>` / `127.0.0.1:<port>` y no envía encabezados CORS (todo es same-origin).
- Los tweaks proceden de `data\tweaks.json` (importación, edición) y se ejecutan con derechos de administrador. Antes de que un tweak se ejecute por primera vez, y después de cada cambio en su comando, el configurador muestra el script exacto y pide confirmación. Las aprobaciones (SHA-256 del tipo y del comando) se guardan en `data\state\tweak_approvals.json`; el programa ejecuta los comandos del catálogo, no lo que envíe la página.
- Las actualizaciones solo se instalan cuando el tamaño y el SHA-256 coinciden con `SHA256SUMS.txt` (consulta [Actualizaciones](#actualizaciones)).

## Perfil de WebView2

Desde la versión 2.3.1, el perfil de WebView2 (tema, idioma, `localStorage`) se encuentra en `data\webview\` junto al exe. Se usa una copia local y parcheada de `webview_go` (`third_party/webview_go`, descrita en `PATCHED.md`). En el primer inicio se copia un perfil existente de `%APPDATA%\CoreDeck.exe` (sin la caché); el original se queda en su sitio y puede eliminarse manualmente más tarde.

## Licencia

CoreDeck se publica bajo la licencia MIT; consulta [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
