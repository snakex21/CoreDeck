package programs

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"launcher/internal/uimsg"
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
	Msg     uimsg.Msg
}

// Success reports whether the outcome counts as a finished installation.
func (e InstallerExit) Success() bool {
	return e.Outcome == OutcomeOK || e.Outcome == OutcomeOKNote
}

// DescribeInstallerExit turns an installer exit code into a short message for
// people, instead of "exit status 1602". Cancelling is not an error.
func DescribeInstallerExit(code int) InstallerExit {
	out := func(outcome string, msg uimsg.Msg) InstallerExit {
		return InstallerExit{Code: code, Outcome: outcome, Msg: msg}
	}
	switch code {
	case 0:
		return out(OutcomeOK, uimsg.New("inst_done", "Instalator zakończony"))
	case codeInstallUserExit, codeCancelled:
		return out(OutcomeCancelled, uimsg.New("install_cancelled", "Instalacja anulowana"))
	case codeRebootRequired:
		return out(OutcomeOKNote, uimsg.New("inst_code_3010", "Zainstalowano. Uruchom ponownie komputer, aby dokończyć instalację."))
	case codeRebootInitiated:
		return out(OutcomeOKNote, uimsg.New("inst_code_1641", "Zainstalowano. Instalator uruchamia ponownie komputer."))
	case codeProductVersion:
		return out(OutcomeOKNote, uimsg.New("inst_code_1638", "Ta lub inna wersja programu jest już zainstalowana. Aby zmienić wersję, najpierw odinstaluj obecną."))
	case codeInstallAlreadyRunning:
		return out(OutcomeError, uimsg.New("inst_code_1618", "Trwa inna instalacja. Poczekaj, aż się skończy, i spróbuj ponownie."))
	case codeInstallFailure:
		return out(OutcomeError, uimsg.New("inst_code_1603", "Instalacja nie powiodła się (błąd krytyczny instalatora, kod 1603). Często pomaga zamknięcie programu, jeśli jest uruchomiony, albo odinstalowanie starszej wersji."))
	case codeInstallServiceFailure:
		return out(OutcomeError, uimsg.New("inst_code_1601", "Usługa Instalator Windows jest niedostępna (kod 1601)."))
	case codeInstallSourceAbsent:
		return out(OutcomeError, uimsg.New("inst_code_1612", "Źródło instalacji jest niedostępne (kod 1612). Sprawdź, czy pendrive jest podłączony."))
	case codeInstallPackageOpen:
		return out(OutcomeError, uimsg.New("inst_code_1619", "Nie można otworzyć pakietu instalacyjnego (kod 1619). Sprawdź, czy plik istnieje i nie jest uszkodzony."))
	case codeInstallPackageInvalid:
		return out(OutcomeError, uimsg.New("inst_code_1620", "Pakiet instalacyjny jest uszkodzony albo niepoprawny (kod 1620)."))
	case codeInstallPackageRejected:
		return out(OutcomeError, uimsg.New("inst_code_1625", "Zasady systemu blokują tę instalację (kod 1625)."))
	case codeInstallPlatform:
		return out(OutcomeError, uimsg.New("inst_code_1633", "Ten pakiet nie pasuje do tej wersji albo architektury Windows (kod 1633)."))
	case codeInvalidCommandLine:
		return out(OutcomeError, uimsg.New("inst_code_1639", "Nieprawidłowe parametry instalatora (kod 1639). Sprawdź pole „Argumenty instalatora”."))
	case codeAccessDenied:
		return out(OutcomeError, uimsg.New("inst_code_5", "Brak uprawnień (kod 5). Uruchom CoreDeck jako administrator."))
	case codeElevationRequired:
		return out(OutcomeError, uimsg.New("inst_code_740", "Instalator wymaga uprawnień administratora (kod 740). Uruchom CoreDeck jako administrator."))
	}
	return out(OutcomeError, uimsg.New("inst_code_other", fmt.Sprintf("Instalator zakończył się błędem (kod %s).", formatExitCode(code)), "code", formatExitCode(code)))
}

// formatExitCode prints NTSTATUS-style codes (e.g. 0xC0000005) in hex.
func formatExitCode(code int) string {
	if code < 0 || code > 0xFFFF {
		return fmt.Sprintf("0x%08X", uint32(code))
	}
	return fmt.Sprintf("%d", code)
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
		r := newResult(name, "installer", true, installerDoneMessage(installerPath, ranFromTemp))
		r.Outcome = OutcomeOK
		return r
	}
	if exit, ok := installerExitFromError(err); ok {
		r := newResult(name, "installer", exit.Success(), exit.Msg)
		r.Cancelled = exit.Outcome == OutcomeCancelled
		r.Outcome = exit.Outcome
		r.ExitCode = exit.Code
		return r
	}
	r := newResult(name, "installer", false, uimsg.New("msg_inst_start_failed", "Błąd uruchamiania: "+err.Error(), "error", err.Error()))
	r.Outcome = OutcomeError
	return r
}
