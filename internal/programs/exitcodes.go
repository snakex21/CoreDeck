package programs

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

// Outcome values reported to the UI in InstallResult.Outcome.
const (
	OutcomeOK        = "ok"        // finished normally
	OutcomeOKNote    = "ok_note"   // finished, but the user should know something (restart, already installed)
	OutcomeCancelled = "cancelled" // the user closed the installer or refused UAC: not an error
	OutcomeError     = "error"
)

// Windows / Windows Installer exit codes the launcher translates.
// https://learn.microsoft.com/windows/win32/msi/error-codes
const (
	codeAccessDenied           = 5
	codeElevationRequired      = 740
	codeCancelled              = 1223 // ERROR_CANCELLED, e.g. "No" in the UAC prompt
	codeInstallServiceFailure  = 1601
	codeInstallUserExit        = 1602 // ERROR_INSTALL_USER_EXIT
	codeInstallFailure         = 1603
	codeInstallSourceAbsent    = 1612
	codeInstallAlreadyRunning  = 1618
	codeInstallPackageOpen     = 1619
	codeInstallPackageInvalid  = 1620
	codeInstallPackageRejected = 1625
	codeInstallPlatform        = 1633
	codeInvalidCommandLine     = 1639
	codeProductVersion         = 1638 // another version is already installed
	codeRebootInitiated        = 1641
	codeRebootRequired         = 3010
)

// InstallerExit describes how an installer's exit code should be presented.
type InstallerExit struct {
	Code    int
	Outcome string
	Message string
}

// Success reports whether the outcome counts as a finished installation.
func (e InstallerExit) Success() bool {
	return e.Outcome == OutcomeOK || e.Outcome == OutcomeOKNote
}

// DescribeInstallerExit turns an installer exit code into a short message for
// people, instead of "exit status 1602". Cancelling is not an error.
func DescribeInstallerExit(code int) InstallerExit {
	out := func(outcome, msg string) InstallerExit {
		return InstallerExit{Code: code, Outcome: outcome, Message: msg}
	}
	switch code {
	case 0:
		return out(OutcomeOK, "Instalator zakończony")
	case codeInstallUserExit, codeCancelled:
		return out(OutcomeCancelled, "Instalacja anulowana")
	case codeRebootRequired:
		return out(OutcomeOKNote, "Zainstalowano. Uruchom ponownie komputer, aby dokończyć instalację.")
	case codeRebootInitiated:
		return out(OutcomeOKNote, "Zainstalowano. Instalator uruchamia ponownie komputer.")
	case codeProductVersion:
		return out(OutcomeOKNote, "Ta lub inna wersja programu jest już zainstalowana. Aby zmienić wersję, najpierw odinstaluj obecną.")
	case codeInstallAlreadyRunning:
		return out(OutcomeError, "Trwa inna instalacja. Poczekaj, aż się skończy, i spróbuj ponownie.")
	case codeInstallFailure:
		return out(OutcomeError, "Instalacja nie powiodła się (błąd krytyczny instalatora, kod 1603). Często pomaga zamknięcie programu, jeśli jest uruchomiony, albo odinstalowanie starszej wersji.")
	case codeInstallServiceFailure:
		return out(OutcomeError, "Usługa Instalator Windows jest niedostępna (kod 1601).")
	case codeInstallSourceAbsent:
		return out(OutcomeError, "Źródło instalacji jest niedostępne (kod 1612). Sprawdź, czy pendrive jest podłączony.")
	case codeInstallPackageOpen:
		return out(OutcomeError, "Nie można otworzyć pakietu instalacyjnego (kod 1619). Sprawdź, czy plik istnieje i nie jest uszkodzony.")
	case codeInstallPackageInvalid:
		return out(OutcomeError, "Pakiet instalacyjny jest uszkodzony albo niepoprawny (kod 1620).")
	case codeInstallPackageRejected:
		return out(OutcomeError, "Zasady systemu blokują tę instalację (kod 1625).")
	case codeInstallPlatform:
		return out(OutcomeError, "Ten pakiet nie pasuje do tej wersji albo architektury Windows (kod 1633).")
	case codeInvalidCommandLine:
		return out(OutcomeError, "Nieprawidłowe parametry instalatora (kod 1639). Sprawdź pole „Argumenty instalatora”.")
	case codeAccessDenied:
		return out(OutcomeError, "Brak uprawnień (kod 5). Uruchom CoreDeck jako administrator.")
	case codeElevationRequired:
		return out(OutcomeError, "Instalator wymaga uprawnień administratora (kod 740). Uruchom CoreDeck jako administrator.")
	}
	return out(OutcomeError, fmt.Sprintf("Instalator zakończył się błędem (%s).", formatExitCode(code)))
}

// formatExitCode prints NTSTATUS-style codes (e.g. 0xC0000005) in hex.
func formatExitCode(code int) string {
	if code < 0 || code > 0xFFFF {
		return fmt.Sprintf("kod 0x%08X", uint32(code))
	}
	return fmt.Sprintf("kod %d", code)
}

// installerExitFromError classifies the error of cmd.Start/cmd.Wait.
// ok is false when err carries no exit code (e.g. file not found).
func installerExitFromError(err error) (InstallerExit, bool) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return DescribeInstallerExit(exitErr.ExitCode()), true
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch int(errno) {
		case codeElevationRequired, codeCancelled, codeAccessDenied:
			return DescribeInstallerExit(int(errno)), true
		}
	}
	return InstallerExit{}, false
}

// installerResult builds the InstallResult for a finished installer process.
func installerResult(name, installerPath string, ranFromTemp bool, err error) InstallResult {
	if err == nil {
		return InstallResult{Name: name, Success: true, Outcome: OutcomeOK, Message: installerDoneMessage(installerPath, ranFromTemp), Type: "installer"}
	}
	if exit, ok := installerExitFromError(err); ok {
		return InstallResult{
			Name:      name,
			Success:   exit.Success(),
			Cancelled: exit.Outcome == OutcomeCancelled,
			Outcome:   exit.Outcome,
			ExitCode:  exit.Code,
			Message:   exit.Message,
			Type:      "installer",
		}
	}
	return InstallResult{Name: name, Success: false, Outcome: OutcomeError, Message: "Błąd uruchamiania: " + err.Error(), Type: "installer"}
}
