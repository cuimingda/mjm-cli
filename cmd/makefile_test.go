package cmd

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMakeInstallTargetWrapsGoInstall(t *testing.T) {
	t.Parallel()

	assertMakeDryRun(t, "install", "go install ./cmd/mjm")
}

func TestMakeTestTargetWrapsGoTest(t *testing.T) {
	t.Parallel()

	assertMakeDryRun(t, "test", "go test ./...")
}

func assertMakeDryRun(t *testing.T, target string, expected string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("make dry runs are only verified on Unix-like environments")
	}

	if _, err := exec.LookPath("make"); err != nil {
		t.Skipf("make is not available: %v", err)
	}

	cmd := exec.Command("make", "-n", target)
	cmd.Dir = filepath.Clean("..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run make -n %s: %v\n%s", target, err, output)
	}

	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
