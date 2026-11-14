@echo off
REM ============================================================
REM CoreDeck builder - jeden plik do testowania i budowania
REM Uzycie:
REM   build.bat          - aplikacja bez okna konsoli
REM   build.bat debug    - build z konsola diagnostyczna
REM ============================================================

setlocal EnableExtensions
cd /d "%~dp0"

set "ROOT=%~dp0"
set "WEB=%ROOT%cmd\launcher\web"
set "LAUNCHER_NEW=%ROOT%CoreDeck.new.exe"
set "HELPER_NEW=%ROOT%data\tools\disk-helper.new.exe"
set "LAUNCHER_EXE=%ROOT%CoreDeck.exe"
set "LEGACY_LAUNCHER_EXE=%ROOT%launcher.exe"
set "HELPER_EXE=%ROOT%data\tools\disk-helper.exe"

where go >nul 2>nul
if errorlevel 1 (
    echo [build.bat] BLAD: Go nie jest dostepne w PATH.
    pause
    exit /b 1
)

where node >nul 2>nul
if errorlevel 1 (
    echo [build.bat] BLAD: Node.js nie jest dostepny w PATH.
    pause
    exit /b 1
)

where python >nul 2>nul
if errorlevel 1 (
    echo [build.bat] BLAD: Python nie jest dostepny w PATH.
    pause
    exit /b 1
)

if not exist "%ROOT%data\tools" mkdir "%ROOT%data\tools"
if exist "%LAUNCHER_NEW%" del /q "%LAUNCHER_NEW%"
if exist "%HELPER_NEW%" del /q "%HELPER_NEW%"

echo ==^> Testy skryptow Python
python -m unittest discover -s "%ROOT%scripts" -q
if errorlevel 1 goto :fail

echo ==^> Generowanie data\generated\disk_map.json
python "%ROOT%scripts\generate_disk_map.py" --launcher-dir "%ROOT%." --output "%ROOT%data\generated\disk_map.json"
if errorlevel 1 goto :fail

echo ==^> Przygotowanie lokalnych danych startowych
for %%F in (programs tweaks drivers knowledge settings) do (
    if not exist "%ROOT%data\%%F.json" copy /y "%WEB%\data\%%F.json" "%ROOT%data\%%F.json" >nul
)

REM Wersja programu (internal\version\version.go) trafia do zasobow EXE.
set "APP_VERSION="
for /f tokens^=2^ delims^=^" %%V in ('findstr /c:"Current =" "%ROOT%internal\version\version.go"') do set "APP_VERSION=%%V"
if not defined APP_VERSION (
    echo [build.bat] BLAD: nie mozna odczytac wersji z internal\version\version.go.
    goto :fail
)

where go-winres >nul 2>nul
if not errorlevel 1 (
    echo ==^> Odswiezanie ikony Windows
    pushd "%ROOT%cmd\launcher"
    go-winres simply --arch amd64,386 --icon "%ROOT%cmd\launcher\assets\CoreDeck.png" ^
        --product-name CoreDeck --file-description CoreDeck --original-filename CoreDeck.exe ^
        --copyright "Copyright (c) 2026 snakex21. MIT License." ^
        --product-version %APP_VERSION% --file-version %APP_VERSION%.0
    if errorlevel 1 (
        popd
        goto :fail
    )
    popd
) else (
    echo [build.bat] go-winres niedostepny - pozostawiam istniejace zasoby ikony.
)

echo ==^> Sprawdzanie skladni JavaScript
for /R "%WEB%\js" %%F in (*.js) do (
    node --check "%%F"
    if errorlevel 1 goto :fail
)

echo ==^> Testy Go
go test ./...
if errorlevel 1 goto :fail

echo ==^> Budowanie CoreDeck.exe
set "CGO_ENABLED=1"
REM Toolchain jest przypiety w go.mod (toolchain go1.26.11, poprawki bezpieczenstwa stdlib).
REM -trimpath -s -w: bez sciezek budowania i tablic symboli (exe ok. 45%% mniejszy).
if /I "%~1"=="debug" (
    go build -trimpath -o "%LAUNCHER_NEW%" .\cmd\launcher
) else (
    go build -trimpath -ldflags="-s -w -H windowsgui" -o "%LAUNCHER_NEW%" .\cmd\launcher
)
if errorlevel 1 goto :fail

echo ==^> Budowanie data\tools\disk-helper.exe
set "CGO_ENABLED=0"
go build -trimpath -ldflags="-s -w" -o "%HELPER_NEW%" .\cmd\disk-helper
if errorlevel 1 goto :fail

move /y "%LAUNCHER_NEW%" "%LAUNCHER_EXE%" >nul
if errorlevel 1 (
    echo [build.bat] BLAD: nie mozna podmienic CoreDeck.exe.
    echo             Zamknij uruchomiony CoreDeck i sproboj ponownie.
    goto :fail
)

move /y "%HELPER_NEW%" "%HELPER_EXE%" >nul
if errorlevel 1 goto :fail

REM Stara nazwa programu potrafila zostawic debugowy build z konsola.
REM Po udanym buildzie w katalogu glownym ma zostac tylko CoreDeck.exe.
if exist "%LEGACY_LAUNCHER_EXE%" del /q "%LEGACY_LAUNCHER_EXE%"

echo.
echo BUILD OK
echo   %LAUNCHER_EXE%
echo   %HELPER_EXE%
endlocal
exit /b 0

:fail
echo.
echo [build.bat] BLAD: build nie powiodl sie.
if exist "%LAUNCHER_NEW%" del /q "%LAUNCHER_NEW%"
if exist "%HELPER_NEW%" del /q "%HELPER_NEW%"
pause
endlocal
exit /b 1
