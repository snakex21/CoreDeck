**Diller:** [English](../README.md) · [Polski](README.pl.md) · [Български](README.bg.md) · [Čeština](README.cs.md) · [Dansk](README.da.md) · [Deutsch](README.de.md) · [Ελληνικά](README.el.md) · [Español](README.es.md) · [Eesti](README.et.md) · [Suomi](README.fi.md) · [Français](README.fr.md) · [Hrvatski](README.hr.md) · [Magyar](README.hu.md) · [Italiano](README.it.md) · [Lietuvių](README.lt.md) · [Latviešu](README.lv.md) · [Norsk bokmål](README.nb.md) · [Nederlands](README.nl.md) · [Português (Brasil)](README.pt-BR.md) · [Română](README.ro.md) · [Русский](README.ru.md) · [Slovenčina](README.sk.md) · [Slovenščina](README.sl.md) · [Srpski](README.sr-Latn.md) · [Svenska](README.sv.md) · [Türkçe](README.tr.md) · [Українська](README.uk.md)

# CoreDeck

[![CI](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/CoreDeck/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#derleme)

Uygulamaları, sürücüleri, kurulum dosyalarını, sistem ince ayarlarını, kurulum profillerini ve yerel belgeleri düzenlemek için taşınabilir bir Windows araç seti. Sürüm **2.4.0** · değişiklikler: [CHANGELOG.md](../CHANGELOG.md).

Depo temiz bir başlangıç noktasıdır. Program, kurulum dosyası, özel bağlantı veya yazara ait veri içermez. İlk açılışta uygulama `data` klasöründe boş dosyalar oluşturur; içerik **Yönetim** ekranında eklenir.

## İçindekiler

- [Özellikler](#özellikler)
- [Hızlı başlangıç](#hızlı-başlangıç)
- [Taşınabilir düzen](#taşınabilir-düzen)
- [Veri klasörü](#veri-klasörü)
- [Güncellemeler](#güncellemeler)
- [Arayüz dili](#arayüz-dili)
- [Derleme](#derleme)
- [Testler](#testler)
- [Proje yapısı](#proje-yapısı)
- [Gizlilik](#gizlilik)
- [Güvenlik](#güvenlik)
- [WebView2 profili](#webview2-profili)
- [Lisans](#lisans)

## Özellikler

- **Katalog** — programlardan ve kurulum dosyalarından oluşan kendi kataloğunuz, program varyantları ve kurulum profilleri, sürücüler ve program ayarları için ayrı sayfalar, kendi HTML sayfalarınız.
- **Çalıştırma** — EXE, MSI, BAT, CMD ve PowerShell. Taşınabilir programlar ve araçlarda **▶ Çalıştır** düğmesi bulunur ve doğrudan USB bellekten, yerinde başlatılırlar (Gezgin'de çift tıklama gibi, UAC dahil): katalogda ayarlanan dosya, yoksa klasördeki tek uygun `.exe` (kaldırıcılar, kurulum dosyaları ve yardımcı araçlar atlanır, 64 bit tercih edilir). Birden fazla aday varsa CoreDeck hangisinin çalıştırılacağını sorar; yalnızca bir arşiv veya kurulum dosyası varsa klasörü açar.
- **Kurulum sonuçları** — 1602 (kurulum dosyasında iptal edildi) ve 1223 (UAC reddedildi) çıkış kodları kısa bir "Kurulum iptal edildi" mesajı gösterir; yaygın MSI kodları (1603, 1618, 1619, 1638, 1639 ve diğerleri) anlaşılır mesajlar alır; 3010/1641 yeniden başlatma notuyla başarılı sayılır.
- **Uygulama içi iletişim kutuları** — tarayıcının `alert/confirm/prompt` kutuları yoktur; iletişim kutuları ve bildirimler klavyeyle çalışır (Enter, Esc, Tab, odağın geri yüklenmesi). Bir kurulum dosyasını, betiği veya programı çalıştırmak onay gerektirmez; onay, bir şeyin silindiği veya sistemin değiştirildiği yerlerde kalır (silme işlemleri, ince ayarlar, birçok programı aynı anda kurma).
- Bağlantı denetimli **KnowledgeBase** ve isteğe bağlı olarak `data\KnowledgeBase\` klasöründen sunulan kendi KnowledgeBase siteniz.
- Verilerin **dışa ve içe aktarılması**.
- **Windows yapılandırıcısı** — betik önizlemesi ve onayıyla ince ayarlar ve debloat (bkz. [Güvenlik](#güvenlik)).
- Yönetimsel işlemlerin `data/logs/audit.jsonl` dosyasındaki **denetim günlüğü**.
- Kararlı ve test kanallı, sağlama toplamı doğrulamalı ve otomatik geri almalı, GitHub sürümlerinden **güncellemeler** (bkz. [Güncellemeler](#güncellemeler)).
- **Arayüz** — pencere boyutu ve modu hatırlanır, koyu ve açık tema, 30 arayüz dili.

## Hızlı başlangıç

**Releases** sayfasından `CoreDeck-Portable.zip` dosyasını indirin ve herhangi bir yere çıkarın. Program taşınabilirdir ve kurulum gerektirmez: `CoreDeck.exe` dosyasını çalıştırın. İlk açılışta exe'nin yanında `data\` klasörü oluşturulur (bkz. [Veri klasörü](#veri-klasörü)).

Yerleşik güncelleyici yerine elle güncellemek için: CoreDeck'i kapatın ve `CoreDeck.exe` (ve `data\tools\disk-helper.exe`) dosyasını değiştirin. `data\` klasörüne dokunulmaz.

## Taşınabilir düzen

USB bellekte önerilen düzen `<drive>\_Launcher\CoreDeck.exe` şeklindedir; program kitaplığı `_Launcher` klasörünün yanındaki klasörlerde bulunur. Katalog yolları exe'ye göre göreli olarak saklanır (`../…`); varsayılan içe/dışa aktarma, program ve sürücü yolları ise CoreDeck'in çalıştığı sürücüden türetilir. Böylece projenin sabit bir yola ihtiyacı olmaz ve başka bir sürücüye veya bilgisayara taşınabilir.

USB bellek farklı bir harf alırsa, eski harfle kaydedilmiş bir yol (ör. sürücü `L:` olduktan sonra `J:\…`) CoreDeck'in çalıştığı sürücüde aranır. Açılışta `programs.json`, `drivers.json` ve `settings.json` içindeki bu tür yollar bir kez göreli biçime yeniden yazılır ve bir `*.drive-<date>.bak` kopyası saklanır.

## Veri klasörü

İlk açılışta exe'nin yanında şunlar oluşturulur:

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

**Özel veriler ve tek bir exe.** Tüm kullanıcı verileri `CoreDeck.exe` yanındaki `data\` klasöründen okunur. Exe'ye gömülü dosyalar (`cmd/launcher/web/data/*.json`) yalnızca boş şablonlardır ve `data\` içindeki dosya mevcut olmadığında kullanılır. Bu sayede Releases'teki herkese açık `CoreDeck.exe` ile özel bir kurulum aynı dosyadır: özel katalog, KnowledgeBase (`knowledge.json`) ve `data\KnowledgeBase\` sitesi yalnızca `data\` içinde bulunur. `data\` içindeki bozuk bir dosya hata olarak bildirilir ve boş şablonla değiştirilmez (bir sonraki kayıt verilerin üzerine yazmaz). Testler, hiçbir özel verinin gömülü şablonlara veya `cmd/launcher/web/` klasörüne girmediğinden emin olur.

## Güncellemeler

Ayarlar → **Güncellemeler**.

- **Kanallar:** Kararlı (varsayılan) ve Test (ön sürümler).
- **Denetim:** açılışta isteğe bağlı denetim (günde en fazla bir kez) ve elle kullanılan **Güncellemeleri denetle** düğmesi. Daha yeni bir sürüm varsa, uygulama içi bir pencere değişiklik listesini **Şimdi güncelle**, **Daha sonra** ve **Bu sürümü atla** seçenekleriyle gösterir.
- **İndirme ve doğrulama:** `CoreDeck-Portable.zip`, `snakex21/CoreDeck` deposunun GitHub sürümünden indirilir; boyutu ve SHA-256 değeri `SHA256SUMS.txt` ile karşılaştırılır. Uyuşmazlık olursa güncelleme iptal edilir ve hiçbir şey değişmez.
- **Kurulum:** yalnızca `CoreDeck.exe` ve `data\tools\disk-helper.exe`, exe'nin yanındaki `.coredeck-update\` hazırlık klasörüne çıkarılır. Önceki ikili dosyaların yedeği `.coredeck-update\backup\<version>\` içine alınır, eski dosyalar `.old` olarak yeniden adlandırılarak dosyalar değiştirilir ve CoreDeck yeniden başlar.
- **Sağlık denetimi:** yeni sürüm 60 saniye içinde düzgün başladığını bildirmezse (veya kapanırsa), önceki sürüm otomatik olarak geri yüklenir ve yeniden başlatılır. **Önceki sürümü geri yükle** seçeneği Ayarlar'da da bulunur.
- Taşınabilir USB düzeninde (`<drive>\_Launcher\CoreDeck.exe`) çalışır, güncellemenin kendisi için yönetici hakları gerektirmez ve `data\` içindeki başka hiçbir şeyi değiştirmez.
- Telemetri yok: yalnızca GitHub API'sine ve sürüm indirmesine bağlanılır.

## Arayüz dili

Arayüz 30 dilde kullanılabilir: English, Polski, Български, Čeština, Dansk, Deutsch, Ελληνικά, Español, Eesti, Suomi, Français, Hrvatski, Magyar, Italiano, Lietuvių, Latviešu, Norsk bokmål, Nederlands, Português (Brasil), Română, Русский, Slovenčina, Slovenščina, Srpski (latinica), Svenska, Türkçe, Українська, 中文, 日本語 ve 한국어.

Windows görüntüleme dili otomatik olarak algılanır. Dil seçici üst çubuktadır ve **Otomatik** seçeneğine sahiptir. Bir çeviride eksik olan metin İngilizce gösterilir. Sürüm notları ve [CHANGELOG](../CHANGELOG.md) Lehçe ve İngilizcedir.

## Derleme

Gereksinimler:

- Windows,
- Go 1.26 veya daha yenisi (`go.mod`, `go` tarafından otomatik indirilen `go1.26.11` araç zincirini sabitler),
- bir C derleyicisi (MinGW gcc, WebView için gereklidir),
- isteğe bağlı olarak `go-winres` (exe'deki simge ve sürüm bilgileri),
- Node.js,
- Python.

```bat
build.bat
```

Bu komut JavaScript sözdizimini denetler, Go testlerini çalıştırır ve yerel olarak `CoreDeck.exe` ile `data` üretir. Yayın derlemeleri sembollerden arındırılır (`-trimpath -ldflags "-s -w"`) ve `build.bat`, sürümü, ürün adını ve telif hakkını Windows kaynaklarına yazar. Tanılama konsollu varyant:

```bat
build.bat debug
```

WebView2 geliştirici araçları ayrı bir derleme olmadan etkinleştirilebilir:

```bat
set COREDECK_WEBVIEW_DEBUG=1
CoreDeck.exe
```

`scripts\package.bat`, `dist\CoreDeck-Portable.zip` yayın arşivini oluşturur.

## Testler

```bat
go vet ./...
go test ./...
python -m unittest discover -s scripts
```

Tarayıcı testleri: depo klasöründe `python -m http.server 8765` komutunu çalıştırın ve `http://127.0.0.1:8765/test/web/dialogs-smoke.html` (iletişim kutuları ve bildirimler, başlıkta "PASS n/n") veya `test/web/launcher-smoke.html` (taklit API'li başlatıcı) sayfasını açın.

CI (GitHub Actions); JS sözdizimi denetimini, Python testlerini, `go vet`, `go test -race`, `govulncheck` ve derlemeyi çalıştırır.

## Proje yapısı

```text
cmd/launcher/         application and embedded interface
cmd/disk-helper/      helper for drive lookup
internal/             program logic
scripts/              build and packaging
test/                 tests
```

Exe'ye gömülü varsayılan boş veriler `cmd/launcher/web/data/` içindedir. Program kullanılırken oluşturulan veriler yereldir ve Git tarafından yok sayılır.

## Gizlilik

Program yerel olarak çalışır. Depo; kullanıcı kataloğu, geçmiş, özel sayfalar, yapay zekâ modelleri veya kurulum dosyası kitaplığı içermez. CoreDeck'in kendisinin tek ağ erişimi güncelleme denetimi ve indirmesidir; bunlar yalnızca GitHub API'sine ve projenin resmî GitHub sürümlerine bağlanır.

## Güvenlik

- Yerel arayüz sunucusu yalnızca `127.0.0.1` adresini dinler, yalnızca `localhost:<port>` / `127.0.0.1:<port>` `Host` başlığına yanıt verir ve CORS başlıkları göndermez (her şey same-origin'dir).
- İnce ayarlar `data\tweaks.json` dosyasından gelir (içe aktarma, düzenleme) ve yönetici haklarıyla çalışır. Bir ince ayar ilk kez çalışmadan önce ve komutundaki her değişiklikten sonra yapılandırıcı tam betiği gösterir ve onay ister. Onaylar (türün ve komutun SHA-256 değeri) `data\state\tweak_approvals.json` içinde saklanır; program, sayfanın gönderdiğini değil, katalogdaki komutları çalıştırır.
- Güncellemeler yalnızca boyut ve SHA-256, `SHA256SUMS.txt` ile eşleştikten sonra kurulur (bkz. [Güncellemeler](#güncellemeler)).

## WebView2 profili

2.3.1 sürümünden bu yana WebView2 profili (tema, dil, `localStorage`) exe'nin yanındaki `data\webview\` klasöründe bulunur. `webview_go` paketinin yerel, yamalı bir kopyası kullanılır (`third_party/webview_go`, `PATCHED.md` içinde açıklanmıştır). İlk açılışta `%APPDATA%\CoreDeck.exe` içindeki mevcut profil (önbellek olmadan) kopyalanır; orijinali yerinde kalır ve daha sonra elle silinebilir.

## Lisans

CoreDeck MIT lisansı altında yayımlanmıştır, bkz. [LICENSE](../LICENSE).
Copyright (c) 2026 snakex21.
