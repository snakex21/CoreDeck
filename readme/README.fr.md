**Langues :** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#compilation)

Une boîte à outils portable pour Windows qui sert à organiser applications, pilotes, installateurs, réglages système, profils de configuration et documentation locale. Version **2.4.0** · modifications : [CHANGELOG.md](../CHANGELOG.md).

Le dépôt est un point de départ vierge. Il ne contient ni programmes, ni installateurs, ni liens privés, ni données de l'auteur. Au premier démarrage, l'application crée des fichiers vides dans le dossier `data` ; le contenu s'ajoute dans l'écran **Gestion**.

## Sommaire

- [Fonctionnalités](#fonctionnalités)
- [Démarrage rapide](#démarrage-rapide)
- [Organisation portable](#organisation-portable)
- [Dossier de données](#dossier-de-données)
- [Mises à jour](#mises-à-jour)
- [Langue de l'interface](#langue-de-linterface)
- [Compilation](#compilation)
- [Tests](#tests)
- [Structure du projet](#structure-du-projet)
- [Confidentialité](#confidentialité)
- [Sécurité](#sécurité)
- [Profil WebView2](#profil-webview2)
- [Licence](#licence)

## Fonctionnalités

- **Catalogue** – votre propre catalogue de programmes et d'installateurs, de variantes de programmes et de profils de configuration, des pages distinctes pour les pilotes et les paramètres des programmes, vos propres pages HTML.
- **Lancement** – EXE, MSI, BAT, CMD et PowerShell. Les programmes et outils portables ont un bouton **▶ Lancer** et démarrent sur place depuis la clé USB (comme un double-clic dans l'Explorateur, UAC compris) : le fichier défini dans le catalogue, sinon l'unique `.exe` correspondant du dossier (désinstalleurs, installateurs et utilitaires sont ignorés, la version 64 bits est préférée). S'il y a plusieurs candidats, CoreDeck demande lequel lancer ; s'il n'y a qu'une archive ou un installateur, il ouvre le dossier.
- **Résultats des installateurs** – les codes de sortie 1602 (annulé dans l'installateur) et 1223 (UAC refusé) affichent un bref « Installation annulée » ; les codes MSI courants (1603, 1618, 1619, 1638, 1639 et d'autres) reçoivent des messages lisibles ; 3010/1641 sont un succès accompagné d'une note de redémarrage.
- **Boîtes de dialogue intégrées** – aucune fenêtre `alert/confirm/prompt` du navigateur ; les boîtes de dialogue et les notifications fonctionnent au clavier (Entrée, Échap, Tab, restauration du focus). Lancer un installateur, un script ou un programme ne demande aucune confirmation ; la confirmation reste là où quelque chose est supprimé ou le système modifié (suppressions, réglages, installation de nombreux programmes à la fois).
- **Base de connaissances** avec vérification des liens, et un site facultatif de base de connaissances propre, servi depuis `data\KnowledgeBase\`.
- **Export et import** des données.
- **Configurateur Windows** – tweaks et debloat, avec aperçu du script et approbation (voir [Sécurité](#sécurité)).
- **Journal d'audit** des opérations d'administration dans `data/logs/audit.jsonl`.
- **Mises à jour** depuis les publications GitHub avec canaux stable et de test, vérification des sommes de contrôle et retour arrière automatique (voir [Mises à jour](#mises-à-jour)).
- **Interface** – taille et mode de la fenêtre mémorisés, thème sombre et clair, 30 langues d'interface.

## Démarrage rapide

Téléchargez `CoreDeck-Portable.zip` depuis **Releases** et décompressez-le où vous voulez. Le programme est portable et ne nécessite aucune installation : lancez `CoreDeck.exe`. Au premier démarrage, le dossier `data\` est créé à côté de l'exe (voir [Dossier de données](#dossier-de-données)).

Pour mettre à jour manuellement au lieu d'utiliser le système de mise à jour intégré : fermez CoreDeck et remplacez `CoreDeck.exe` (et `data\tools\disk-helper.exe`). Le dossier `data\` reste tel quel.

## Organisation portable

L'organisation prévue sur une clé USB est `<drive>\_Launcher\CoreDeck.exe`, avec la bibliothèque de programmes dans des dossiers à côté de `_Launcher`. Les chemins du catalogue sont enregistrés relativement à l'exe (`../…`), et les chemins par défaut d'import/export, des programmes et des pilotes sont déduits du lecteur depuis lequel CoreDeck s'exécute ; le projet n'a donc besoin d'aucun chemin fixe et peut être déplacé vers un autre lecteur ou un autre ordinateur.

Si la clé USB reçoit une autre lettre, un chemin enregistré avec l'ancienne lettre (p. ex. `J:\…` après que le lecteur est devenu `L:`) est recherché sur le lecteur depuis lequel CoreDeck s'exécute. Au démarrage, ces chemins dans `programs.json`, `drivers.json` et `settings.json` sont réécrits une seule fois sous forme relative, avec conservation d'une copie `*.drive-<date>.bak`.

## Dossier de données

Au premier démarrage, les éléments suivants sont créés à côté de l'exe :

```text
data/
├── programs.json
├── drivers.json
├── settings.json
├── knowledge.json
├── tweaks.json
├── KnowledgeBase/     (facultatif : votre propre site de base de connaissances, /KnowledgeBase/)
├── icons/
├── user_pages/
├── webview/           (profil WebView2 : thème, langue, localStorage)
├── logs/
└── state/
```

**Données privées et un seul exe.** Toutes les données utilisateur sont lues depuis `data\` à côté de `CoreDeck.exe`. Les fichiers intégrés à l'exe (`cmd/launcher/web/data/*.json`) ne sont que des modèles vides, utilisés lorsque le fichier dans `data\` n'existe pas. Ainsi, le `CoreDeck.exe` public des Releases et une installation privée sont le même fichier : le catalogue privé, la base de connaissances (`knowledge.json`) et le site `data\KnowledgeBase\` ne se trouvent que dans `data\`. Un fichier corrompu dans `data\` est signalé comme une erreur et n'est pas remplacé par le modèle vide (l'enregistrement suivant n'écrasera pas les données). Des tests garantissent qu'aucune donnée privée ne se retrouve dans les modèles intégrés ou dans `cmd/launcher/web/`.

## Mises à jour

Paramètres → **Mises à jour**.

- **Canaux :** Stable (par défaut) et Test (préversions).
- **Recherche :** recherche facultative au démarrage (au plus une fois par jour) et un bouton **Rechercher des mises à jour** pour une recherche manuelle. Lorsqu'une version plus récente existe, une fenêtre de l'application affiche son journal des modifications avec **Mettre à jour maintenant**, **Plus tard** et **Ignorer cette version**.
- **Téléchargement et vérification :** `CoreDeck-Portable.zip` est téléchargé depuis la publication GitHub de `snakex21/CoreDeck` ; sa taille et son SHA-256 sont comparés à `SHA256SUMS.txt`. En cas de non-concordance, la mise à jour est interrompue et rien n'est modifié.
- **Installation :** seuls `CoreDeck.exe` et `data\tools\disk-helper.exe` sont décompressés, dans le dossier de préparation `.coredeck-update\` à côté de l'exe. Les binaires précédents sont sauvegardés dans `.coredeck-update\backup\<version>\`, les fichiers sont échangés en renommant les anciens en `.old`, puis CoreDeck redémarre.
- **Contrôle de démarrage :** si la nouvelle version ne signale pas un démarrage correct dans les 60 secondes (ou se ferme), la version précédente est automatiquement restaurée et relancée. **Restaurer la version précédente** est également disponible dans les Paramètres.
- Fonctionne dans l'organisation portable sur clé USB (`<drive>\_Launcher\CoreDeck.exe`), ne nécessite pas de droits d'administrateur pour la mise à jour elle-même et ne modifie jamais rien d'autre dans `data\`.
- Aucune télémétrie : seuls l'API GitHub et le téléchargement de la publication sont contactés.

## Langue de l'interface

L'interface est disponible en 30 langues : English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 et 한국어.

La langue d'affichage de Windows est détectée automatiquement. Le sélecteur de langue se trouve dans la barre supérieure et propose une option **Automatique**. Les textes absents d'une traduction s'affichent en anglais. Les notes de version et le [CHANGELOG](../CHANGELOG.md) sont rédigés en polonais et en anglais.

## Compilation

Prérequis :

- Windows,
- Go 1.26 ou plus récent (`go.mod` fixe la chaîne d'outils `go1.26.11`, téléchargée automatiquement par `go`),
- un compilateur C (MinGW gcc, requis par WebView),
- facultativement `go-winres` (icône et informations de version dans l'exe),
- Node.js,
- Python.

```bat
build.bat
```

Cette commande vérifie la syntaxe JavaScript, exécute les tests Go et produit localement `CoreDeck.exe` et `data`. Les builds de publication sont allégés (`-trimpath -ldflags "-s -w"`), et `build.bat` inscrit la version, le nom du produit et le copyright dans les ressources Windows. Variante avec console de diagnostic :

```bat
build.bat debug
```

Les outils de développement de WebView2 peuvent être activés sans build séparé :

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat` construit l'archive de publication `dist\CoreDeck-Portable.zip`.

## Tests

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Tests dans le navigateur : exécutez `python -m http.server 8765` dans le dossier du dépôt et ouvrez `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (boîtes de dialogue et notifications, « PASS n/n » dans le titre) ou `test/web/launcher-smoke.html` (le lanceur avec une API simulée).

La CI (GitHub Actions) exécute la vérification de la syntaxe JS, les tests Python, `go vet`, `go test -race`, `govulncheck` et le build.

## Structure du projet

```text
cmd/launcher/         application et interface intégrée
cmd/disk-helper/      utilitaire de recherche du lecteur
internal/             logique du programme
scripts/              build et empaquetage
test/                 tests
```

Les données vides par défaut intégrées à l'exe se trouvent dans `cmd/launcher/web/data/`. Les données créées pendant l'utilisation du programme sont locales et ignorées par Git.

## Confidentialité

Le programme fonctionne localement. Le dépôt ne contient ni catalogue utilisateur, ni historique, ni pages privées, ni modèles d'IA, ni bibliothèque d'installateurs. Le seul accès réseau de CoreDeck lui-même est la recherche et le téléchargement des mises à jour, qui ne contactent que l'API GitHub et les publications GitHub officielles du projet.

## Sécurité

- Le serveur local de l'interface n'écoute que sur `127.0.0.1`, ne répond qu'à l'en-tête `Host` `localhost:<port>` / `127.0.0.1:<port>` et n'envoie aucun en-tête CORS (tout est same-origin).
- Les tweaks proviennent de `data\tweaks.json` (import, modification) et s'exécutent avec des droits d'administrateur. Avant la première exécution d'un tweak, et après chaque modification de sa commande, le configurateur affiche le script exact et demande une confirmation. Les approbations (SHA-256 du type et de la commande) sont stockées dans `data\state\tweak_approvals.json` ; le programme exécute les commandes du catalogue, et non ce que la page envoie.
- Les mises à jour ne sont installées qu'une fois que la taille et le SHA-256 correspondent à `SHA256SUMS.txt` (voir [Mises à jour](#mises-à-jour)).

## Profil WebView2

Depuis la version 2.3.1, le profil WebView2 (thème, langue, `localStorage`) se trouve dans `data\webview\` à côté de l'exe. Une copie locale et modifiée de `webview_go` est utilisée (`third_party/webview_go`, décrite dans `PATCHED.md`). Au premier démarrage, un profil existant dans `%APPDATA%\CoreDeck.exe` est copié (sans le cache) ; l'original reste en place et peut être supprimé manuellement plus tard.

## Licence

CoreDeck est publié sous licence MIT, voir [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
