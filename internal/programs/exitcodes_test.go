package programs

import (
	"strings"
	"testing"
)

func TestDescribeInstallerExitCancelled(t *testing.T) {
	// 1602 = ERROR_INSTALL_USER_EXIT (Cancel in the installer),
	// 1223 = ERROR_CANCELLED (e.g. "No" in the UAC prompt).
	for _, code := range []int{1602, 1223} {
		got := DescribeInstallerExit(code)
		if got.Outcome != OutcomeCancelled || got.Success() {
			t.Fatalf("code %d: outcome %q, success %v; want cancelled", code, got.Outcome, got.Success())
		}
		if got.Message != "Instalacja anulowana" {
			t.Fatalf("code %d: message %q", code, got.Message)
		}
	}
}

func TestDescribeInstallerExitKnownCodes(t *testing.T) {
	cases := []struct {
		code    int
		outcome string
		want    string
	}{
		{0, OutcomeOK, "zakończony"},
		{3010, OutcomeOKNote, "Uruchom ponownie komputer"},
		{1641, OutcomeOKNote, "uruchamia ponownie"},
		{1638, OutcomeOKNote, "już zainstalowana"},
		{1618, OutcomeError, "Trwa inna instalacja"},
		{1603, OutcomeError, "1603"},
		{1619, OutcomeError, "Nie można otworzyć pakietu"},
		{1639, OutcomeError, "Argumenty instalatora"},
		{740, OutcomeError, "administratora"},
	}
	for _, c := range cases {
		got := DescribeInstallerExit(c.code)
		if got.Outcome != c.outcome || !strings.Contains(got.Message, c.want) {
			t.Errorf("code %d: got (%q, %q); want outcome %q containing %q", c.code, got.Outcome, got.Message, c.outcome, c.want)
		}
	}
}

func TestDescribeInstallerExitUnknownCodeIsReadable(t *testing.T) {
	got := DescribeInstallerExit(42)
	if got.Outcome != OutcomeError || got.Message != "Instalator zakończył się błędem (kod 42)." {
		t.Fatalf("got %+v", got)
	}
	if strings.Contains(got.Message, "exit status") {
		t.Fatalf("raw Go error leaked into message: %q", got.Message)
	}
	// NTSTATUS crash codes read better in hex.
	if got := DescribeInstallerExit(int(int32(-1073741819))); !strings.Contains(got.Message, "0xC0000005") {
		t.Fatalf("crash code: %q", got.Message)
	}
}
