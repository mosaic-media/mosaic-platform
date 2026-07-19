// Package sdkboundary holds the standing enforcement of ADR 0016: the
// published contract surface must be importable and self-contained for a
// Module built outside the Platform's module.
package sdkboundary_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestPublishedSurfaceCompilesFromAnExternalModule builds test/sdkprobe, a
// separate Go module that imports only contracts/platform/v1. Go forbids an
// external module from importing anything under internal/, so if any public
// signature in v1 referenced an internal type — a leak of Platform plumbing
// into the SDK — v1 would import internal/ and this build would fail with
// "use of internal package". A green build is the proof the surface is clean.
//
// This replaces the manual probe that first found the internal/ barrier with
// a check the suite runs on every change.
func TestPublishedSurfaceCompilesFromAnExternalModule(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}

	dir, err := filepath.Abs(filepath.Join("..", "sdkprobe"))
	if err != nil {
		t.Fatalf("resolve probe dir: %v", err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("an external module failed to compile against contracts/platform/v1 "+
			"— an internal type has leaked into the published surface (ADR 0016):\n%s", out)
	}
}
