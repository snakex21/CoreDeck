@echo off
REM ============================================================
REM Tworzy gotowa paczke CoreDeck-Portable.zip w katalogu dist
REM ============================================================

setlocal
cd /d "%~dp0.."

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0package.ps1"
if errorlevel 1 (
    echo.
    echo [package.bat] BLAD: tworzenie paczki nie powiodlo sie.
    pause
    exit /b 1
)

echo.
echo PACZKA GOTOWA
echo   %CD%\dist\CoreDeck-Portable.zip

endlocal
exit /b 0
