# i18n glossary

Canonical UI terms for the 30 interface languages in
`cmd/launcher/web/js/i18n/translations.js` (and `tweak-translations.js`).
When a string mentions one of these concepts, use the term below. Inflect it
as the grammar needs (case, plural, compounds), but do not swap in a synonym.

`scripts/test_i18n_glossary.py` checks the most important parts of this file
(tweak, driver, profile and update terms; the language decisions below).

## Decisions

These were decided by the project owner and apply everywhere:

- **Portuguese `pt` is Brazilian Portuguese (pt-BR).** The language picker
  shows "Português (Brasil)". European forms are replaced: ficheiro → arquivo,
  guardar → salvar, definições → configurações, ecrã → tela,
  "A carregar…" → "Carregando…" (gerund, not "a + infinitive"),
  "O Teu / A Tua" → "Seu / Sua", utilizador → usuário,
  transferir → baixar, secção → seção, registo → registro,
  eliminar → excluir.
- **Turkish `tr` uses correct Turkish letters** (ç ğ ı İ ö ş ü) everywhere.
  ASCII spellings such as "Klasoru", "Suruculer", "sec", "ac" are errors
  ("Klasörü", "Sürücüler", "seç", "aç").
- **Dutch `nl` addresses the user informally** with je/jouw (jij when
  stressed). Never "u" or "uw".
- **Chinese, Japanese, Korean use one term per concept:**
  - Korean tweak = **트윅** (the usual Korean transliteration of "tweak" in
    Windows-tuning communities). 튜크 is not a Korean word and is replaced.
  - Chinese profile = **配置文件** (used 16× vs 预设 6×). 预设 means "preset" and
    is replaced. Chinese tweak = **优化** (优化项 where a countable item reads
    better); 调整 is replaced.
  - Japanese strings contain Japanese only — no Chinese sentences (e.g. the
    old 此調整是選定預設的一部分) and no half-translated words (ランcherー).
    Japanese tweak = **ツイーク** (not トゥイーク).
- Where a language mixed two words for the same concept, the term used most
  often in `translations.js` wins (e.g. Spanish tweak = "ajuste" 16× vs
  "tweak" 7×; Czech "úprava"; Swedish "justering").

## Terms

Native language names are the ones shown in the language picker.

### Table A

| Code | Language | launcher (the app) | tweak (system tweak) | driver | portable (program) | installer | update (noun) | update channel |
|---|---|---|---|---|---|---|---|---|
| pl | Polski | launcher | tweak | sterownik | program portable | instalator | aktualizacja | kanał aktualizacji |
| en | English | launcher | tweak | driver | portable program | installer | update | update channel |
| bg | Български | стартер | оптимизация | драйвер | преносима програма | инсталатор | актуализация | канал за актуализации |
| cs | Čeština | launcher | úprava | ovladač | přenosný program | instalátor | aktualizace | kanál aktualizací |
| da | Dansk | starter | justering | driver | portabelt program | installationsprogram | opdatering | opdateringskanal |
| de | Deutsch | Launcher | Tweak | Treiber | portables Programm | Installationsprogramm | Update | Update-Kanal |
| el | Ελληνικά | εκκινητής | βελτιστοποίηση | πρόγραμμα οδήγησης | φορητό πρόγραμμα | πρόγραμμα εγκατάστασης | ενημέρωση | κανάλι ενημερώσεων |
| es | Español | launcher | ajuste | controlador | programa portable | instalador | actualización | canal de actualización |
| et | Eesti | käivitaja | kohandus | draiver | kaasaskantav programm | installer | värskendus | värskenduskanal |
| fi | Suomi | käynnistin | säätö | ajuri | portable-ohjelma | asennusohjelma | päivitys | päivityskanava |
| fr | Français | lanceur | tweak | pilote | programme portable | installateur | mise à jour | canal de mise à jour |
| hr | Hrvatski | pokretač | prilagodba | upravljački program | prijenosni program | instalacijski program | ažuriranje | kanal ažuriranja |
| hu | Magyar | indító | finomhangolás | illesztőprogram | hordozható program | telepítő | frissítés | frissítési csatorna |
| it | Italiano | launcher | tweak (invariable: i tweak) | driver | programma portable | installer | aggiornamento | canale di aggiornamento |
| lt | Lietuvių | paleidiklis | patobulinimas | tvarkyklė | nešiojamoji programa | diegimo programa | naujinimas | naujinimų kanalas |
| lv | Latviešu | palaidējs | pielāgojums | draiveris | portatīva programma | instalētājs | atjauninājums | atjauninājumu kanāls |
| nb | Norsk bokmål | startprogram | justering | driver | portabelt program | installasjonsprogram | oppdatering | oppdateringskanal |
| nl | Nederlands | launcher | aanpassing | stuurprogramma | portable programma | installatieprogramma | update (verb: bijwerken) | updatekanaal |
| pt | Português (Brasil) | launcher | tweak | driver | programa portátil | instalador | atualização | canal de atualização |
| ro | Română | lansator | ajustare | driver | program portabil | program de instalare | actualizare | canal de actualizare |
| ru | Русский | лаунчер | твик | драйвер | портативная программа | установщик | обновление | канал обновлений |
| sk | Slovenčina | spúšťač | úprava | ovládač | prenosný program | inštalátor | aktualizácia | kanál aktualizácií |
| sl | Slovenščina | zaganjalnik | prilagoditev | gonilnik | prenosni program | namestitveni program | posodobitev | kanal posodobitev |
| sr-Latn | Srpski (latinica) | pokretač | (sistemska) izmena | drajver | portable program | instaler | ažuriranje | kanal ažuriranja |
| sv | Svenska | launcher | justering | drivrutin | portabelt program | installationsprogram | uppdatering | uppdateringskanal |
| tr | Türkçe | launcher | ince ayar | sürücü | taşınabilir program | yükleyici | güncelleme | güncelleme kanalı |
| uk | Українська | лаунчер | твік | драйвер | портативна програма | інсталятор | оновлення | канал оновлень |
| zh | 中文（简体） | 启动器 | 优化（优化项） | 驱动 | 便携版 | 安装程序 | 更新 | 更新通道 |
| ja | 日本語 | ランチャー | ツイーク | ドライバー | ポータブル | インストーラー | アップデート | アップデートチャネル |
| ko | 한국어 | 런처 | 트윅 | 드라이버 | 포터블 | 설치 프로그램 | 업데이트 | 업데이트 채널 |

### Table B

| Code | Language | Stable / Test channel | catalogue | profile | knowledge base | debloat | script | Settings |
|---|---|---|---|---|---|---|---|---|
| pl | Polski | Stabilny / Testowy | katalog | profil | Baza Wiedzy | debloat | skrypt | Ustawienia |
| en | English | Stable / Test | catalog | profile | Knowledge Base | debloat | script | Settings |
| bg | Български | Стабилен / Тестов | каталог | профил | База знания | debloat | скрипт | Настройки |
| cs | Čeština | Stabilní / Testovací | katalog | profil | Báze znalostí | debloat | skript | Nastavení |
| da | Dansk | Stabil / Test | katalog | profil | Vidensbase | debloat | script | Indstillinger |
| de | Deutsch | Stabil / Test | Katalog | Profil | Wissensdatenbank | Debloat | Skript | Einstellungen |
| el | Ελληνικά | Σταθερό / Δοκιμαστικό | κατάλογος | προφίλ | Βάση γνώσεων | debloat | script | Ρυθμίσεις |
| es | Español | Estable / Prueba | catálogo | perfil | Base de conocimiento | debloat | script | Configuración |
| et | Eesti | Stabiilne / Test | kataloog | profiil | Teadmusbaas | debloat | skript | Seaded |
| fi | Suomi | Vakaa / Testi | luettelo | profiili | Tietopankki | debloat | skripti | Asetukset |
| fr | Français | Stable / Test | catalogue | profil | Base de connaissances | debloat | script | Paramètres |
| hr | Hrvatski | Stabilni / Testni | katalog | profil | Baza znanja | debloat | skripta | Postavke |
| hu | Magyar | Stabil / Teszt | katalógus | profil | Tudásbázis | debloat | szkript | Beállítások |
| it | Italiano | Stabile / Test | catalogo | profilo | Knowledge Base | debloat | script | Impostazioni |
| lt | Lietuvių | Stabilus / Bandomasis | katalogas | profilis | Žinių bazė | debloat | scenarijus | Nustatymai |
| lv | Latviešu | Stabilais / Testa | katalogs | profils | Zināšanu bāze | debloat | skripts | Iestatījumi |
| nb | Norsk bokmål | Stabil / Test | katalog | profil | Kunnskapsbase | debloat | skript | Innstillinger |
| nl | Nederlands | Stabiel / Test | catalogus | profiel | Kennisbank | debloat | script | Instellingen |
| pt | Português (Brasil) | Estável / Teste | catálogo | perfil | Base de Conhecimento | debloat | script | Configurações |
| ro | Română | Stabil / Test | catalog | profil | Bază de cunoștințe | debloat | script | Setări |
| ru | Русский | Стабильный / Тестовый | каталог | профиль | База знаний | деблоат | скрипт | Настройки |
| sk | Slovenčina | Stabilný / Testovací | katalóg | profil | Databáza znalostí | debloat | skript | Nastavenia |
| sl | Slovenščina | Stabilni / Testni | katalog | profil | Baza znanja | debloat | skripta | Nastavitve |
| sr-Latn | Srpski (latinica) | Stabilni / Test | katalog | profil | Baza znanja | debloat | skripta | Podešavanja |
| sv | Svenska | Stabil / Test | katalog | profil | Kunskapsbas | debloat | skript | Inställningar |
| tr | Türkçe | Kararlı / Test | katalog | profil | Bilgi Bankası | debloat | betik | Ayarlar |
| uk | Українська | Стабільний / Тестовий | каталог | профіль | База знань | debloat | скрипт | Налаштування |
| zh | 中文（简体） | 稳定版 / 测试版 | 目录 | 配置文件 | 知识库 | 精简 | 脚本 | 设置 |
| ja | 日本語 | 安定版 / テスト版 | カタログ | プロファイル | ナレッジベース | デブロート | スクリプト | 設定 |
| ko | 한국어 | 안정판 / 테스트 | 카탈로그 | 프로필 | 지식 베이스 | 디블로트 | 스크립트 | 설정 |

Notes:

- "Settings" is the app's settings (`upd_skipped` "…later in Settings"). In
  Spanish it is "Configuración", never "Ajustes", because "ajuste" is the
  tweak term.
- "Debloat" stays an English loanword in Latin-script languages (lower case in
  running text, capitalised as a tab name). Russian already writes деблоат.
- The Test channel name may carry a short explanation, e.g.
  "Testowy (wersje przedpremierowe)".

## Register

How the UI addresses the user. Use the same form in buttons, hints and
dialogs.

| Code | Register |
|---|---|
| pl | informal "ty" (masz, możesz; buttons in the imperative) |
| en | neutral "you" |
| bg | formal "Вие" (Изберете, можете) |
| cs | formal "vy" (Vyberte, můžete) |
| da | informal "du" |
| de | formal "Sie" |
| el | formal plural (Επιλέξτε, μπορείτε) |
| es | informal "tú" (Selecciona, puedes) |
| et | informal "sina" (Vali, saad) |
| fi | informal "sinä" (Valitse, voit) |
| fr | formal "vous" |
| hr | formal "vi" (Odaberite) |
| hu | formal "Ön" (Válasszon, telepítheti) |
| it | informal "tu" (Seleziona, puoi) |
| lt | formal "jūs" (Pažymėkite) |
| lv | formal "jūs" (Atlasiet) or impersonal |
| nb | informal "du" |
| nl | informal "je/jouw" (decision above) |
| pt | "você" (pt-BR): Selecione, você pode |
| ro | formal "dumneavoastră" (Selectați); short button labels may use the 2nd-person singular imperative (Editează) |
| ru | formal "вы" (Выберите) |
| sk | formal "vy" (Vyberte) |
| sl | formal "vi" (Izberite) |
| sr-Latn | formal "vi" (Izaberite) |
| sv | informal "du" |
| tr | formal "siz" (Seçin, yükleyebilirsiniz) |
| uk | formal "ви" (Виберіть) |
| zh | polite 您 |
| ja | polite です/ます; short labels in noun form |
| ko | polite 합니다/하세요 |

## Rules

- **Placeholders** such as `{version}`, `{current}`, `{count}`, `{n}`,
  `{error}`, `{file}`, `{percent}` are copied exactly, never translated or
  removed. Their order may change. Turkish writes the percent sign first
  (`%{percent}`).
- **Product and file names are never translated:** CoreDeck, Windows,
  PowerShell, Ninite, NVIDIA, AMD, Radeon, GeForce, FreeSync, Microsoft,
  Firefox, `data\`, `tweaks.json`, `programs.json`, `settings.json`,
  `knowledge.json`, `CoreDeck-Portable.zip`, `SHA256SUMS.txt`, SHA-256.
- **Plurals:** a string that contains a number uses a plural key family
  `base_zero/_one/_two/_few/_many/_other` (CLDR categories for that language)
  and is shown with `tn(base, n)`. Never build a plural by gluing a number to a
  fixed word. `_other` is always required; add the other categories that
  `Intl.PluralRules` returns for the language (e.g. pl: one/few/many/other,
  ru/uk: one/few/many/other, cs/sk: one/few/many/other, lt: one/few/many/other,
  lv: zero/one/other, sl: one/two/few/other, ja/ko/zh: other only).
- **Every language has every key** of `pl`/`en`
  (`cmd/launcher/i18n_test.go`).
- **Script:** each language is written in its own script with its own
  letters — no ASCII-only Turkish, German without umlauts, French without
  accents; no Chinese text in `ja` or `ko`; no kana in `zh`.
- Keep UI punctuation of the language: French puts a space before `:` `?` `!`;
  Chinese/Japanese use full-width punctuation in sentences; the ellipsis is
  `…` or `...` as already used in the key.
