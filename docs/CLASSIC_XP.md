# CoreDeck Classic: plan dla starszych Windowsów (XP – 8.1)

Status: **plan na później**, nic z tego nie jest zaimplementowane.
Analiza z listopada 2026 (wersja 2.3.2).

## Co dziś blokuje starsze systemy

| Element | Ograniczenie |
|---|---|
| Go 1.26 (`go.mod`) | Go 1.21+ działa tylko na Windows 10+. Ostatnie Go dla 7/8.1 to 1.20 (bez poprawek bezpieczeństwa), dla XP 1.10. |
| Kod Go | `log/slog` w 17 plikach (od Go 1.21), pojedyncze `slices`/`min`. |
| WebView2 | Na 7/8.1 zatrzymany na wersji 109 (bez wsparcia od 2023). Na XP nie istnieje. |
| UI | ok. 830 KB nowoczesnego JS (`?.`, `??`, async). Chromium 109 tak, przeglądarki z XP nie. |

## Opcje

1. **Win7/8.1: Go 1.20 + WebView2 109** – 1–2 dni plus testy na VM.
   Przestarzałe, niewspierane składniki i dwa toolchainy. Nie polecam.
2. **Natywny interfejs Win32 (XP → 11)** – jedna binarka, własne rysowanie
   przez GDI/GDI+ (oba są na XP): ciemny motyw, zaokrąglone karty
   z antyaliasingiem, ikony z inicjałami, hover (`TrackMouseEvent`),
   przyciski rysowane samodzielnie. Ten sam format danych (`data\*.json`),
   te same ścieżki względne i ta sama logika wyboru `.exe`
   co `internal/programs/launch.go`.
3. **Tylko XP** – wykonalne wyłącznie w wariancie 2.

## Język i toolchain dla binarki XP

- **C + MinGW-w64 i686 (msvcrt)** – tak. Podsystem `-Wl,--subsystem,windows:5.01`,
  bez API Vista+ (uwaga na winpthreads/`std::thread`). JSON: cJSON (MIT).
  W MSYS2 potrzebny pakiet `mingw-w64-i686-gcc`.
- **C++ z toolsetem v141_xp (VS2017)** – działa, ale Microsoft oznaczył go
  jako wycofywany.
- **Free Pascal / Lazarus** – pełne wsparcie Win32/XP, ale inny język.
- **Zig 0.15.2** – target `x86-windows.xp` daje PE32 z wersją podsystemu 5.01,
  ale kod startowy biblioteki standardowej importuje
  `ntdll!RtlExitUserProcess` (najpewniej Vista+, niesprawdzone na XP),
  a sama biblioteka standardowa zakłada Windows 10+. Na XP trzeba by pisać
  bez niej i z własnym punktem wejścia. Odradzam.
- **Go** – nie (ostatnie Go dla XP to 1.10).

## Ograniczenia XP

- Czcionki: brak Segoe UI. Inter da się dołączyć przez `AddFontMemResourceEx`,
  ale renderowanie GDI wygląda inaczej. Brak kolorowych emoji – ikony
  zamiast emoji.
- Ikony: PNG przez GDI+. Pliki ICO z kompresją PNG (256 px) nie ładują się
  przez `LoadImage`.
- DPI: tylko systemowe (`GetDeviceCaps(LOGPIXELSX)`), bez per-monitor.
- Ścieżki Unicode: API „W” działa.
- TLS 1.2: brak w Schannel na XP. Sprawdzanie aktualizacji wymaga
  dołączonej biblioteki TLS (np. mbedTLS) albo trzeba je wyłączyć.
- UAC nie istnieje (`runas` pokazuje „Uruchom jako”). Windows Installer 3.1.
- Brak DWM: ciemny pasek tytułu tylko przy oknie bez ramki, rysowanym
  samodzielnie.
- Katalog: większość programów (LM Studio, Python 3.14, VirtualBox 7,
  ComfyUI…) nie działa na XP. Potrzebne pole `minWindows` i filtr.

## Jak blisko obecnego wyglądu

- **Tak samo:** kolory, układ (pasek boczny, siatka kart, wyszukiwarka),
  zielone przyciski, hover, inicjały, szczegóły, otwieranie folderu,
  kopiowanie ścieżki, sterowniki. Szacunkowo 80–90% wyglądu.
- **Prościej:** brak rozmycia i animacji CSS, inne renderowanie fontów.
  `details` (HTML) jako zwykły tekst. Baza Wiedzy i strony użytkownika
  w domyślnej przeglądarce. Bez wyszukiwania AI (llama.cpp nie działa na XP).

## Pracochłonność

| Zakres | Szacunek |
|---|---|
| Classic: przeglądanie, wyszukiwanie, uruchamianie, szczegóły, sterowniki; katalog tylko do odczytu | ok. 2–3 tygodnie |
| Pełna zgodność funkcji (Zarządzanie, Konfigurator, tweaki, Baza Wiedzy, import/eksport) | 2–3 miesiące |

## Rekomendacja

Nie zastępować teraz obecnego UI. Zrobić osobny, mały
`CoreDeck-Classic.exe` (C, Win32/GDI+, MinGW-w64 i686, msvcrt) obok
`CoreDeck.exe`. Czytałby ten sam `data\` (tylko odczyt) i służył jako
wersja dla XP–8.1. Edycja katalogu zostaje w obecnym UI na Windows 10/11.
Po sprawdzeniu się Classic można zdecydować, czy ma przejąć całość.
